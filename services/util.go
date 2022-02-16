package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	. "kasen/cache"
	. "kasen/database"

	"kasen/config"
	"kasen/errs"
	"kasen/models"
	"kasen/modext"

	"github.com/gosimple/slug"
	"github.com/pkg/errors"
	"github.com/rs1703/logger"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"golang.org/x/time/rate"
)

const mdBaseURL = "https://api.mangadex.org"

var mdGlobRateLimiter = rate.NewLimiter(rate.Every(time.Second/5), 1)    // 5 requests/s
var mdAtHomeRateLimiter = rate.NewLimiter(rate.Every(time.Minute/40), 1) // 40 requests/minute

// JoinURL joins the given URL parts.
func JoinURL(base string, paths ...string) string {
	u, _ := url.Parse(base)
	for _, path := range paths {
		u.Path = filepath.Join(u.Path, strings.TrimLeft(strings.TrimRight(path, "/"), "/"))
	}
	return u.String()
}

// rgx is a regexp for extracting the page number from the file name.
var rgx = regexp.MustCompile("[0-9]+")

var sourceWhitelist = []string{
	"mangadex.org",
	"uploads.mangadex.org",
}

func isSourceAllowed(source string) bool {
	u, err := url.Parse(source)
	if err != nil {
		return false
	}
	return stringsContains(sourceWhitelist, u.Host)
}

// downloadFile downloads the given URL to a temporary file and returns the file.
// The caller is responsible for closing and removing the file.
func downloadFile(source string) (*os.File, error) {
	if !isSourceAllowed(source) {
		return nil, errors.New("Source is not allowed")
	}

	tmp, err := os.CreateTemp(GetTempDir(), "tmp-")
	if err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	res, err := http.Get(source)
	if err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}
	defer res.Body.Close()

	if _, err = io.Copy(tmp, res.Body); err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}
	return tmp, nil
}

// getPageNum extracts the page number from the given file name.
//
// It trims left zero padding and returns the first consecutive
// sequence of digits.
func getPageNum(fileName string) int {
	fileName = strings.TrimLeft(fileName, "0")
	n, _ := strconv.Atoi(rgx.FindString(fileName))
	return n
}

// makeCacheKey creates a cache key.
//
// This function simply calls json.Marshal on the given object
// and returns the result as a string.
func makeCacheKey(v interface{}) string {
	buf, _ := json.Marshal(v)
	return string(buf)
}

var emailRgx = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

// isEmail checks if the given string is a valid email address.
func isEmail(e string) bool {
	return emailRgx.MatchString(e)
}

// stringsContains checks if the given string is in the given slice.
func stringsContains(slice []string, search string) bool {
	for _, str := range slice {
		if str == search {
			return true
		}
	}
	return false
}

type ResizeOptions struct {
	Width  int
	Height int
	Crop   bool
}

var resizer struct {
	Map map[string]*sync.Mutex
	sync.Mutex
	sync.Once
}

func init() {
	resizer.Map = make(map[string]*sync.Mutex)
}

func resizeImage(filepath, outputPath string, o ResizeOptions) error {
	resizer.Lock()
	mu, ok := resizer.Map[outputPath]
	if !ok {
		mu = &sync.Mutex{}
		resizer.Map[outputPath] = mu
	}
	resizer.Unlock()

	mu.Lock()
	defer func() {
		mu.Unlock()

		resizer.Lock()
		delete(resizer.Map, outputPath)
		resizer.Unlock()
	}()

	if ok {
		return nil
	}

	w := strconv.Itoa(o.Width)
	h := strconv.Itoa(o.Height)
	crop := strconv.FormatBool(o.Crop)

	// This is a workaround for vips consuming too much memory.
	//
	// Another todo is to use grpc or websocket to communicate with
	// the vips process instead of spawning a new process for every
	// resize, and restart the process if it consumes too much memory.
	//
	// Using another library such as imaging is not an option
	// because they're too slow.
	buf, err := runCommand(getImageBinPath(), filepath, w, h, crop)
	if err != nil {
		return err
	}

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = buf.WriteTo(out)
	return err
}

// sanitizeOrder sanitizes the given order.
func sanitizeOrder(order string) string {
	if strings.EqualFold(order, "asc") {
		return "asc"
	}
	return "desc"
}

// sanitizeChapterSort sanitizes the given chapter sort.
func sanitizeChapterSort(column string) string {
	switch {
	case strings.EqualFold(column, ChapterCols.ID):
		return ChapterCols.ID
	case strings.EqualFold(column, ChapterCols.UpdatedAt):
		return ChapterCols.UpdatedAt
	case strings.EqualFold(column, ChapterCols.PublishedAt):
		return ChapterCols.PublishedAt
	case strings.EqualFold(column, ChapterCols.Chapter):
		return ChapterCols.Chapter
	case strings.EqualFold(column, ChapterCols.Volume):
		return ChapterCols.Volume
	default:
		return ChapterCols.CreatedAt
	}
}

// sanitizeChapterRels sanitizes the given chapter relations
// and returns a list of relations that should be preloaded.
//
// The list of relations returns normalized relations and
// sorted in alphabetical order so that they can be used as
// a cache key.
func sanitizeChapterRels(allowStats bool, preloads ...string) (result []string) {
	for _, rel := range preloads {
		switch {
		case strings.EqualFold(rel, ChapterRels.Project):
			result = append(result, ChapterRels.Project)
		case strings.EqualFold(rel, ChapterRels.Uploader):
			result = append(result, ChapterRels.Uploader)
		case strings.EqualFold(rel, ChapterRels.ScanlationGroups):
			result = append(result, ChapterRels.ScanlationGroups)
		case strings.EqualFold(rel, ChapterRels.Statistic) && allowStats:
			result = append(result, ChapterRels.Statistic)
		}
	}
	sort.Strings(result)
	return
}

// sanitizeProjectSort sanitizes the given chapter sort.
func sanitizeProjectSort(column string) string {
	switch {
	case strings.EqualFold(column, ProjectCols.ID):
		return ProjectCols.ID
	case strings.EqualFold(column, ProjectCols.UpdatedAt):
		return ProjectCols.UpdatedAt
	case strings.EqualFold(column, ProjectCols.PublishedAt):
		return ProjectCols.PublishedAt
	case strings.EqualFold(column, ProjectCols.Title):
		return ProjectCols.Title
	default:
		return ChapterCols.CreatedAt
	}
}

// sanitizeProjectRels sanitizes the given project relations
// and returns a list of relations that should be preloaded.
//
// The list of relations returns normalized relations and
// sorted in alphabetical order so that they can be used as
// a cache key.
func sanitizeProjectRels(allowStats bool, preloads ...string) (result []string) {
	for _, v := range preloads {
		switch {
		case strings.EqualFold(v, ProjectRels.Cover):
			result = append(result, ProjectRels.Cover)
		case strings.EqualFold(v, ProjectRels.Artists):
			result = append(result, ProjectRels.Artists)
		case strings.EqualFold(v, ProjectRels.Authors):
			result = append(result, ProjectRels.Authors)
		case strings.EqualFold(v, ProjectRels.Statistic) && allowStats:
			result = append(result, ProjectRels.Statistic)
		case strings.EqualFold(v, ProjectRels.Tags):
			result = append(result, ProjectRels.Tags)
		}
	}
	sort.Strings(result)
	return result
}

// FormatChapter formats modext.Chapter into a human readable string,
// it concatenates the volume and chapter number, and the title if it exists.
//
// e.g. "Vol. 1 Ch. 1 - Title"
func FormatChapter(c *modext.Chapter, short ...interface{}) string {
	var str string
	if len(c.Volume) > 0 {
		str += fmt.Sprintf("Vol. %s ", c.Volume)
	}

	str += fmt.Sprintf("Ch. %s", c.Chapter)
	if len(c.Title) > 0 && len(short) == 0 {
		str += fmt.Sprintf(" - %s", c.Title)
	}
	return str
}

// formatChapterModel formats models.Chapter into a human readable string,
// it concatenates the volume and chapter number, and the title if it exists.
//
// e.g. "Vol. 1 Ch. 1 - Title"
func formatChapterModel(c *models.Chapter, short ...interface{}) string {
	var str string
	if len(c.Volume.String) > 0 {
		str += fmt.Sprintf("Vol. %s ", c.Volume.String)
	}

	str += fmt.Sprintf("Ch. %s", c.Chapter)
	if len(c.Title.String) > 0 && len(short) == 0 {
		str += fmt.Sprintf(" - %s", c.Title.String)
	}
	return str
}

// createProjectDir creates a directory for the given project.
//
// The directory will be created in the <data> directory, using the
// project's slug as the directory name. Chapters and covers directory
// will also be created inside the project directory. They will be
// created with os.ModePerm.
//
// A symbolic link to the project directory will be created in the
// <data>/symlinks directory, using the project's ID as the link name.
//
// The project directory will be renamed when the project's slug has
// been changed, and the symlink will be updated accordingly.
//
// The project's directory and its subdirectories, and the symlink
// which points to it, will be removed when the project is deleted.
func createProjectDir(p *models.Project) error {
	dir := getProjectDir(p.Slug)
	if err := os.MkdirAll(filepath.Join(dir, "chapters"), os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create chapters directory")
	}

	if err := os.MkdirAll(filepath.Join(dir, "covers"), os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create covers directory")
	}

	if err := os.Symlink(dir, getProjectSymlink(p.ID)); err != nil {
		return errors.Wrap(err, "failed to create project symlink")
	}
	return nil
}

// removeProjectDir removes the project directory and its
// subdirectories. It also removes the symlink which points to it.
func removeProjectDir(p *models.Project) error {
	dir := getProjectDir(p.Slug)
	sDir := GetChaptersSymlinksDir()

	walk := filepath.WalkDir // alias
	walkFn := func(cDir string, d fs.DirEntry, _ error) error {
		if !d.IsDir() {
			return nil
		}
		return walk(sDir, func(symlink string, i fs.DirEntry, _ error) error {
			if !i.IsDir() {
				if v, _ := os.Readlink(symlink); v == cDir {
					return os.Remove(symlink)
				}
			}
			return nil
		})
	}
	walk(filepath.Join(dir, "chapters"), walkFn)

	if err := os.Remove(getProjectSymlink(p.ID)); err != nil {
		return errors.Wrap(err, "failed to remove project symlink")
	}

	if err := os.RemoveAll(dir); err != nil {
		return errors.Wrap(err, "failed to remove project directory")
	}

	return nil
}

// createChapterDir creates a directory for the given chapter.
//
// The directory will be created inside the project's directory,
// using human readable format of the chapter as the directory name.
// It will be created with os.ModePerm.
//
// A symbolic link to the chapter directory will be created in the
// <data>/symlinks/chapters directory, using the chapter's id as the
// link name.
//
// The chapter directory will be renamed when the chapter's chapter,
// volume number or title has been changed, and the symlink will be
// updated accordingly.
//
// The chapter's directory and the symlink which points to it will
// be removed when the chapter is deleted.
func createChapterDir(c *models.Chapter) error {
	if c.ID <= 0 {
		return nil
	}

	pDir, err := getProjectDirByID(c.ProjectID)
	if err != nil {
		return err
	}

	dir := filepath.Join(pDir, "chapters", slug.Make(formatChapterModel(c)))
	for i := 1; true; i++ {
		if _, err := os.Stat(dir); !os.IsNotExist(err) {
			dir = fmt.Sprintf("%s_-_%d", dir, i)
		} else {
			break
		}
	}

	if err := os.MkdirAll(dir, os.ModePerm); err != nil {
		return errors.Wrap(err, "failed to create chapter directory")
	}

	if err := os.Symlink(dir, getChapterSymlink(c.ID)); err != nil {
		return errors.Wrap(err, "failed to create chapter symlink")
	}

	return nil
}

// renameChapterDir renames the chapter directory.
//
// The chapter directory will be renamed inside the project's directory,
// using human readable format of the chapter as the directory name.
//
// The chapter's directory and the symlink which points to it will
// be removed when the chapter is deleted.
func renameChapterDir(c *models.Chapter) error {
	if c.ID <= 0 {
		return nil
	}

	oldDir, err := getChapterDir(c.ID)
	if err != nil {
		return err
	}

	pDir, err := getProjectDirByID(c.ProjectID)
	if err != nil {
		return err
	}

	dir := filepath.Join(pDir, "chapters", slug.Make(formatChapterModel(c)))
	if oldDir == dir {
		return nil
	}

	if err := os.Rename(oldDir, dir); err != nil {
		return errors.Wrap(err, "failed to rename chapter directory")
	}

	symlink := getChapterSymlink(c.ID)
	if err := os.Remove(symlink); err != nil {
		return errors.Wrap(err, "failed to remove the old chapter symlink")
	}

	if err := os.Symlink(dir, symlink); err != nil {
		return errors.Wrap(err, "failed to create a new symlink")
	}

	return nil
}

// removeChapterDir removes the directory of the given chapter.
// It also removes the symlink which points to it.
func removeChapterDir(c *models.Chapter) error {
	symlink := getChapterSymlink(c.ID)

	dir, err := os.Readlink(symlink)
	if err != nil {
		return errors.Wrap(err, "failed to read chapter symlink")
	}

	if err := os.Remove(symlink); err != nil {
		return errors.Wrap(err, "failed to remove chapter symlink")
	}

	return os.RemoveAll(dir)
}

// renameProjectDir renames the project directory.
//
// The project directory will be renamed inside the <data>/projects
// directory, using the project's slug as the directory name.
//
// The project's directory and the symlink which points to it will
// be removed when the project is deleted.
func renameProjectDir(p *models.Project) error {
	oldDir, err := getProjectDirByID(p.ID)
	if err != nil {
		return err
	}

	dir := getProjectDir(p.Slug)
	if oldDir == dir {
		return nil
	}

	if err := os.Rename(oldDir, dir); err != nil {
		return errors.Wrap(err, "failed to rename project directory")
	}

	symlink := getProjectSymlink(p.ID)
	if err := os.Remove(symlink); err != nil {
		return errors.Wrap(err, "failed to remove the old project symlink")
	}

	if err := os.Symlink(dir, symlink); err != nil {
		return errors.Wrap(err, "failed to create a new symlink")
	}

	return nil
}

// getImageBinPath gets the absolute path of the image binary.
func getImageBinPath() string {
	exec, _ := filepath.Abs("kasen-image")

	appendArch := func() {
		currentName := os.Args[0]
		if strings.Contains(currentName, "_linux") {
			arch := strings.Split(currentName, "_")[1]
			exec = fmt.Sprintf("%s_%s", exec, arch)
		}
	}
	appendArch()

	if _, err := os.Stat(exec); os.IsNotExist(err) {
		exec, _ = filepath.Abs("/usr/local/bin/kasen-image")
		appendArch()
	}
	return exec
}

// GetDataDir gets the absolute path of the <data> directory.
func GetDataDir() string {
	return filepath.Join(config.GetDirectories().Root, "data")
}

// GetTempDir gets the absolute path of the <data>/temp directory.
func GetTempDir() string {
	return filepath.Join(GetDataDir(), "tmp")
}

// GetSymlinksDir gets the absolute path of the symlinks directory.
func GetSymlinksDir() string {
	return filepath.Join(GetDataDir(), "symlinks")
}

// GetChaptersSymlinksDir gets the absolute path of the chapters
// symlinks directory.
func GetChaptersSymlinksDir() string {
	return filepath.Join(GetSymlinksDir(), "chapters")
}

// getProjectSymlink gets the absolute path of the project symlink.
func getProjectSymlink(id int64) string {
	return filepath.Join(GetSymlinksDir(), strconv.Itoa(int(id)))
}

// getProjectDir gets the absolute path of the project directory.
func getProjectDir(slug string) string {
	return filepath.Join(GetDataDir(), slug)
}

// getProjectDirByID gets the absolute path of the project directory.
func getProjectDirByID(id int64) (string, error) {
	dir, err := os.Readlink(getProjectSymlink(id))
	if err != nil {
		return "", errors.Wrap(err, "failed to read project symlink")
	}
	return dir, nil
}

// getChapterSymlink gets the absolute path of the chapter symlink.
func getChapterSymlink(id int64) string {
	return filepath.Join(GetChaptersSymlinksDir(), strconv.Itoa(int(id)))
}

// getChapterDir gets the absolute path of the chapter directory.
func getChapterDir(id int64) (string, error) {
	dir, err := os.Readlink(getChapterSymlink(id))
	if err != nil {
		return "", errors.Wrap(err, "failed to read chapter symlink")
	}
	return dir, nil
}

// removeCoverFiles removes the files of the given cover.
// It also removes the resized cover files.
func removeCoverFiles(c *models.Cover) error {
	if c.ProjectID <= 0 || len(c.FileName) == 0 {
		return nil
	}

	dir, err := getProjectDirByID(c.ProjectID)
	if err != nil {
		return err
	} else if _, err := os.Stat(dir); os.IsNotExist(err) {
		return err
	}

	walkFn := func(p string, d fs.DirEntry, e error) error {
		if !d.IsDir() && strings.Contains(p, c.FileName) {
			os.Remove(p)
		}
		return e
	}
	return filepath.WalkDir(filepath.Join(dir, "covers"), walkFn)
}

func refreshTemplatesCache() {
	keys := TemplatesCache.Keys()
	TemplatesCache.Purge()

	for _, k := range keys {
		go func(k interface{}) {
			res, err := http.Get(k.(string))
			if err != nil {
				return
			}
			res.Body.Close()
		}(k)
	}
}

// This function will be called when the given chapter is updated.
func refreshChapterCache(id int64) {
	keys := ChapterCache.KeysWithPrefix(id)
	ChapterCache.PurgeWithPrefix(id)

	for _, k := range keys {
		opts := GetChapterOptions{}
		if err := json.Unmarshal([]byte(k), &opts); err != nil {
			logger.Err.Println(err)
			continue
		}
		GetChapter(id, opts)
	}
}

// This function will be called when a new chapter is created,
// or when a chapter is updated or deleted.
func refreshChaptersCache() {
	keys := ChapterCache.KeysWithPrefix("global")
	ChapterCache.PurgeWithPrefix("global")

	for _, k := range keys {
		opts := GetChaptersOptions{}
		if err := json.Unmarshal([]byte(k), &opts); err != nil {
			logger.Err.Println(err)
			continue
		}
		GetChapters(opts)
	}
}

// This function will be called when a new chapter is added to the given project,
// or when one of the chapters of the project is updated or deleted.
func refreshProjectChaptersCache(pid int64) {
	keys := ChapterCache.KeysWithPrefix(pid)
	ChapterCache.PurgeWithPrefix(pid)

	for _, k := range keys {
		opts := GetChaptersOptions{}
		if err := json.Unmarshal([]byte(k), &opts); err != nil {
			logger.Err.Println(err)
			continue
		}
		GetChapters(opts)
	}
}

// This function will be called when then given project is updated,
// or when the main cover of the project has been changed.
func refreshProjectCache(id int64) {
	keys := ProjectCache.KeysWithPrefix(id)
	ProjectCache.PurgeWithPrefix(id)

	for _, k := range keys {
		opts := GetProjectOptions{}
		if err := json.Unmarshal([]byte(k), &opts); err != nil {
			logger.Err.Println(err)
			continue
		}
		GetProject(id, opts)
	}
}

// This function will be called when a project is created, updated or deleted.
func refreshProjectsCache() {
	keys := ProjectCache.KeysWithPrefix("global")
	ProjectCache.PurgeWithPrefix("global")

	for _, k := range keys {
		opts := GetProjectsOptions{}
		if err := json.Unmarshal([]byte(k), &opts); err != nil {
			logger.Err.Println(err)
			continue
		}
		GetProjects(opts)
	}
}

// This function will be called when the main cover of the given project
// has been changed.
func refreshCoverCache(pid int64) {
	if !CoverCache.HasWithInt64(pid) {
		return
	}

	CoverCache.RemoveWithInt64(pid)
	GetCover(pid)
}

// This function will be called when a new cover is added to the given project,
// or one of the covers of the project is updated or deleted.
func refreshCoversCache(pid int64) {
	if !CoverCache.HasWithPrefix(pid, "covers") {
		return
	}

	CoverCache.RemoveWithPrefix(pid, "covers")
	GetCovers(pid)
}

// This function will be called when a page is added or removed from the
// given chapter.
func refreshPagesCache(cid int64, pages []string) {
	if !PagesCache.HasWithInt64(cid) {
		return
	}

	result := &GetPagesResult{}
	result.Pages = pages

	PagesCache.RemoveWithInt64(cid)
	PagesCache.SetWithInt64(cid, result, 0)
}

var remapSymlinksMutex sync.Mutex
var remapSymlinksState bool

// RemapSymlinks remaps the symbolic links of all projects and chapters.
//
// It will simply remove the symbolic links and recreate them.
// Use this when symbolic links are broken or when the working
// directory or data directory has been changed.
func RemapSymlinks() error {
	remapSymlinksMutex.Lock()
	state := remapSymlinksState
	remapSymlinksMutex.Unlock()

	if state {
		return nil
	}

	state = true
	defer func() {
		state = false
	}()

	symlinksDir := GetSymlinksDir()
	os.RemoveAll(symlinksDir)

	chaptersSymlinksDir := filepath.Join(symlinksDir, "chapters")
	os.MkdirAll(chaptersSymlinksDir, os.ModePerm)

	projects, err := models.Projects(qm.Load(ProjectRels.Chapters)).All(ReadDB)
	if err != nil {
		return err
	}

	for _, project := range projects {
		projectDir := getProjectDir(project.Slug)
		os.Symlink(projectDir, getProjectSymlink(project.ID))

		for _, chapter := range project.R.Chapters {
			dir := filepath.Join(projectDir, "chapters", slug.Make(formatChapterModel(chapter)))
			for i := 1; true; i++ {
				if _, err := os.Stat(dir); os.IsNotExist(err) {
					break
				}

				ok := false
				filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
					if err != nil || info.IsDir() || ok {
						return err
					}

					fileName := filepath.Base(path)
					for _, page := range chapter.Pages {
						if fileName == page {
							ok = true
							os.Symlink(dir, getChapterSymlink(chapter.ID))
							break
						}
					}

					return nil
				})

				if ok {
					break
				} else {
					dir = fmt.Sprintf("%s_-_%d", dir, i)
				}
			}
		}
	}
	return nil
}

// runCommand runs the given command and returns the output.
func runCommand(path string, args ...string) (*bytes.Buffer, error) {
	cmd := exec.Command(path, args...)
	cmd.Env = os.Environ()

	var buf bytes.Buffer
	var stderr bytes.Buffer

	cmd.Stdout = &buf
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return nil, err
	}

	if err := stderr.String(); len(err) > 0 {
		return nil, errors.New(err)
	}

	return &buf, nil
}

type Pagination struct {
	CurrentPage int
	Pages       []int
	TotalPages  int
}

const maxPages = 10

func CreatePagination(currentPage, totalPages int) *Pagination {
	if currentPage < 1 {
		currentPage = 1
	} else if currentPage > totalPages {
		currentPage = totalPages
	}

	pagination := &Pagination{
		CurrentPage: currentPage,
		TotalPages:  totalPages,
	}

	var first, last int
	if totalPages <= maxPages {
		first = 1
		last = totalPages
	} else {
		min := int(math.Floor(float64(maxPages) / 2))
		max := int(math.Ceil(float64(maxPages)/2)) - 1
		if currentPage <= min {
			first = 1
			last = maxPages
		} else if currentPage+max >= totalPages {
			first = totalPages - maxPages + 1
			last = totalPages
		} else {
			first = currentPage - min
			last = currentPage + max
		}
	}

	pagination.Pages = make([]int, last-first+1)
	for i := 0; i < last+1-first; i++ {
		pagination.Pages[i] = first + i
	}

	return pagination
}

type ChapterPagination struct {
	Previous *modext.Chapter
	Current  *modext.Chapter
	Next     *modext.Chapter
}

func CreateChapterPagination(currentChapter *modext.Chapter, chapters []*modext.Chapter) *ChapterPagination {
	if currentChapter == nil || len(chapters) == 0 {
		return nil
	}

	pagination := &ChapterPagination{
		Current: currentChapter,
	}

	currentIdx := -1
	for i, c := range chapters {
		if c.ID == currentChapter.ID {
			currentIdx = i
			break
		}
	}

	if currentIdx > 0 {
		pagination.Next = chapters[currentIdx-1]
	}

	if currentIdx < len(chapters)-1 {
		pagination.Previous = chapters[currentIdx+1]
	}

	return pagination
}
