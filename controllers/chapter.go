package controllers

import (
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"kasen/modext"
	"kasen/server"
	"kasen/services"
)

func Chapter(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.HTML(http.StatusBadRequest, "error.html")
		return
	}

	if strings.EqualFold(c.Query("download"), "true") {
		if err := services.DownloadChapter(id, c.Writer); err != nil {
			c.Status(http.StatusInternalServerError)
		}
		return
	}

	legacyQuery := c.Query("legacy")
	legacyCookie, _ := c.Cookie("legacy")

	var isLegacy bool
	if strings.EqualFold(legacyQuery, "false") {
		c.SetCookie("legacy", "", nil)
	} else {
		isLegacyQuery := strings.EqualFold(legacyQuery, "true")
		isLegacy = isLegacyQuery || strings.EqualFold(legacyCookie, "true")

		if isLegacyQuery {
			expr := time.Now().Add(time.Hour * 24 * 365)
			c.SetCookie("legacy", "true", &expr)
		}
	}

	templateName := "reader.html"
	if isLegacy {
		templateName = "reader_legacy.html"
	}

	if c.IsCached(templateName) {
		go func(ip string) {
			result := services.GetChapterStats(id)
			if result.Err == nil {
				services.IncreaseViewCount(result.Stats, ip)
			}
		}(c.ClientIP())
	} else {
		var chapter *modext.Chapter
		{
			result := services.GetChapter(id, services.GetChapterOptions{
				Preloads: []string{
					services.ChapterRels.Project,
					services.ChapterRels.Statistic,
				},
			})
			if result.Err != nil {
				c.SetData("error", result.Err)
				c.HTML(http.StatusBadRequest, "error.html")
				return
			}
			chapter = result.Chapter
		}
		{
			result := services.GetPages(id)
			if result.Err != nil {
				c.SetData("error", result.Err)
				c.HTML(http.StatusBadRequest, "error.html")
				return
			}
			chapter.Pages = result.Pages
		}

		var chapters []*modext.Chapter
		{
			result := services.GetChapters(services.GetChaptersOptions{
				ProjectID: chapter.Project.ID,
				Sort:      services.ChapterCols.Chapter,
				GetAll:    true,
			})
			if result.Err != nil {
				c.SetData("error", result.Err)
				c.HTML(http.StatusBadRequest, "error.html")
			}
			chapters = result.Chapters
		}

		c.SetData("chapter", chapter)
		c.SetData("chapters", chapters)
		c.SetData("pagination", services.CreateChapterPagination(chapter, chapters))

		services.IncreaseViewCount(chapter.Stats, c.ClientIP())
		chapter.Stats = nil
	}

	c.Cache(http.StatusOK, templateName)
}

type ChaptersQueries struct {
	Uploader string   `form:"uploader"`
	Groups   []string `form:"scanlation_group"`
	Page     int      `form:"page"`
	Sort     string   `form:"sort"`
	Order    string   `form:"order"`
}

func Chapters(c *server.Context) {
	templateName := "chapters.html"
	if c.TryCache(templateName) {
		return
	}

	q := &ChaptersQueries{}
	c.BindQuery(q)

	c.SetData("query", q)
	c.SetData("hasQueries", len(q.Uploader) > 0 || len(q.Groups) > 0)

	limit := 24
	result := services.GetChapters(services.GetChaptersOptions{
		Uploader: q.Uploader,
		Groups:   q.Groups,

		Limit:  limit,
		Offset: limit * (q.Page - 1),

		Sort:  q.Sort,
		Order: q.Order,

		Preloads: []string{
			services.ChapterRels.Uploader,
			services.ChapterRels.ScanlationGroups,
		},

		GetThumbnail: true,
		GetTags:      true,
	})
	c.SetData("chapters", result.Chapters)
	c.SetData("total", result.Total)

	totalPages := int(math.Ceil(float64(result.Total) / float64(limit)))
	c.SetData("pagination", services.CreatePagination(q.Page, totalPages))

	c.Cache(http.StatusOK, templateName)
}

func Page(c *server.Context) {
	id, err := c.ParamInt64("id")
	fileName := c.Param("fileName")

	if err != nil || len(fileName) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	str := strings.TrimPrefix(c.Param("width"), "/")
	width, _ := strconv.Atoi(strings.TrimSuffix(str, filepath.Ext(str)))

	c.Header("Cache-Control", "public, max-age=300")
	services.ServePage(id, fileName, int(width), c.Writer, c.Request)
}
