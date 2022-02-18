package services

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"

	. "kasen/cache"
	. "kasen/database"

	"kasen/config"
	"kasen/constants"
	"kasen/errs"
	"kasen/models"
	"kasen/modext"

	"github.com/gabriel-vasile/mimetype"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var CoverCols = models.CoverColumns
var imageMimeTypes = []string{"image/png", "image/jpeg"}

// GetCoverCacheStats gets the cache stats of the cover LRU cache.
func GetCoverCacheStats() *CacheStats {
	return CoverCache.GetStats()
}

// ServeCover serves the cover file.
func ServeCover(slug, fn string, width int, rw http.ResponseWriter, r *http.Request) {
	dir := getProjectDir(slug)
	fp := filepath.Join(dir, "covers", fn)
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		rw.WriteHeader(http.StatusNotFound)
		return
	}

	if width > 0 && width <= 1024 && width%64 == 0 {
		original := fp
		fp = fmt.Sprintf("%s.%d.jpg", fp, width)

		if _, err := os.Stat(fp); os.IsNotExist(err) {
			o := ResizeOptions{
				Width:  width,
				Height: width * 3 / 2,
				Crop:   true,
			}

			err := resizeImage(original, fp, o)
			if err != nil {
				log.Println(err)
				fp = original
			}
		}
	}

	http.ServeFile(rw, r, fp)
}

// This function simply calls UploadCoverEx with the global Write connection.
func UploadCover(pid int64, fileName string, f *os.File, uploader *modext.User) (*modext.Cover, error) {
	return UploadCoverEx(WriteDB, pid, fileName, f, uploader)
}

// UploadCoverEx creates a cover for the given project
// and returns the created cover,
// or the existing cover if the file already exists.
func UploadCoverEx(e boil.Executor, pid int64, fileName string, f *os.File, uploader *modext.User) (*modext.Cover, error) {
	if f == nil {
		return nil, errs.ErrCoverInvalid
	}

	if !uploader.HasPermissions(constants.PermUploadCover) {
		return nil, errs.ErrForbidden
	}

	stat, err := f.Stat()
	if err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	if sz := int(stat.Size()); sz <= 0 {
		return nil, errs.ErrCoverInvalid
	} else if sz > config.GetService().CoverMaxFileSize {
		return nil, errs.ErrCoverTooLarge
	}

	f.Seek(0, io.SeekStart)
	mime, err := mimetype.DetectReader(f)
	if err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	if !stringsContains(imageMimeTypes, mime.String()) {
		return nil, errs.ErrCoverUnsupportedFormat
	}

	if !uploader.HasPermissions(constants.PermUploadCover) {
		return nil, errs.ErrForbidden
	}

	dir, err := getProjectDirByID(pid)
	if err != nil {
		return nil, err
	}

	hasher := sha256.New()
	f.Seek(0, io.SeekStart)
	if _, err := io.Copy(hasher, f); err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	hash := hasher.Sum(nil)
	hasher.Reset()

	ext := filepath.Ext(fileName)

	fn := fmt.Sprintf("%x%s", hash, ext)
	fp := filepath.Join(dir, "covers", fn)

	if _, err := os.Stat(fp); os.IsNotExist(err) {
		var buf bytes.Buffer
		f.Seek(0, io.SeekStart)
		if _, err = io.Copy(&buf, f); err != nil {
			log.Println(err)
			return nil, errs.ErrUnknown
		}

		if err := WriteFile(fp, buf.Bytes()); err != nil {
			log.Println(err)
			return nil, errs.ErrUnknown
		}
	}

	c, err := models.Covers(Where("project_id = ? AND file_name = ?", pid, fn)).One(e)
	if err == sql.ErrNoRows {
		c = &models.Cover{ProjectID: pid, FileName: fn}
		if err = c.Insert(e, boil.Infer()); err != nil {
			log.Println(err)
			return nil, errs.ErrUnknown
		}

		CoverCache.PurgeWithPrefix(c.ProjectID)
	} else if err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	return modext.NewCover(c), nil
}

// This function simlply calls UploadCoverMultipartEx with the global Write connection.
func UploadCoverMultipart(pid int64, fh *multipart.FileHeader, uploader *modext.User) (*modext.Cover, error) {
	return UploadCoverMultipartEx(WriteDB, pid, fh, uploader)
}

// UploadCoverMultipartEx uploads a cover from multipart.FileHeader for the given project
// and returns the created cover.
//
// This function simply converts the multipart.FileHeader to []byte and calls UploadCoverEx.
func UploadCoverMultipartEx(e boil.Executor, pid int64, fh *multipart.FileHeader, uploader *modext.User) (*modext.Cover, error) {
	tmp, err := os.CreateTemp(GetTempDir(), "tmp-")
	if err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	f, err := fh.Open()
	if err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	_, err = io.Copy(tmp, f)
	f.Close()

	if err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	return UploadCoverEx(e, pid, fh.Filename, tmp, uploader)
}

// This function simply calls UploadCoverFromSourceEx with the global Write connection.
func UploadCoverFromSource(pid int64, source string, uploader *modext.User) (*modext.Cover, error) {
	return UploadCoverFromSourceEx(WriteDB, pid, source, uploader)
}

// UploadCoverFromSourceEx uploads a cover from the given source for the given project and returns the created cover.
//
// This function simply downloads the file from the source, coverts it to []byte
// and calls UploadCoverEx.
func UploadCoverFromSourceEx(e boil.Executor, pid int64, source string, uploader *modext.User) (*modext.Cover, error) {
	tmp, err := downloadFile(source)
	if err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	fn := filepath.Base(source)
	return UploadCoverEx(e, pid, fn, tmp, uploader)
}

// GetCoverResult represents the result of function GetCover.
type GetCoverResult struct {
	Cover *modext.Cover `json:"data,omitempty"`
	Err   error         `json:"error,omitempty"`
}

// This function simply calls GetCoverEx with the global Read connection.
func GetCover(pid int64) *GetCoverResult {
	return GetCoverEx(ReadDB, pid)
}

// GetCoverEx returns the main cover of the given project.
//
// The returned value will be cached in the LRU cache if
// Cover or Err is not nil.
func GetCoverEx(e boil.Executor, pid int64) (result *GetCoverResult) {
	if c, err := CoverCache.GetWithInt64(pid); err == nil {
		return c.(*GetCoverResult)
	}

	result = &GetCoverResult{}
	defer func() {
		if result.Cover != nil || result.Err != nil {
			CoverCache.RemoveWithInt64(pid)
			CoverCache.SetWithInt64(pid, result, 0)
		}
	}()

	p, err := models.Projects(
		Where("id = ?", pid),
		Load(ProjectRels.Cover),
	).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			result.Err = errs.ErrProjectNotFound
			return
		}
		log.Println(err)
		result.Err = errs.ErrUnknown
		return
	}

	if p.R != nil && p.R.Cover != nil {
		result.Cover = modext.NewCover(p.R.Cover)
	}
	return
}

// GetCoversResult represents the result of function GetCovers.
type GetCoversResult struct {
	Covers []*modext.Cover `json:"data,omitempty"`
	Err    error           `json:"error,omitempty"`
}

// This function simply calls GetCoversEx with the global Read connection.
func GetCovers(pid int64) *GetCoversResult {
	return GetCoversEx(ReadDB, pid)
}

// GetCoversEx returns the covers of the given project.
//
// The returned value will be cached in the LRU cache if
// Covers or Err is not nil.
func GetCoversEx(e boil.Executor, pid int64) (result *GetCoversResult) {
	cacheKey := "covers"
	if c, err := CoverCache.GetWithPrefix(pid, cacheKey); err == nil {
		return c.(*GetCoversResult)
	}

	result = &GetCoversResult{}
	defer func() {
		if result.Covers != nil || result.Err != nil {
			CoverCache.RemoveWithPrefix(pid, cacheKey)
			CoverCache.SetWithPrefix(pid, cacheKey, result, 0)
		}
	}()

	p, err := models.Projects(
		Where("id = ?", pid),
		Load(ProjectRels.Covers, OrderBy("id ASC")),
	).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			result.Err = errs.ErrProjectNotFound
			return
		}
		log.Println(err)
		result.Err = errs.ErrUnknown
		return
	}

	if p.R != nil && p.R.Covers != nil {
		result.Covers = make([]*modext.Cover, len(p.R.Covers))
		for i, c := range p.R.Covers {
			result.Covers[i] = modext.NewCover(c)
		}
	}
	return
}

// This function simply calls SetCoverEx with the global Write connection.
func SetCover(pid, cid int64) error {
	return SetCoverEx(WriteDB, pid, cid)
}

// SetCoverEx sets the main cover of the given project.
func SetCoverEx(e boil.Executor, pid, cid int64) error {
	p, err := models.Projects(Where("id = ?", pid), Load(ProjectRels.Covers)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrProjectNotFound
		}
		log.Println()
		return errs.ErrUnknown
	}

	if p.Locked.Bool {
		return errs.ErrProjectLocked
	}

	exists := false
	for _, c := range p.R.Covers {
		if c.ID == cid {
			exists = true
			break
		}
	}

	if !exists {
		return errs.ErrCoverNotFound
	}

	p.CoverID.Int64 = cid
	p.CoverID.Valid = true

	if err := p.Update(e, boil.Whitelist(ProjectCols.CoverID, ProjectCols.UpdatedAt)); err != nil {
		log.Println(err)
		return errs.ErrUnknown
	}

	go func() {
		refreshProjectCache(pid)
		refreshProjectsCache()
	}()

	refreshCoverCache(pid)

	return nil
}

// This function simply calls DeleteCoverEx with the global Write connection.
func DeleteCover(id int64) error {
	return DeleteCoverEx(WriteDB, id)
}

// DeleteCoverEx deletes a cover.
func DeleteCoverEx(e boil.Executor, id int64) error {
	c, err := models.FindCover(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrCoverNotFound
		}
		log.Println(err)
		return errs.ErrUnknown
	}

	if err := c.Delete(e); err != nil {
		log.Println(err)
		return errs.ErrUnknown
	}

	CoverCache.PurgeWithPrefix(c.ProjectID)

	go refreshTemplatesCache()
	go removeCoverFiles(c)

	return nil
}
