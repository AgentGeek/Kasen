package modext

import "kasen/models"

type Project struct {
	ID            int64  `json:"id"`
	Slug          string `json:"slug"`
	Locked        bool   `json:"locked,omitempty"`
	CreatedAt     int64  `json:"createdAt"`
	UpdatedAt     int64  `json:"updatedAt"`
	PublishedAt   int64  `json:"publishedAt,omitempty"`
	Title         string `json:"title"`
	Description   string `json:"description,omitempty"`
	ProjectStatus string `json:"projectStatus"`
	SeriesStatus  string `json:"seriesStatus"`
	Demographic   string `json:"demographic,omitempty"`
	Rating        string `json:"rating,omitempty"`

	Artists  []*Author     `json:"artists,omitempty"`
	Authors  []*Author     `json:"authors,omitempty"`
	Tags     []*Tag        `json:"tags,omitempty"`
	Cover    *Cover        `json:"cover,omitempty"`
	Covers   []*Cover      `json:"-"`
	Chapters []*Chapter    `json:"-"`
	Stats    *ProjectStats `json:"stats,omitempty"`
}

func NewProject(project *models.Project) *Project {
	if project == nil {
		return nil
	}

	p := &Project{
		ID:            project.ID,
		Slug:          project.Slug,
		Locked:        project.Locked.Bool,
		CreatedAt:     project.CreatedAt.Unix(),
		UpdatedAt:     project.UpdatedAt.Unix(),
		Title:         project.Title,
		Description:   project.Description.String,
		ProjectStatus: project.ProjectStatus,
		SeriesStatus:  project.SeriesStatus,
		Demographic:   project.Demographic.String,
		Rating:        project.Rating.String,
	}

	if project.PublishedAt.Valid {
		p.PublishedAt = project.PublishedAt.Time.Unix()
	}

	return p
}

func (p *Project) LoadRels(project *models.Project) *Project {
	if project == nil || project.R == nil {
		return p
	}

	p.LoadArtists(project)
	p.LoadAuthors(project)
	p.LoadTags(project)
	p.LoadCover(project)
	p.LoadCovers(project)
	p.LoadChapters(project)
	p.LoadStats(project)

	return p
}

func (p *Project) LoadArtists(project *models.Project) *Project {
	if project == nil || project.R == nil || len(project.R.Artists) == 0 {
		return p
	}

	p.Artists = make([]*Author, len(project.R.Artists))
	for i, artist := range project.R.Artists {
		p.Artists[i] = NewAuthor(artist)
	}

	return p
}

func (p *Project) LoadAuthors(project *models.Project) *Project {
	if project == nil || project.R == nil || len(project.R.Authors) == 0 {
		return p
	}

	p.Authors = make([]*Author, len(project.R.Authors))
	for i, author := range project.R.Authors {
		p.Authors[i] = NewAuthor(author)
	}

	return p
}

func (p *Project) LoadTags(project *models.Project) *Project {
	if project == nil || project.R == nil || len(project.R.Tags) == 0 {
		return p
	}

	p.Tags = make([]*Tag, len(project.R.Tags))
	for i, tag := range project.R.Tags {
		p.Tags[i] = NewTag(tag)
	}

	return p
}

func (p *Project) LoadCover(project *models.Project) *Project {
	if project == nil || project.R == nil || project.R.Cover == nil {
		return p
	}

	p.Cover = NewCover(project.R.Cover)

	return p
}

func (p *Project) LoadCovers(project *models.Project) *Project {
	if project == nil || project.R == nil || len(project.R.Covers) == 0 {
		return p
	}

	p.Covers = make([]*Cover, len(project.R.Covers))
	for i, cover := range project.R.Covers {
		p.Covers[i] = NewCover(cover)
	}

	return p
}

func (p *Project) LoadChapters(project *models.Project) *Project {
	if project == nil || project.R == nil || len(project.R.Chapters) == 0 {
		return p
	}

	p.Chapters = make([]*Chapter, len(project.R.Chapters))
	for i, chapter := range project.R.Chapters {
		p.Chapters[i] = NewChapter(chapter)
	}

	return p
}

func (p *Project) LoadStats(project *models.Project) *Project {
	if project == nil || project.R == nil || project.R.Statistic == nil {
		return p
	}

	p.Stats = NewProjectStats(project.R.Statistic)

	return p
}
