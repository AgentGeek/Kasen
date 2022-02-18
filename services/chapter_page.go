package services

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	. "kasen/cache"
	. "kasen/database"

	"kasen/config"
	"kasen/constants"
	"kasen/errs"
	"kasen/models"
	"kasen/modext"

	"github.com/gabriel-vasile/mimetype"
	"github.com/pkg/errors"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

// GetCoverCacheStats gets the cache stats of the cover LRU cache.
func GetPagesCacheStats() *CacheStats {
	return PagesCache.GetStats()
}

// ServePage serves the page file.
func ServePage(id int64, fn string, width int, w http.ResponseWriter, r *http.Request) {
	dir, err := getChapterDir(id)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fp := filepath.Join(dir, fn)
	if _, err := os.Stat(fp); os.IsNotExist(err) {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if width > 0 && width <= 1024 && width%64 == 0 {
		original := fp
		fp = fmt.Sprintf("%s.%d.jpg", fp, width)

		if _, err := os.Stat(fp); os.IsNotExist(err) {
			if err := resizeImage(original, fp, ResizeOptions{Width: width}); err != nil {
				log.Println(err)
				fp = original
			}
		}
	}

	http.ServeFile(w, r, fp)
}

// This function simply calls GetCoverEx with the global Write connection.
func UploadPage(cid int64, fileName string, f *os.File, uploader *modext.User) ([]string, error) {
	return UploadPageEx(WriteDB, cid, fileName, f, uploader)
}

// UploadPageEx uploads a page for the given chapter
// and returns the updated chapter pages.
func UploadPageEx(e boil.Executor, cid int64, fileName string, f *os.File, uploader *modext.User) ([]string, error) {
	if f == nil {
		return nil, errs.ErrPageInvalid
	}

	stat, err := f.Stat()
	if err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	if sz := int(stat.Size()); sz <= 0 {
		return nil, errs.ErrPageInvalid
	} else if sz > config.GetService().PageMaxFileSize {
		return nil, errs.ErrPageTooLarge
	}

	f.Seek(0, io.SeekStart)
	mime, err := mimetype.DetectReader(f)
	if err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	if !stringsContains(imageMimeTypes, mime.String()) {
		return nil, errs.ErrPageUnsupportedFormat
	}

	c, err := models.FindChapter(e, cid)
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

	if !uploader.HasPermissions(constants.PermEditChapters) {
		if c.UploaderID.Int64 != uploader.ID || !uploader.HasPermissions(constants.PermEditChapter) {
			return nil, errs.ErrForbidden
		}
	}

	dir, err := getChapterDir(cid)
	if err != nil {
		return nil, err
	}

	pageNum := getPageNum(fileName)
	if pageNum == 0 {
		pageNum = len(c.Pages) + 1
	}

	hasher := sha256.New()
	f.Seek(0, io.SeekStart)
	if _, err := io.Copy(hasher, f); err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	hash := hasher.Sum(nil)
	hasher.Reset()

	ext := mime.Extension()

	fn := fmt.Sprintf("%d-%x%s", pageNum, hash, ext)
	fp := filepath.Join(dir, fn)

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

	if !stringsContains(c.Pages, fn) {
		c.Pages = append(c.Pages, fn)
		sort.SliceStable(c.Pages, func(i, j int) bool {
			prev, _ := strconv.ParseInt(rgx.FindString(c.Pages[i]), 10, 64)
			next, _ := strconv.ParseInt(rgx.FindString(c.Pages[j]), 10, 64)
			return prev < next
		})

		if err := c.Update(e, boil.Whitelist(ChapterCols.Pages, ChapterCols.UpdatedAt)); err != nil {
			log.Println(err)
			return nil, errs.ErrUnknown
		}

		refreshPagesCache(cid, c.Pages)
		go chapterAfterUpdateHook(c)
	}

	return c.Pages, nil
}

// This function simply calls UploadPageMultipartEx with the global Write connection.
func UploadPageMultipart(cid int64, fh *multipart.FileHeader, uploader *modext.User) ([]string, error) {
	return UploadPageMultipartEx(WriteDB, cid, fh, uploader)
}

// UploadPageMultipartEx uploads a page from multipart.FileHeader for the given chapter
// and returns the updated chapter pages.
//
// This function simply converts the multipart.FileHeader to []byte and calls UploadPageEx.
func UploadPageMultipartEx(e boil.Executor, cid int64, fh *multipart.FileHeader, uploader *modext.User) ([]string, error) {
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

	return UploadPageEx(e, cid, fh.Filename, tmp, uploader)
}

// This function simply calls UploadPageFromSourceEx with the global Write connection.
func UploadPageFromSource(cid int64, source string, uploader *modext.User) ([]string, error) {
	return UploadPageFromSourceEx(WriteDB, cid, source, uploader)
}

// UploadPageFromSourceEx uploads a page from the given source for the given chapter
// and returns the updated chapter pages.
//
// This function simply downloads the file from the source, converts it to []byte
// and calls UploadPageEx.
func UploadPageFromSourceEx(e boil.Executor, cid int64, source string, uploader *modext.User) ([]string, error) {
	tmp, err := downloadFile(source)
	if err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}
	defer tmp.Close()
	defer os.Remove(tmp.Name())

	fn := filepath.Base(source)
	return UploadPageEx(WriteDB, cid, fn, tmp, uploader)
}

// GetPagesResult represents the result of function GetPages.
type GetPagesResult struct {
	Pages []string `json:"data,omitempty"`
	Err   error    `json:"error,omitempty"`
}

// This function simply calls GetPagesEx with the global Read connection.
func GetPages(cid int64) *GetPagesResult {
	return GetPagesEx(ReadDB, cid)
}

// GetPagesEx returns the pages of the given chapter.
//
// The returned value will be cached in the LRU cache if
// Pages or Err is not empty.
func GetPagesEx(e boil.Executor, cid int64) (result *GetPagesResult) {
	if c, err := PagesCache.GetWithInt64(cid); err == nil {
		return c.(*GetPagesResult)
	}

	result = &GetPagesResult{Pages: []string{}}
	defer func() {
		if result.Pages != nil || result.Err != nil {
			PagesCache.RemoveWithInt64(cid)
			PagesCache.SetWithInt64(cid, result, time.Hour)
		}
	}()

	c, err := models.FindChapter(e, cid, ChapterCols.Pages)
	if err != nil {
		if err == sql.ErrNoRows {
			result.Err = errs.ErrChapterNotFound
			return
		}
		log.Println(err)
		result.Err = errs.ErrUnknown
		return
	}

	result.Pages = c.Pages
	return
}

// PagesMd represents the result of function GetPagesMd.
type PagesMd struct {
	BaseURL string   `json:"baseUrl"`
	Hash    string   `json:"hash"`
	Pages   []string `json:"pages"`
}

// GetPagesMd gets the chapter pages from MangaDex.
func GetPagesMd(id string) (*PagesMd, error) {
	if strings.Contains(id, "/") {
		return nil, errors.New("Invalid chapter id")
	}

	mdGlobRateLimiter.Wait(context.Background())
	mdAtHomeRateLimiter.Wait(context.Background())

	res, err := http.Get(fmt.Sprintf("https://api.mangadex.org/at-home/server/%s", id))
	if err != nil {
		log.Println(err)
		return nil, errs.ErrPageMdFetchFailed
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusNotFound {
			return nil, errs.ErrChapterNotFound
		}
		return nil, errs.ErrPageMdFetchFailed
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	body := &struct {
		BaseURL string
		Chapter struct {
			Hash string
			Data []string
		}
	}{}
	if err := json.Unmarshal(buf, body); err != nil {
		return nil, err
	}

	return &PagesMd{
		BaseURL: body.BaseURL,
		Hash:    body.Chapter.Hash,
		Pages:   body.Chapter.Data,
	}, nil
}

// This function simply calls DeletePageEx with the global Write connection.
func DeletePage(cid int64, fileName string, user *modext.User) ([]string, error) {
	return DeletePageEx(WriteDB, cid, fileName, user)
}

// DeletePageEx deletes a page of the given chapter
// and returns the updated chapter pages.
func DeletePageEx(e boil.Executor, cid int64, fileName string, user *modext.User) ([]string, error) {
	c, err := models.FindChapter(e, cid)
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

	for i, fn := range c.Pages {
		if strings.EqualFold(fn, fileName) {
			c.Pages = append(c.Pages[:i], c.Pages[i+1:]...)
			break
		}
	}

	dir, err := getChapterDir(c.ID)
	if err != nil {
		return nil, err
	}

	if _, err := os.Stat(dir); !os.IsNotExist(err) {
		err := filepath.Walk(dir, func(path string, info fs.FileInfo, err error) error {
			if !info.IsDir() && strings.Contains(filepath.Base(path), fileName) {
				os.Remove(path)
			}
			return err
		})
		if err != nil {
			log.Println(err)
		}
	}

	if err := c.Update(e, boil.Whitelist(
		ChapterCols.Pages,
		ChapterCols.UpdatedAt,
	)); err != nil {
		log.Println(err)
		return nil, errs.ErrUnknown
	}

	refreshPagesCache(cid, c.Pages)
	go chapterAfterUpdateHook(c)
	return c.Pages, nil
}
