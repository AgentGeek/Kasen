package services

import (
	"archive/zip"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	. "kasen/cache"
	. "kasen/database"

	"kasen/constants"
	"kasen/errs"
	"kasen/models"
	"kasen/modext"

	"github.com/gin-gonic/gin"
	"github.com/gosimple/slug"
	"github.com/pkg/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// GetChapterCacheStats gets the cache stats of the chapter LRU cache.
func GetChapterCacheStats() *CacheStats {
	return ChapterCache.GetStats()
}

// DownloadChapter downloads all pages of the given chapter as a zip.
func DownloadChapter(cid int64, rw gin.ResponseWriter) error {
	dir, err := getChapterDir(cid)
	if err != nil {
		return errors.Wrap(err, "Failed to get chapter dir")
	}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return errs.ErrChapterNotFound
	}

	projectName := filepath.Base(filepath.Dir(filepath.Dir(dir)))
	chapterName := filepath.Base(dir)

	rw.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%s_%s.zip", projectName, chapterName))

	w := zip.NewWriter(rw)
	defer w.Close()
	defer rw.Flush()

	walkFn := func(path string, info fs.DirEntry, err error) error {
		if err != nil || info.IsDir() {
			return err
		}

		fileName := filepath.Base(path)
		if len(filepath.Ext(strings.TrimSuffix(fileName, filepath.Ext(fileName)))) > 0 {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return errors.Wrap(err, "Failed to open page file")
		}
		defer file.Close()

		f, err := w.Create(fileName)
		if err != nil {
			return errors.Wrap(err, "Failed to create file in zip")
		}

		_, err = io.Copy(f, file)
		if err != nil {
			return errors.Wrap(err, "Failed to copy file to zip")
		}

		return nil
	}

	return filepath.WalkDir(dir, walkFn)
}

var ChapterCols = models.ChapterColumns
var ChapterRels = models.ChapterRels

// ChapterDraft represents a chapter draft.
type ChapterDraft struct {
	Chapter          string   `json:"chapter"`
	Volume           string   `json:"volume"`
	Title            string   `json:"title"`
	ScanlationGroups []string `json:"scanlationGroups"`
}

func (draft *ChapterDraft) validate() error {
	draft.Chapter = strings.TrimSpace(draft.Chapter)
	draft.Volume = strings.TrimSpace(draft.Volume)
	draft.Title = strings.TrimSpace(draft.Title)

	for i, group := range draft.ScanlationGroups {
		draft.ScanlationGroups[i] = strings.TrimSpace(group)
	}

	if len(draft.Chapter) == 0 {
		return errs.ErrChapterNumRequired
	} else if len(draft.Chapter) > 8 {
		return errs.ErrChapterNumTooLong
	} else if len(draft.Volume) > 8 {
		return errs.ErrChapterVolumeTooLong
	} else if len(draft.Title) > 128 {
		return errs.ErrChapterTitleTooLong
	}
	return nil
}

// This function simply calls CreateChapterEx with the global Write connection.
func CreateChapter(pid int64, draft ChapterDraft, uploader *modext.User) (*modext.Chapter, error) {
	tx, err := WriteDB.Begin()
	if err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}
	return CreateChapterEx(tx, pid, draft, uploader)
}

func refreshChapterRels(tx *sql.Tx, c *models.Chapter, draft *ChapterDraft) error {
	var scanlationGroups []*models.ScanlationGroup
	for _, g := range draft.ScanlationGroups {
		g, err := CreateScanlationGroupEx(tx, g)
		if err != nil {
			return err
		}
		scanlationGroups = append(scanlationGroups, &models.ScanlationGroup{
			ID:   g.ID,
			Slug: g.Slug,
			Name: g.Name,
		})
	}
	if err := c.SetScanlationGroups(tx, false, scanlationGroups...); err != nil {
		log.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

// CreateChapterEx creates a chapter for the given project.
func CreateChapterEx(tx *sql.Tx, pid int64, draft ChapterDraft, uploader *modext.User) (*modext.Chapter, error) {
	if err := draft.validate(); err != nil {
		return nil, err
	}

	p, err := models.FindProject(tx, pid)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrProjectNotFound
		}
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	c := &models.Chapter{
		ProjectID: p.ID,
		Chapter:   draft.Chapter,
		Volume:    null.StringFrom(draft.Volume),
		Title:     null.StringFrom(draft.Title),
	}

	if uploader != nil {
		c.UploaderID = null.Int64From(uploader.ID)
	}

	if err := c.Insert(tx, boil.Infer()); err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	if err := refreshChapterRels(tx, c, &draft); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	go refreshTemplatesCache()
	go createChapterDir(c)
	go func() {
		refreshProjectChaptersCache(c.ProjectID)
		refreshChaptersCache()
	}()

	return modext.NewChapter(c).LoadRels(c), nil
}

// GetChapterOptions represents the options for getting a chapter.
type GetChapterOptions struct {
	Preloads       []string `form:"preload" json:"1,omitempty"`
	IncludesDrafts bool     `form:"includesDrafts" json:"2,omitempty"`
}

// GetChapterResult represents the result of function GetChapter.
type GetChapterResult struct {
	Chapter *modext.Chapter `json:"data,omitempty"`
	Err     error           `json:"error,omitempty"`
}

// This function simply calls GetChapterEx with the global Read connection.
func GetChapter(cid int64, opts GetChapterOptions) *GetChapterResult {
	return GetChapterEx(ReadDB, cid, opts)
}

// GetChapterEx gets a chapter by id with the given options.
//
// The returned value will be cached in the LRU cache if
// Chapter or Err is not nil.
func GetChapterEx(e boil.Executor, cid int64, opts GetChapterOptions) (result *GetChapterResult) {
	opts.Preloads = sanitizeChapterRels(true, opts.Preloads...)

	cacheKey := makeCacheKey(opts)
	if c, err := ChapterCache.GetWithPrefix(cid, cacheKey); err == nil {
		return c.(*GetChapterResult)
	}

	result = &GetChapterResult{}
	defer func() {
		if result.Chapter != nil || result.Err != nil {
			ChapterCache.RemoveWithPrefix(cid, cacheKey)
			ChapterCache.SetWithPrefix(cid, cacheKey, result, 0)
		}
	}()

	selectQueries := []QueryMod{Where("id = ?", cid)}
	if !opts.IncludesDrafts {
		selectQueries = append(selectQueries,
			Where(`chapter.published_at IS NOT NULL
							AND chapter.project_id NOT IN
								(SELECT id FROM project WHERE published_at IS NULL)`),
		)
	}

	var loadStats bool
	for _, preload := range opts.Preloads {
		if strings.EqualFold(preload, ChapterRels.Statistic) {
			loadStats = true
		} else {
			selectQueries = append(selectQueries, Load(preload))
		}
	}

	c, err := models.Chapters(selectQueries...).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			result.Err = errs.ErrChapterNotFound
			return
		}
		log.Println(err)
		result.Err = errs.ErrUnknown
		return
	}

	result.Chapter = modext.NewChapter(c).LoadRels(c)
	if loadStats {
		result.Chapter.Stats = GetChapterStats(c.ID).Stats
	}

	return
}

// GetChapterMD gets chapter metadata from MangaDex.
func GetChapterMd(id string) (*ChapterDraft, error) {
	if strings.Contains(id, "/") {
		return nil, errors.New("Invalid chapter id")
	}

	mdGlobRateLimiter.Wait(context.Background())

	res, err := http.Get(fmt.Sprintf("%s/chapter/%s?includes[]=scanlation_group", mdBaseURL, id))
	if err != nil {
		log.Println(err)
		return nil, errs.ErrChapterMdFetchFailed
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusNotFound {
			return nil, errs.ErrChapterNotFound
		}
		return nil, errs.ErrChapterMdFetchFailed
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	body := &struct {
		Data struct {
			Attributes struct {
				Chapter string
				Volume  string
				Title   string
			}
			Relationships []struct {
				Type       string
				Attributes struct {
					Name     string
					FileName string
				}
			}
		}
	}{}
	if err := json.Unmarshal(buf, body); err != nil {
		return nil, err
	}

	draft := &ChapterDraft{
		Chapter: body.Data.Attributes.Chapter,
		Volume:  body.Data.Attributes.Volume,
		Title:   body.Data.Attributes.Title,
	}

	for _, rel := range body.Data.Relationships {
		if rel.Type == "scanlation_group" {
			draft.ScanlationGroups = append(draft.ScanlationGroups, rel.Attributes.Name)
		}
	}

	return draft, nil
}

// GetChaptersOptions represents the options for getting chapters.
type GetChaptersOptions struct {
	ProjectID      int64    `form:"projectId" json:"1,omitempty"`
	Uploader       string   `form:"uploader" json:"2,omitempty"`
	Groups         []string `form:"scanlation_group" json:"3,omitempty"`
	Limit          int      `form:"limit" json:"4,omitempty"`
	Offset         int      `form:"offset" json:"5,omitempty"`
	Preloads       []string `form:"preload" json:"6,omitempty" `
	Sort           string   `form:"sort" json:"7,omitempty"`
	Order          string   `form:"order" json:"8,omitempty"`
	IncludesDrafts bool     `form:"includesDrafts" json:"9,omitempty"`

	GetThumbnail bool `form:"-" json:"10,omitempty"`
	GetTags      bool `form:"-" json:"11,omitempty"`
	GetAll       bool `form:"-" json:"12,omitempty"`
}

func (opts *GetChaptersOptions) validate() {
	opts.Uploader = strings.ToLower(opts.Uploader)

	for i, group := range opts.Groups {
		opts.Groups[i] = slug.Make(group)
	}
	sort.Strings(opts.Groups)

	if opts.Limit <= 0 {
		opts.Limit = 20
	} else if opts.Limit > 100 {
		opts.Limit = 100
	}

	if opts.Offset < 0 {
		opts.Offset = 0
	}

	opts.Preloads = sanitizeChapterRels(false, opts.Preloads...)
	opts.Sort = sanitizeChapterSort(opts.Sort)
	opts.Order = sanitizeOrder(opts.Order)
}

// toQueries converts GetChaptersOptions to a list of queries.
//
// Returns selectQueries which will be used to select chapters, and
// countQueries which will be used to count the number of results.
func (opts *GetChaptersOptions) toQueries() (selectQueries, countQueries []QueryMod) {
	if opts.ProjectID > 0 {
		selectQueries = append(selectQueries,
			Where("project_id = ?", opts.ProjectID),
		)
	}

	if len(opts.Uploader) > 0 {
		selectQueries = append(selectQueries,
			InnerJoin("user_account ON user_account.id = chapter.uploader_id"),
			Where("user_account.name ILIKE ?", opts.Uploader),
		)
	}

	if len(opts.Groups) > 0 {
		selectQueries = append(selectQueries,
			InnerJoin("chapter_scanlation_groups cg ON cg.chapter_id = chapter.id"),
			InnerJoin("scanlation_group g ON g.id = cg.scanlation_group_id"),
		)
		for i, group := range opts.Groups {
			q := Where
			if i > 0 {
				q = Or
			}
			selectQueries = append(selectQueries, q("g.slug = ?", group))
		}
	}

	if !opts.IncludesDrafts {
		selectQueries = append(selectQueries,
			Where(`chapter.published_at IS NOT NULL
							AND chapter.project_id NOT IN
								(SELECT id FROM project WHERE published_at IS NULL)`),
		)
	}
	countQueries = append(countQueries, selectQueries...)

	selectQueries = append(selectQueries, Offset(opts.Offset))
	if opts.Sort == ChapterCols.Chapter || opts.Sort == ChapterCols.Volume {
		selectQueries = append(selectQueries,
			OrderBy(fmt.Sprintf(`
			CAST(NULLIF(regexp_replace(chapter.volume, '\D', '', 'g'), '') AS double precision) %[1]s,
			COALESCE(chapter.chapter, '')::bytea %[1]s`, opts.Order)),
		)
	} else {
		selectQueries = append(selectQueries,
			OrderBy(fmt.Sprintf("%s %s", opts.Sort, opts.Order)),
		)
	}

	if !opts.GetAll {
		selectQueries = append(selectQueries, Limit(opts.Limit))
	}

	for _, preload := range opts.Preloads {
		selectQueries = append(selectQueries, Load(preload))
	}

	if opts.GetThumbnail {
		selectQueries = append(selectQueries,
			Load(Rels(ChapterRels.Project, ProjectRels.Cover)),
		)
	}

	if opts.GetTags {
		selectQueries = append(selectQueries,
			Load(Rels(ChapterRels.Project, ProjectRels.Tags)),
		)
	}

	return
}

// GetChaptersResult represents the result of function GetChapters.
type GetChaptersResult struct {
	Chapters []*modext.Chapter `json:"data"`
	Total    int64             `json:"total"`
	Err      error             `json:"error,omitempty"`
}

// This function simply calls GetChaptersEx with the global Read connection.
func GetChapters(opts GetChaptersOptions) *GetChaptersResult {
	return GetChaptersEx(ReadDB, opts)
}

// GetChaptersEx gets chapters with the given options.
//
// The returned value will be cached in the LRU cache if
// Chapters, Total or Err is not empty.
func GetChaptersEx(e boil.Executor, opts GetChaptersOptions) (result *GetChaptersResult) {
	opts.validate()

	var prefix interface{}
	if opts.ProjectID > 0 {
		prefix = opts.ProjectID
	} else {
		prefix = "global"
	}
	cacheKey := makeCacheKey(opts)

	if c, err := ChapterCache.GetWithPrefix(prefix, cacheKey); err == nil {
		return c.(*GetChaptersResult)
	}

	result = &GetChaptersResult{Chapters: []*modext.Chapter{}}
	defer func() {
		if len(result.Chapters) > 0 || result.Total > 0 || result.Err != nil {
			ChapterCache.RemoveWithPrefix(prefix, cacheKey)
			ChapterCache.SetWithPrefix(prefix, cacheKey, result, time.Hour)
		}
	}()

	selectQueries, countQueries := opts.toQueries()
	chapters, err := models.Chapters(selectQueries...).All(e)
	if err != nil {
		log.Println(err)
		result.Err = errs.ErrUnknown
		return
	}

	count, err := models.Chapters(countQueries...).Count(e)
	if err != nil {
		log.Println(err)
		result.Err = errs.ErrUnknown
		return
	}

	result.Chapters = make([]*modext.Chapter, len(chapters))
	result.Total = count

	for i, c := range chapters {
		result.Chapters[i] = modext.NewChapter(c).LoadRels(c)
		if c.R != nil && c.R.Project != nil {
			if opts.GetThumbnail {
				result.Chapters[i].Project.LoadCover(c.R.Project)
			}
			if opts.GetTags {
				result.Chapters[i].Project.LoadTags(c.R.Project)
			}
		}
		if opts.GetThumbnail {
			result.Chapters[i].GetThumbnail(c)
		}
	}

	return
}

type ChapterMd struct {
	ID string `json:"id"`
	*ChapterDraft
}

// GetChaptersMd gets chapters of the given manga id from MangaDex.
func GetChaptersMd(mangaId string) ([]*ChapterMd, error) {
	if strings.Contains(mangaId, "/") {
		return nil, errors.New("Invalid manga id")
	}

	mdGlobRateLimiter.Wait(context.Background())

	u := fmt.Sprintf(
		"%s/manga/%s/feed?translatedLanguage[]=en&limit=100&order[chapter]=asc&includes[]=scanlation_group",
		mdBaseURL,
		mangaId,
	)
	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	body := &struct {
		Data []struct {
			ID         string
			Attributes struct {
				Chapter string
				Volume  string
				Title   string
			}
			Relationships []struct {
				Type       string
				Attributes struct {
					Name     string
					FileName string
				}
			}
		}
	}{}
	if err := json.Unmarshal(buf, body); err != nil {
		return nil, err
	}

	chapters := make([]*ChapterMd, len(body.Data))
	for i, c := range body.Data {
		draft := &ChapterMd{
			ID: c.ID,
			ChapterDraft: &ChapterDraft{
				Chapter: c.Attributes.Chapter,
				Volume:  c.Attributes.Volume,
				Title:   c.Attributes.Title,
			},
		}
		for _, rel := range c.Relationships {
			if rel.Type == "scanlation_group" {
				draft.ScanlationGroups = append(draft.ScanlationGroups, rel.Attributes.Name)
			}
		}
		chapters[i] = draft
	}
	return chapters, nil
}

// This function simply calls UpdateChapterEx with the global Write connection.
func UpdateChapter(id int64, draft *ChapterDraft, user *modext.User) (*modext.Chapter, error) {
	tx, err := WriteDB.Begin()
	if err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}
	return UpdateChapterEx(tx, id, draft, user)
}

// UpdateChapterEx updates a chapter
// and returns the updated chapter if successful.
//
// This function will return an error if the chapter is locked.
// It will also return an error if the user does not have the necessary permissions.
func UpdateChapterEx(tx *sql.Tx, id int64, draft *ChapterDraft, user *modext.User) (*modext.Chapter, error) {
	if err := draft.validate(); err != nil {
		return nil, err
	}

	c, err := models.FindChapter(tx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrChapterNotFound
		}
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	if c.Locked.Bool {
		return nil, errs.ErrChapterLocked
	}

	if !user.HasPermissions(constants.PermEditChapters) {
		if c.UploaderID.Int64 != user.ID || !user.HasPermissions(constants.PermEditChapter) {
			return nil, errs.ErrForbidden
		}
	}

	prevChapter := c.Chapter
	prevVolume := c.Volume.String
	prevTitle := c.Title.String

	c.Chapter = draft.Chapter
	c.Volume = null.StringFrom(draft.Volume)
	c.Title = null.StringFrom(draft.Title)

	if err := c.Update(tx, boil.Whitelist(
		ChapterCols.Chapter,
		ChapterCols.Volume,
		ChapterCols.Title,
		ChapterCols.UpdatedAt,
	)); err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	if err := refreshChapterRels(tx, c, draft); err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	if err := tx.Commit(); err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	if prevChapter != draft.Chapter || prevVolume != draft.Volume || prevTitle != draft.Title {
		go renameChapterDir(c)
	}

	go chapterAfterUpdateHook(c)
	return modext.NewChapter(c).LoadRels(c), nil
}

// This function simply calls PublishChapterEx with the global Write connection.
func PublishChapter(id int64, user *modext.User) (*modext.Chapter, error) {
	return PublishChapterEx(WriteDB, id, user)
}

// PublishChapterEx publishes a chapter
// and returns the updated chapter if successful.
//
// This function will return an error if the chapter is locked.
// It will also return an error if the user does not have the necessary permissions.
func PublishChapterEx(e boil.Executor, id int64, user *modext.User) (*modext.Chapter, error) {
	c, err := models.FindChapter(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrChapterNotFound
		}
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	if c.Locked.Bool {
		return nil, errs.ErrChapterLocked
	}

	if !user.HasPermissions(constants.PermPublishChapters) {
		if c.UploaderID.Int64 != user.ID || !user.HasPermissions(constants.PermPublishChapter) {
			return nil, errs.ErrForbidden
		}
	}

	updatedAt := c.UpdatedAt
	c.PublishedAt = null.TimeFrom(time.Now().UTC())

	if err := c.Update(e, boil.Whitelist(ChapterCols.PublishedAt)); err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	c.UpdatedAt = updatedAt
	go chapterAfterUpdateHook(c)
	return modext.NewChapter(c), nil
}

// This function simply calls UnpublishChapterEx with the global Write connection.
func UnpublishChapter(id int64, user *modext.User) (*modext.Chapter, error) {
	return UnpublishChapterEx(WriteDB, id, user)
}

// UnpublishChapterEx unpublishes a chapter
// and returns the updated chapter if successful.
//
// This function will return an error if the chapter is locked.
// It will also return an error if the user does not have the necessary permissions.
func UnpublishChapterEx(e boil.Executor, id int64, user *modext.User) (*modext.Chapter, error) {
	c, err := models.FindChapter(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrChapterNotFound
		}
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	if c.Locked.Bool {
		return nil, errs.ErrChapterLocked
	}

	if !user.HasPermissions(constants.PermUnpublishChapters) {
		if c.UploaderID.Int64 != user.ID || !user.HasPermissions(constants.PermUnpublishChapter) {
			return nil, errs.ErrForbidden
		}
	}

	updatedAt := c.UpdatedAt
	c.PublishedAt.Valid = false

	if err := c.Update(e, boil.Whitelist(ChapterCols.PublishedAt)); err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	c.UpdatedAt = updatedAt
	go chapterAfterUpdateHook(c)
	return modext.NewChapter(c), nil
}

// This function simply calls LockChapterEx with the global Write connection.
func LockChapter(id int64, user *modext.User) (*modext.Chapter, error) {
	return LockChapterEx(WriteDB, id, user)
}

// LockChapterEx locks a chapter
// and returns the updated chapter if successful.
//
// This function will return an error if the user does not have the necessary permissions.
func LockChapterEx(e boil.Executor, id int64, user *modext.User) (*modext.Chapter, error) {
	c, err := models.FindChapter(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrChapterNotFound
		}
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	if !user.HasPermissions(constants.PermLockChapters) {
		if c.UploaderID.Int64 != user.ID || !user.HasPermissions(constants.PermLockChapter) {
			return nil, errs.ErrForbidden
		}
	}

	updatedAt := c.UpdatedAt
	c.Locked = null.BoolFrom(true)

	if err := c.Update(e, boil.Whitelist(ChapterCols.Locked)); err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	c.UpdatedAt = updatedAt
	go chapterAfterUpdateHook(c)
	return modext.NewChapter(c), nil
}

// This function simply calls UnlockChapterEx with the global Write connection.
func UnlockChapter(id int64, user *modext.User) (*modext.Chapter, error) {
	return UnlockChapterEx(WriteDB, id, user)
}

// UnlockChapterEx unlocks a chapter
// and returns the updated chapter if successful.
//
// This function will return an error if the user does not have the necessary permissions.
func UnlockChapterEx(e boil.Executor, id int64, user *modext.User) (*modext.Chapter, error) {
	c, err := models.FindChapter(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrChapterNotFound
		}
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	if !user.HasPermissions(constants.PermUnlockChapters) {
		if c.UploaderID.Int64 != user.ID || !user.HasPermissions(constants.PermUnlockChapter) {
			return nil, errs.ErrForbidden
		}
	}

	updatedAt := c.UpdatedAt
	c.Locked = null.NewBool(false, false)

	if err := c.Update(e, boil.Whitelist(ChapterCols.Locked)); err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	c.UpdatedAt = updatedAt
	go chapterAfterUpdateHook(c)
	return modext.NewChapter(c), nil
}

// This function simply calls DeleteChapterEx with the global Write connection.
func DeleteChapter(id int64, user *modext.User) error {
	return DeleteChapterEx(WriteDB, id, user)
}

// DeleteChapterEx deletes a chapter.
//
// This function will return an error if the chapter is locked.
// It will also return an error if the user does not have the necessary permissions.
func DeleteChapterEx(e boil.Executor, id int64, user *modext.User) error {
	c, err := models.FindChapter(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrChapterNotFound
		}
		log.Println(err)
		return errs.ErrUnknown
	}

	if c.Locked.Bool {
		return errs.ErrChapterLocked
	}

	if !user.HasPermissions(constants.PermDeleteChapters) {
		if c.UploaderID.Int64 != user.ID || !user.HasPermissions(constants.PermDeleteChapter) {
			return errs.ErrForbidden
		}
	}

	if err := c.Delete(e); err != nil {
		log.Println(err)
		return errs.ErrUnknown
	}

	ChapterCache.PurgeWithPrefix(c.ID)
	PagesCache.RemoveWithInt64(c.ID)
	go refreshTemplatesCache()

	go removeChapterDir(c)
	go func() {
		refreshProjectChaptersCache(c.ProjectID)
		refreshChaptersCache()
	}()

	return nil
}
