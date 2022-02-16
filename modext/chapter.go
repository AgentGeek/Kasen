package modext

import (
	"fmt"
	"math"
	"sort"

	"kasen/models"
)

type Chapter struct {
	ID          int64    `json:"id"`
	Locked      bool     `json:"locked,omitempty"`
	CreatedAt   int64    `json:"createdAt"`
	UpdatedAt   int64    `json:"updatedAt"`
	PublishedAt int64    `json:"publishedAt,omitempty"`
	Chapter     string   `json:"chapter"`
	Volume      string   `json:"volume,omitempty"`
	Title       string   `json:"title,omitempty"`
	Pages       []string `json:"pages,omitempty"`

	Project          *Project           `json:"project,omitempty"`
	Uploader         *User              `json:"uploader,omitempty"`
	ScanlationGroups []*ScanlationGroup `json:"scanlationGroups,omitempty"`
	Stats            *ChapterStats      `json:"stats,omitempty"`

	Thumbnail string `json:"-"`
}

func NewChapter(chapter *models.Chapter) *Chapter {
	if chapter == nil {
		return nil
	}

	c := &Chapter{
		ID:        chapter.ID,
		Locked:    chapter.Locked.Bool,
		CreatedAt: chapter.CreatedAt.Unix(),
		UpdatedAt: chapter.UpdatedAt.Unix(),
		Chapter:   chapter.Chapter,
		Volume:    chapter.Volume.String,
		Title:     chapter.Title.String,
	}

	if chapter.PublishedAt.Valid {
		c.PublishedAt = chapter.PublishedAt.Time.Unix()
	}

	return c
}

// LoadRels loads all relations of the chapter.
func (c *Chapter) LoadRels(chapter *models.Chapter) *Chapter {
	if chapter == nil || chapter.R == nil {
		return c
	}

	c.LoadProject(chapter)
	c.LoadUploader(chapter)
	c.LoadScanlationGroups(chapter)
	c.LoadStats(chapter)

	return c
}

func (c *Chapter) LoadPages(chapter *models.Chapter) *Chapter {
	if chapter == nil {
		return c
	}

	c.Pages = chapter.Pages
	return c
}

func (c *Chapter) LoadProject(chapter *models.Chapter) *Chapter {
	if chapter == nil || chapter.R == nil || chapter.R.Project == nil {
		return c
	}
	c.Project = NewProject(chapter.R.Project)
	return c
}

func (c *Chapter) LoadUploader(chapter *models.Chapter) *Chapter {
	if chapter == nil || chapter.R == nil || chapter.R.Uploader == nil {
		return c
	}
	c.Uploader = NewUser(chapter.R.Uploader)
	return c
}

func (c *Chapter) LoadScanlationGroups(chapter *models.Chapter) *Chapter {
	if chapter == nil || chapter.R == nil || len(chapter.R.ScanlationGroups) == 0 {
		return c
	}

	c.ScanlationGroups = make([]*ScanlationGroup, len(chapter.R.ScanlationGroups))
	for i, group := range chapter.R.ScanlationGroups {
		c.ScanlationGroups[i] = NewScanlationGroup(group)
	}
	sort.Slice(c.ScanlationGroups, func(i, j int) bool {
		return c.ScanlationGroups[i].Name < c.ScanlationGroups[j].Name
	})

	return c
}

func (c *Chapter) LoadStats(chapter *models.Chapter) *Chapter {
	if chapter == nil || chapter.R == nil {
		return c
	}
	c.Stats = NewChapterStats(chapter.R.Statistic)
	return c
}

func (c *Chapter) GetThumbnail(chapter *models.Chapter) {
	pages := chapter.Pages

	if len(pages) <= 1 {
		if c.Project != nil && c.Project.Cover != nil {
			c.Thumbnail = c.Project.Cover.Path(c.Project)
			return
		}
		return
	}

	min := math.Max(1, math.Ceil(float64(len(pages))/4))
	max := math.Min(float64(len(pages)), min)
	i := int(max)

	c.Thumbnail = fmt.Sprintf("/pages/%d/%s", c.ID, pages[i])
}
