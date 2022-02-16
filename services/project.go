package services

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	. "kasen/cache"
	. "kasen/database"

	"kasen/errs"
	"kasen/models"
	"kasen/modext"

	"github.com/gosimple/slug"
	"github.com/pkg/errors"
	"github.com/rs1703/logger"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

var (
	Demographic   = []string{"none", "shounen", "shoujo", "josei", "seinen"}
	ProjectStatus = []string{"ongoing", "finished", "dropped"}
	Rating        = []string{"none", "safe", "suggestive", "erotica", "pornographic"}
	SeriesStatus  = []string{"ongoing", "completed", "hiatus", "cancelled"}
)

// GetProjectCacheStats gets the cache stats of the project LRU cache.
func GetProjectCacheStats() *CacheStats {
	return ProjectCache.GetStats()
}

var ProjectRels = models.ProjectRels
var ProjectCols = models.ProjectColumns

// ProjectDraft represents a project draft.
type ProjectDraft struct {
	Title         string   `json:"title"`
	Description   string   `json:"description,omitempty"`
	CoverURL      string   `json:"coverUrl,omitempty"`
	ProjectStatus string   `json:"projectStatus,omitempty"`
	SeriesStatus  string   `json:"seriesStatus,omitempty"`
	Demographic   string   `json:"demographic,omitempty"`
	Rating        string   `json:"rating,omitempty"`
	Artists       []string `json:"artists,omitempty"`
	Authors       []string `json:"authors,omitempty"`
	Tags          []string `json:"tags,omitempty"`
}

func (draft *ProjectDraft) validate() error {
	draft.Title = strings.TrimSpace(draft.Title)
	draft.Description = strings.TrimSpace(draft.Description)
	draft.CoverURL = strings.TrimSpace(draft.CoverURL)
	draft.ProjectStatus = strings.TrimSpace(draft.ProjectStatus)
	draft.SeriesStatus = strings.TrimSpace(draft.SeriesStatus)
	draft.Demographic = strings.TrimSpace(draft.Demographic)
	draft.Rating = strings.TrimSpace(draft.Rating)

	for i := range draft.Artists {
		draft.Artists[i] = strings.TrimSpace(draft.Artists[i])
	}

	for i := range draft.Authors {
		draft.Authors[i] = strings.TrimSpace(draft.Authors[i])
	}

	for i := range draft.Tags {
		draft.Tags[i] = strings.TrimSpace(draft.Tags[i])
	}

	switch {
	case len(draft.Title) == 0:
		return errs.ErrProjectTitleRequired
	case len(draft.Title) > 128:
		return errs.ErrProjectTitleTooLong
	case len(draft.Description) > 4096:
		return errs.ErrProjectDescriptionTooLong
	case len(draft.ProjectStatus) == 0:
		return errs.ErrProjectStatusRequired
	case !stringsContains(ProjectStatus, draft.ProjectStatus):
		return errs.ErrInvalidProjectStatus
	case len(draft.SeriesStatus) == 0:
		return errs.ErrProjectSeriesStatusRequired
	case !stringsContains(SeriesStatus, draft.SeriesStatus):
		return errs.ErrInvalidSeriesStatus
	case len(draft.Demographic) > 0 && !stringsContains(Demographic, draft.Demographic):
		return errs.ErrInvalidDemographic
	case len(draft.Rating) > 0 && !stringsContains(Rating, draft.Rating):
		return errs.ErrInvalidRating
	}
	return nil
}

// This function simply calls CreateProjectEx with a new write transaction.
func CreateProject(draft *ProjectDraft) (*modext.Project, error) {
	tx, err := WriteDB.Begin()
	if err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	return CreateProjectEx(tx, draft)
}

func refreshProjectRels(tx *sql.Tx, p *models.Project, draft *ProjectDraft) error {
	var artists []*models.Author
	for _, a := range draft.Artists {
		a, err := CreateAuthorEx(tx, a)
		if err != nil {
			return err
		}
		artists = append(artists, &models.Author{
			ID:   a.ID,
			Slug: a.Slug,
			Name: a.Name,
		})
	}
	if err := p.SetArtists(tx, false, artists...); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	var authors []*models.Author
	for _, a := range draft.Authors {
		a, err := CreateAuthorEx(tx, a)
		if err != nil {
			return err
		}
		authors = append(authors, &models.Author{
			ID:   a.ID,
			Slug: a.Slug,
			Name: a.Name,
		})
	}
	if err := p.SetAuthors(tx, false, authors...); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	var tags []*models.Tag
	for _, t := range draft.Tags {
		t, err := CreateTagEx(tx, t)
		if err != nil {
			return err
		}
		tags = append(tags, &models.Tag{
			ID:   t.ID,
			Slug: t.Slug,
			Name: t.Name,
		})
	}
	if err := p.SetTags(tx, false, tags...); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	return nil
}

// CreateProjectEx creates a new project.
func CreateProjectEx(tx *sql.Tx, draft *ProjectDraft) (*modext.Project, error) {
	defer logger.Track()()

	if err := draft.validate(); err != nil {
		return nil, err
	}

	p := &models.Project{
		Title:         draft.Title,
		Description:   null.StringFrom(draft.Description),
		ProjectStatus: draft.ProjectStatus,
		SeriesStatus:  draft.SeriesStatus,
		Demographic:   null.StringFrom(draft.Demographic),
		Rating:        null.StringFrom(draft.Rating),
	}
	if err := p.Insert(tx, boil.Infer()); err != nil {
		if strings.Contains(err.Error(), `unique constraint "project_slug"`) {
			return nil, errs.ErrProjectAlreadyExists
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	if err := refreshProjectRels(tx, p, draft); err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	go createProjectDir(p)
	go refreshProjectsCache()

	return modext.NewProject(p).LoadRels(p), nil
}

// GetProjectOptions represents the options for getting a project.
type GetProjectOptions struct {
	Preloads       []string `form:"preload" json:"1,omitempty"`
	IncludesDrafts bool     `form:"includesDrafts" json:"2,omitempty"`
}

// GetProjectResult represents the result of GetProject.
type GetProjectResult struct {
	Project *modext.Project `json:"data"`
	Err     error           `json:"error,omitempty"`
}

// This function simply calls GetProjectEx with the global Read connection.
func GetProject(id int64, opts GetProjectOptions) *GetProjectResult {
	return GetProjectEx(ReadDB, id, opts)
}

// GetProjectEx gets a project with the given options.
//
// The returned value will be cached in the LRU cache if
// Project or Err is not nil.
func GetProjectEx(e boil.Executor, id int64, opts GetProjectOptions) (result *GetProjectResult) {
	opts.Preloads = sanitizeProjectRels(true, opts.Preloads...)

	cacheKey := makeCacheKey(opts)
	if c, err := ProjectCache.GetWithPrefix(id, cacheKey); err == nil {
		return c.(*GetProjectResult)
	}

	defer logger.Track()()

	result = &GetProjectResult{}
	defer func() {
		if result.Project != nil || result.Err != nil {
			ProjectCache.RemoveWithPrefix(id, cacheKey)
			ProjectCache.SetWithPrefix(id, cacheKey, result, time.Hour)
		}
	}()

	selectQueries := []QueryMod{Where("id = ?", id)}
	var loadStats bool
	for _, preload := range opts.Preloads {
		if strings.EqualFold(preload, ProjectRels.Statistic) {
			loadStats = true
		} else {
			selectQueries = append(selectQueries, Load(preload))
		}
	}

	if !opts.IncludesDrafts {
		selectQueries = append(selectQueries, Where("published_at IS NOT NULL"))
	}

	p, err := models.Projects(selectQueries...).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			result.Err = errs.ErrProjectNotFound
			return
		}
		logger.Err.Println(err)
		result.Err = errs.ErrUnknown
		return
	}

	result.Project = modext.NewProject(p).LoadRels(p)
	if loadStats {
		result.Project.Stats = GetProjectStats(p.ID).Stats
	}

	return
}

// GetPRojectMD gets project metadata from MangaDex.
func GetProjectMd(id string) (*ProjectDraft, error) {
	if strings.Contains(id, "/") {
		return nil, errors.New("Invalid chapter id")
	}

	mdGlobRateLimiter.Wait(context.Background())

	u := fmt.Sprintf("%s/manga/%s?includes[]=artist&includes[]=author&includes[]=cover_art", mdBaseURL, id)
	res, err := http.Get(u)
	if err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrProjectMdFetchFailed
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		if res.StatusCode == http.StatusNotFound {
			return nil, errs.ErrProjectNotFound
		}
		return nil, errs.ErrProjectMdFetchFailed
	}

	buf, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	body := &struct {
		Data struct {
			Attributes struct {
				Title struct {
					EN         string `json:"en"`
					JA         string `json:"ja"`
					JAromanize string `json:"ja-ro"`
					JP         string `json:"jp"`
				}
				Description struct {
					EN string `json:"en"`
				}
				Status                 string
				PublicationDemographic string
				ContentRating          string
				Tags                   []struct {
					Attributes struct {
						Name struct {
							EN string `json:"en"`
						}
					}
				}
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
	if err := json.Unmarshal(buf, &body); err != nil {
		return nil, err
	}

	draft := &ProjectDraft{
		Description:  body.Data.Attributes.Description.EN,
		SeriesStatus: strings.ToLower(body.Data.Attributes.Status),
		Demographic:  strings.ToLower(body.Data.Attributes.PublicationDemographic),
		Rating:       strings.ToLower(body.Data.Attributes.ContentRating),
	}

	if len(body.Data.Attributes.Title.EN) > 0 {
		draft.Title = body.Data.Attributes.Title.EN
	} else if len(body.Data.Attributes.Title.JA) > 0 {
		draft.Title = body.Data.Attributes.Title.JA
	} else if len(body.Data.Attributes.Title.JP) > 0 {
		draft.Title = body.Data.Attributes.Title.JP
	} else if len(body.Data.Attributes.Title.JAromanize) > 0 {
		draft.Title = body.Data.Attributes.Title.JAromanize
	}

	for _, tag := range body.Data.Attributes.Tags {
		if len(tag.Attributes.Name.EN) > 0 {
			draft.Tags = append(draft.Tags, tag.Attributes.Name.EN)
		}
	}

	for _, rel := range body.Data.Relationships {
		if rel.Type == "artist" {
			draft.Artists = append(draft.Artists, rel.Attributes.Name)
		} else if rel.Type == "author" {
			draft.Authors = append(draft.Authors, rel.Attributes.Name)
		} else if rel.Type == "cover_art" {
			draft.CoverURL = fmt.Sprintf("https://uploads.mangadex.org/covers/%s/%s", id, rel.Attributes.FileName)
		}
	}

	return draft, nil
}

// GetProjectsOptions represents the options for getting projects.
type GetProjectsOptions struct {
	Title                 string   `form:"title" json:"1,omitempty"`
	ProjectStatus         []string `form:"projectStatus" json:"2,omitempty"`
	SeriesStatus          []string `form:"seriesStatus" json:"3,omitempty"`
	Demographic           []string `form:"demographic" json:"4,omitempty"`
	Rating                []string `form:"rating" json:"5,omitempty"`
	ExcludedProjectStatus []string `form:"excludeProjectStatus" json:"6,omitempty"`
	ExcludedSeriesStatus  []string `form:"excludeSeriesStatus" json:"7,omitempty"`
	ExcludedDemographic   []string `form:"excludeDemographic" json:"8,omitempty"`
	ExcludedRating        []string `form:"excludeRating" json:"9,omitempty"`
	Artists               []string `form:"artist" json:"10,omitempty"`
	Authors               []string `form:"author" json:"11,omitempty"`
	Tags                  []string `form:"tag" json:"12,omitempty"`
	ExcludedTags          []string `form:"excludeTag" json:"13,omitempty"`
	Limit                 int      `form:"limit" json:"14,omitempty"`
	Offset                int      `form:"offset" json:"15,omitempty"`
	Preloads              []string `form:"preload" json:"16,omitempty"`
	Sort                  string   `form:"sort" json:"17,omitempty"`
	Order                 string   `form:"order" json:"18,omitempty"`
	IncludesDrafts        bool     `form:"includesDrafts" json:"19,omitempty"`
}

func (o *GetProjectsOptions) validate() error {
	o.Title = slug.Make(o.Title)

	for i, projectStatus := range o.ProjectStatus {
		projectStatus = strings.ToLower(projectStatus)
		o.ProjectStatus[i] = projectStatus

		if !stringsContains(ProjectStatus, projectStatus) {
			return errs.ErrInvalidProjectStatus
		}
	}
	sort.Strings(o.ProjectStatus)

	for i, seriesStatus := range o.SeriesStatus {
		seriesStatus = strings.ToLower(seriesStatus)
		o.SeriesStatus[i] = seriesStatus

		if !stringsContains(SeriesStatus, seriesStatus) {
			return errs.ErrInvalidSeriesStatus
		}
	}
	sort.Strings(o.SeriesStatus)

	for i, demographic := range o.Demographic {
		demographic = strings.ToLower(demographic)
		o.Demographic[i] = demographic

		if !stringsContains(Demographic, demographic) {
			return errs.ErrInvalidDemographic
		}
	}
	sort.Strings(o.Demographic)

	for i, rating := range o.Rating {
		rating = strings.ToLower(rating)
		o.Rating[i] = rating

		if !stringsContains(Rating, rating) {
			return errs.ErrInvalidRating
		}
	}
	sort.Strings(o.Rating)

	for i, projectStatus := range o.ExcludedProjectStatus {
		projectStatus = strings.ToLower(projectStatus)
		o.ExcludedProjectStatus[i] = projectStatus

		if !stringsContains(ProjectStatus, projectStatus) {
			return errs.ErrInvalidProjectStatus
		}
	}
	sort.Strings(o.ExcludedProjectStatus)

	for i, seriesStatus := range o.ExcludedSeriesStatus {
		seriesStatus = strings.ToLower(seriesStatus)
		o.ExcludedSeriesStatus[i] = seriesStatus

		if !stringsContains(SeriesStatus, seriesStatus) {
			return errs.ErrInvalidSeriesStatus
		}
	}
	sort.Strings(o.ExcludedSeriesStatus)

	for i, demographic := range o.ExcludedDemographic {
		demographic = strings.ToLower(demographic)
		o.ExcludedDemographic[i] = demographic

		if !stringsContains(Demographic, demographic) {
			return errs.ErrInvalidDemographic
		}
	}
	sort.Strings(o.ExcludedDemographic)

	for i, rating := range o.ExcludedRating {
		rating = strings.ToLower(rating)
		o.ExcludedRating[i] = rating

		if !stringsContains(Rating, rating) {
			return errs.ErrInvalidRating
		}
	}
	sort.Strings(o.ExcludedRating)

	for i, author := range o.Authors {
		o.Authors[i] = slug.Make(author)
	}

	for _, artist := range o.Artists {
		artist = slug.Make(artist)
		if !stringsContains(o.Authors, artist) {
			o.Authors = append(o.Authors, artist)
		}
	}

	o.Artists = nil
	sort.Strings(o.Authors)

	for i, tag := range o.Tags {
		o.Tags[i] = slug.Make(tag)
	}
	sort.Strings(o.Tags)

	for i, tag := range o.ExcludedTags {
		o.ExcludedTags[i] = slug.Make(tag)
	}
	sort.Strings(o.ExcludedTags)

	if o.Limit <= 0 {
		o.Limit = 20
	} else if o.Limit > 100 {
		o.Limit = 100
	}

	if o.Offset < 0 {
		o.Offset = 0
	}

	o.Preloads = sanitizeProjectRels(false, o.Preloads...)
	o.Sort = sanitizeProjectSort(o.Sort)
	o.Order = sanitizeOrder(o.Order)

	return nil
}

// GetProjectsResult represents the result of function GetProjects.
type GetProjectsResult struct {
	Projects []*modext.Project `json:"data"`
	Total    int64             `json:"total"`
	Err      error             `json:"error,omitempty"`
}

// This function simply calls GetProjectsEx with the global Read connection.
//
// The returned value will be cached in the LRU cache if
// Projects, Total or Err is not empty.
func GetProjects(opts GetProjectsOptions) *GetProjectsResult {
	return GetProjectsEx(ReadDB, opts)
}

// GetProjectsEx gets projects with the given options.
func GetProjectsEx(e boil.Executor, opts GetProjectsOptions) (result *GetProjectsResult) {
	if err := opts.validate(); err != nil {
		return &GetProjectsResult{Projects: []*modext.Project{}, Err: err}
	}

	prefix := "global"
	cacheKey := makeCacheKey(opts)
	if c, err := ProjectCache.GetWithPrefix(prefix, cacheKey); err == nil {
		return c.(*GetProjectsResult)
	}

	defer logger.Track()()

	result = &GetProjectsResult{Projects: []*modext.Project{}}
	defer func() {
		if len(result.Projects) > 0 || result.Total > 0 || result.Err != nil {
			ProjectCache.RemoveWithPrefix(prefix, cacheKey)
			ProjectCache.SetWithPrefix(prefix, cacheKey, result, time.Hour)
		}
	}()

	selectQueries := []QueryMod{GroupBy("project.id")}

	var queries []string
	var args []interface{}

	if len(opts.Title) > 0 {
		queries = append(queries, "project.slug ILIKE '%' || ? || '%'")
		args = append(args, opts.Title)
	}

	if len(opts.ProjectStatus) > 0 {
		var q []string
		for _, projectStatus := range opts.ProjectStatus {
			q = append(q, "project.project_status = ?")
			args = append(args, projectStatus)
		}
		queries = append(queries, fmt.Sprintf("(%s)", strings.Join(q, " OR ")))
	}

	if len(opts.SeriesStatus) > 0 {
		var q []string
		for _, seriesStatus := range opts.SeriesStatus {
			q = append(q, "project.series_status = ?")
			args = append(args, seriesStatus)
		}
		queries = append(queries, fmt.Sprintf("(%s)", strings.Join(q, " OR ")))
	}

	if len(opts.Demographic) > 0 {
		var q []string
		for _, demographic := range opts.Demographic {
			q = append(q, "project.demographic = ?")
			args = append(args, demographic)
		}
		queries = append(queries, fmt.Sprintf("(%s)", strings.Join(q, " OR ")))
	}

	if len(opts.Rating) > 0 {
		var q []string
		for _, rating := range opts.Rating {
			q = append(q, "project.rating = ?")
			args = append(args, rating)
		}
		queries = append(queries, fmt.Sprintf("(%s)", strings.Join(q, " OR ")))
	}

	if len(opts.ExcludedProjectStatus) > 0 {
		var q []string
		for _, projectStatus := range opts.ExcludedProjectStatus {
			q = append(q, "project.project_status != ?")
			args = append(args, projectStatus)
		}
		queries = append(queries, fmt.Sprintf("(%s)", strings.Join(q, " AND ")))
	}

	if len(opts.ExcludedSeriesStatus) > 0 {
		var q []string
		for _, seriesStatus := range opts.ExcludedSeriesStatus {
			q = append(q, "project.series_status != ?")
			args = append(args, seriesStatus)
		}
		queries = append(queries, fmt.Sprintf("(%s)", strings.Join(q, " AND ")))
	}

	if len(opts.ExcludedDemographic) > 0 {
		var q []string
		for _, demographic := range opts.ExcludedDemographic {
			q = append(q, "project.demographic != ?")
			args = append(args, demographic)
		}
		queries = append(queries, fmt.Sprintf("(%s)", strings.Join(q, " AND ")))
	}

	if len(opts.ExcludedRating) > 0 {
		var q []string
		for _, rating := range opts.ExcludedRating {
			q = append(q, "project.rating != ?")
			args = append(args, rating)
		}
		queries = append(queries, fmt.Sprintf("(%s)", strings.Join(q, " AND ")))
	}

	if len(opts.Authors) > 0 {
		selectQueries = append(selectQueries,
			InnerJoin("project_artists par ON par.project_id = project.id"),
			InnerJoin("project_authors pau ON pau.project_id = project.id"),
			InnerJoin("author a ON a.id IN (par.artist_id, pau.author_id)"),
		)

		var q []string
		for _, author := range opts.Authors {
			q = append(q, "a.slug ILIKE '%' || ? || '%'")
			args = append(args, author)
		}
		queries = append(queries, fmt.Sprintf("(%s)", strings.Join(q, " OR ")))
	}

	if len(opts.Tags) > 0 || len(opts.ExcludedTags) > 0 {
		selectQueries = append(selectQueries,
			InnerJoin("project_tags pt ON pt.project_id = project.id"),
			InnerJoin("tag t ON t.id = pt.tag_id"),
		)
	}

	if len(opts.Tags) > 0 {
		var q []string
		for _, tag := range opts.Tags {
			q = append(q, "t.slug = ?")
			args = append(args, tag)
		}
		queries = append(queries, fmt.Sprintf("(%s)", strings.Join(q, " OR ")))
	}

	if len(opts.ExcludedTags) > 0 {
		var q []string
		for _, tag := range opts.ExcludedTags {
			q = append(q, "t.slug != ?")
			args = append(args, tag)
		}
		queries = append(queries, fmt.Sprintf("(%s)", strings.Join(q, " AND ")))
	}

	if !opts.IncludesDrafts {
		queries = append(queries, "project.published_at IS NOT NULL")
	}

	if len(queries) > 0 {
		selectQueries = append(selectQueries, Where(strings.Join(queries, " AND "), args...))
	}

	countQueries := append([]QueryMod{Select("1")}, selectQueries...)

	selectQueries = append(selectQueries, Limit(opts.Limit), Offset(opts.Offset))
	selectQueries = append(selectQueries, OrderBy(fmt.Sprintf("%s %s", opts.Sort, opts.Order)))

	for _, rel := range opts.Preloads {
		selectQueries = append(selectQueries, Load(rel))
	}

	projects, err := models.Projects(selectQueries...).All(e)
	if err != nil {
		logger.Err.Println(err)
		result.Err = errs.ErrUnknown
		return
	}

	count, err := models.Projects(countQueries...).All(e)
	if err != nil {
		logger.Err.Println(err)
		result.Err = errs.ErrUnknown
		return
	}

	result.Projects = make([]*modext.Project, len(projects))
	result.Total = int64(len(count))

	for i, p := range projects {
		result.Projects[i] = modext.NewProject(p).LoadRels(p)
	}
	return
}

// This function simply creates a new write transaction and calls UpdateProjectEx.
func UpdateProject(id int64, draft *ProjectDraft) (*modext.Project, error) {
	tx, err := WriteDB.Begin()
	if err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}
	return UpdateProjectEx(tx, id, draft)
}

// UpdateProjectEx updates a project.
// Returns the updated project if successful.
func UpdateProjectEx(tx *sql.Tx, id int64, draft *ProjectDraft) (*modext.Project, error) {
	defer logger.Track()()

	if err := draft.validate(); err != nil {
		return nil, err
	}

	p, err := models.FindProject(tx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrProjectNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	if p.Locked.Bool {
		return nil, errs.ErrProjectLocked
	}

	prevTitle := p.Title

	p.Title = draft.Title
	p.Description = null.StringFrom(draft.Description)
	p.ProjectStatus = draft.ProjectStatus
	p.SeriesStatus = draft.SeriesStatus
	p.Demographic = null.StringFrom(draft.Demographic)
	p.Rating = null.StringFrom(draft.Rating)

	if err := p.Update(tx, boil.Whitelist(
		ProjectCols.Title,
		ProjectCols.Slug,
		ProjectCols.Description,
		ProjectCols.ProjectStatus,
		ProjectCols.SeriesStatus,
		ProjectCols.Demographic,
		ProjectCols.Rating,
		ProjectCols.UpdatedAt,
	)); err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	if err := refreshProjectRels(tx, p, draft); err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	if err := tx.Commit(); err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	if prevTitle != p.Title {
		go renameProjectDir(p)
	}

	go projectAfterUpdateHook(p)
	return modext.NewProject(p).LoadRels(p), nil
}

// This function simply calls PublishProjectEx with the global Write connection.
func PublishProject(id int64) (*modext.Project, error) {
	return PublishProjectEx(WriteDB, id)
}

// PublishProjectEx publishes a project.
// Returns the updated project if successful.
func PublishProjectEx(e boil.Executor, id int64) (*modext.Project, error) {
	defer logger.Track()()

	p, err := models.FindProject(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrProjectNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	if p.Locked.Bool {
		return nil, errs.ErrProjectLocked
	}

	updatedAt := p.UpdatedAt
	p.PublishedAt = null.TimeFrom(time.Now().UTC())

	if err := p.Update(e, boil.Whitelist(ProjectCols.PublishedAt)); err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	p.UpdatedAt = updatedAt
	go func() {
		projectAfterUpdateHook(p)
		projectAfterPublishStateUpdateHook(p)
	}()
	return modext.NewProject(p), nil
}

// This function simply calls UnpublishProjectEx with the global Write connection.
func UnpublishProject(id int64) (*modext.Project, error) {
	return UnpublishProjectEx(WriteDB, id)
}

// UnpublishProjectEx unpublishes a project.
// Returns the updated project if successful.
func UnpublishProjectEx(e boil.Executor, id int64) (*modext.Project, error) {
	defer logger.Track()()

	p, err := models.FindProject(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrProjectNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	if p.Locked.Bool {
		return nil, errs.ErrProjectLocked
	}

	updatedAt := p.UpdatedAt
	p.PublishedAt.Valid = false

	if err := p.Update(e, boil.Whitelist(ProjectCols.PublishedAt)); err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	p.UpdatedAt = updatedAt
	go func() {
		projectAfterUpdateHook(p)
		projectAfterPublishStateUpdateHook(p)
	}()
	return modext.NewProject(p), nil
}

// This function simply calls LockProjectEx with the global Write connection.
func LockProject(id int64) (*modext.Project, error) {
	return LockProjectEx(WriteDB, id)
}

// LockProjectEx locks a project.
// Returns the updated project if successful.
func LockProjectEx(e boil.Executor, id int64) (*modext.Project, error) {
	defer logger.Track()()

	p, err := models.FindProject(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrProjectNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	updatedAt := p.UpdatedAt
	p.Locked = null.BoolFrom(true)

	if err := p.Update(e, boil.Whitelist(ProjectCols.Locked)); err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	p.UpdatedAt = updatedAt
	go projectAfterUpdateHook(p)
	return modext.NewProject(p), nil
}

// This function simply calls UnlockProjectEx with the global Write connection.
func UnlockProject(id int64) (*modext.Project, error) {
	return UnlockProjectEx(WriteDB, id)
}

// UnlockProjectEx unlocks a project.
// Returns the updated project if successful.
func UnlockProjectEx(e boil.Executor, id int64) (*modext.Project, error) {
	defer logger.Track()()

	p, err := models.FindProject(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errs.ErrProjectNotFound
		}
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	updatedAt := p.UpdatedAt
	p.Locked = null.NewBool(false, false)

	if err := p.Update(e, boil.Whitelist(ProjectCols.Locked)); err != nil {
		logger.Err.Println(err)
		return nil, errs.ErrUnknown
	}

	p.UpdatedAt = updatedAt
	go projectAfterUpdateHook(p)
	return modext.NewProject(p), nil
}

// This function simply calls DeleteProjectEx with the global Write connection.
func DeleteProject(id int64) error {
	return DeleteProjectEx(WriteDB, id)
}

// DeleteProjectEx deletes a project.
func DeleteProjectEx(e boil.Executor, id int64) error {
	defer logger.Track()()

	p, err := models.FindProject(e, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return errs.ErrProjectNotFound
		}
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	if p.Locked.Bool {
		return errs.ErrProjectLocked
	}

	if err := p.Delete(e); err != nil {
		logger.Err.Println(err)
		return errs.ErrUnknown
	}

	ProjectCache.PurgeWithPrefix(p.ID)
	CoverCache.PurgeWithPrefix(p.ID)
	ChapterCache.PurgeWithPrefix(p.ID)

	go refreshTemplatesCache()
	go removeProjectDir(p)
	go func() {
		refreshProjectsCache()
		refreshChaptersCache()
	}()

	return nil
}

// This function simply calls CheckProjectExistsEx with the global Read connection.
func CheckProjectExists(id int64) (int64, string) {
	return CheckProjectExistsEx(ReadDB, id)
}

// CheckProjectExistsEx checks if a project exists and returns
// id and slug of the project if successful.
func CheckProjectExistsEx(e boil.Executor, id int64) (int64, string) {
	defer logger.Track()()

	p, err := models.FindProject(e, id)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Err.Println(err)
		}
		return 0, ""
	}

	return p.ID, p.Slug
}

// This function simply calls CheckProjectExistsBySlugEx with the global Read connection.
func CheckProjectExistsBySlug(slug string) (int64, string) {
	return CheckProjectExistsBySlugEx(ReadDB, slug)
}

// CheckProjectExistsBySlugEx checks if a project exists by slug and
// returns id and slug of the project if successful.
func CheckProjectExistsBySlugEx(e boil.Executor, slug string) (int64, string) {
	defer logger.Track()()

	p, err := models.Projects(Where("slug ILIKE ?", slug)).One(e)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Err.Println(err)
		}
		return 0, ""
	}

	return p.ID, p.Slug
}

// This function simply calls CheckProjectExistsByTitleEx with the global Read connection.
func CheckProjectExistsByTitle(title string) (int64, string) {
	return CheckProjectExistsByTitleEx(ReadDB, title)
}

// CheckProjectExistsByTitleEx checks if a project exists by title and
// returns id and slug of the project if successful.
func CheckProjectExistsByTitleEx(e boil.Executor, title string) (int64, string) {
	defer logger.Track()()

	p, err := models.Projects(Where("title ILIKE ?", title)).One(e)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Err.Println(err)
		}
		return 0, ""
	}

	return p.ID, p.Slug
}

// This function simply calls CheckProjectExistsBySlugOrTitleEx with the global Read connection.
func CheckProjectExistsBySlugOrTitle(slugOrTitle string) (int64, string) {
	return CheckProjectExistsBySlugOrTitleEx(ReadDB, slugOrTitle)
}

// CheckProjectExistsBySlugOrTitleEx checks if a project exists by slug or title and
// returns id and slug of the project if successful.
func CheckProjectExistsBySlugOrTitleEx(e boil.Executor, slugOrTitle string) (int64, string) {
	defer logger.Track()()

	p, err := models.Projects(Where("slug ILIKE ? OR title ILIKE ?", slugOrTitle, slugOrTitle)).One(e)
	if err != nil {
		if err != sql.ErrNoRows {
			logger.Err.Println(err)
		}
		return 0, ""
	}

	return p.ID, p.Slug
}
