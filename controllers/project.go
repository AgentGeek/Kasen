package controllers

import (
	"fmt"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"

	"kasen/errs"
	"kasen/server"
	"kasen/services"
)

func Cover(c *server.Context) {
	slug := c.Param("slug")
	fileName := c.Param("fileName")

	if len(slug) == 0 || len(fileName) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	str := strings.TrimPrefix(c.Param("width"), "/")
	width, _ := strconv.Atoi(strings.TrimSuffix(str, filepath.Ext(str)))

	c.Header("Cache-Control", "public, max-age=300")
	services.ServeCover(slug, fileName, int(width), c.Writer, c.Request)
}

type ProjectParams struct {
	ID   int64  `uri:"id"`
	Slug string `uri:"slug"`
}

func Project(c *server.Context) {
	params := &ProjectParams{}
	c.BindUri(params)

	page, _ := strconv.Atoi(c.Query("page"))
	if page <= 0 {
		page = 1
	}

	if params.ID <= 0 {
		c.SetData("error", errs.ErrProjectNotFound)
		c.HTML(http.StatusInternalServerError, "error.html")
		return
	}

	templateName := "project.html"
	if c.TryCache(templateName) {
		return
	}

	o := services.GetProjectOptions{
		Preloads: []string{
			services.ProjectRels.Artists,
			services.ProjectRels.Authors,
			services.ProjectRels.Cover,
			services.ProjectRels.Tags,
			services.ProjectRels.Statistic,
		},
	}

	pResult := services.GetProject(params.ID, o)
	if pResult.Err != nil {
		c.SetData("error", pResult.Err)
		c.HTML(http.StatusInternalServerError, "error.html")
		return
	}

	project := pResult.Project
	if project.Slug != strings.TrimPrefix(params.Slug, "/") {
		c.Redirect(http.StatusFound, fmt.Sprintf("/projects/%d/%s", project.ID, project.Slug))
		return
	}

	limit := 20
	cResult := services.GetChapters(services.GetChaptersOptions{
		ProjectID: pResult.Project.ID,

		Limit:  limit,
		Offset: limit * (page - 1),

		Preloads: []string{
			services.ChapterRels.Uploader,
			services.ChapterRels.ScanlationGroups,
			services.ChapterRels.Statistic,
		},
		GetThumbnail: true,
	})
	if cResult.Err != nil {
		c.SetData("error", cResult.Err)
		c.HTML(http.StatusInternalServerError, "error.html")
		return
	}

	services.IncreaseViewCount(project.Stats, c.ClientIP())
	project.Chapters = cResult.Chapters

	for _, c := range project.Chapters {
		c.Stats = services.GetChapterStats(c.ID).Stats
	}

	c.SetData("project", project)
	c.SetData("chapters", cResult.Chapters)
	c.SetData("totalChapters", cResult.Total)

	totalPages := int(math.Ceil(float64(cResult.Total) / float64(limit)))
	c.SetData("pagination", services.CreatePagination(page, totalPages))

	c.Cache(http.StatusOK, templateName)
}

type ProjectsQueries struct {
	Title                 string   `form:"title"`
	ProjectStatus         []string `form:"projectStatus"`
	SeriesStatus          []string `form:"seriesStatus"`
	Demographic           []string `form:"demographic"`
	Rating                []string `form:"rating"`
	ExcludedProjectStatus []string `form:"excludeProjectStatus"`
	ExcludedSeriesStatus  []string `form:"excludeSeriesStatus"`
	ExcludedDemographic   []string `form:"excludeDemographic"`
	ExcludedRating        []string `form:"excludeRating"`
	Artists               []string `form:"artist"`
	Authors               []string `form:"author"`
	Tags                  []string `form:"tag"`
	ExcludedTags          []string `form:"excludeTag"`
	Page                  int      `form:"page"`
	Sort                  string   `form:"sort"`
	Order                 string   `form:"order"`
}

func Projects(c *server.Context) {
	q := &ProjectsQueries{}
	c.BindQuery(q)

	templateName := "projects.html"
	if c.TryCache(templateName) {
		return
	}

	c.SetData("queries", q)
	c.SetData("hasQueries",
		len(q.Title) > 0 ||
			len(q.ProjectStatus) > 0 ||
			len(q.SeriesStatus) > 0 ||
			len(q.Demographic) > 0 ||
			len(q.Rating) > 0 ||
			len(q.ExcludedProjectStatus) > 0 ||
			len(q.ExcludedSeriesStatus) > 0 ||
			len(q.ExcludedDemographic) > 0 ||
			len(q.ExcludedRating) > 0 ||
			len(q.Artists) > 0 ||
			len(q.Authors) > 0 ||
			len(q.Tags) > 0 ||
			len(q.ExcludedTags) > 0)

	limit := 24
	result := services.GetProjects(services.GetProjectsOptions{
		Title:                 q.Title,
		ProjectStatus:         q.ProjectStatus,
		SeriesStatus:          q.SeriesStatus,
		Demographic:           q.Demographic,
		Rating:                q.Rating,
		ExcludedProjectStatus: q.ExcludedProjectStatus,
		ExcludedSeriesStatus:  q.ExcludedSeriesStatus,
		ExcludedDemographic:   q.ExcludedDemographic,
		ExcludedRating:        q.ExcludedRating,
		Artists:               q.Artists,
		Authors:               q.Authors,
		Tags:                  q.Tags,
		ExcludedTags:          q.ExcludedTags,
		Limit:                 limit,
		Offset:                limit * (q.Page - 1),
		Sort:                  q.Sort,
		Order:                 q.Order,

		Preloads: []string{
			services.ProjectRels.Cover,
			services.ProjectRels.Tags,
		},
	})
	c.SetData("projects", result.Projects)
	c.SetData("total", result.Total)

	totalPages := int(math.Ceil(float64(result.Total) / float64(limit)))
	c.SetData("pagination", services.CreatePagination(q.Page, totalPages))

	tags, _ := services.GetTags()
	c.SetData("tags", tags)

	c.Cache(http.StatusOK, templateName)
}
