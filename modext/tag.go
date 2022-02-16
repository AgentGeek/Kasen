package modext

import "kasen/models"

type Tag struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`

	Projects []*Project `json:"-"`
}

func NewTag(tag *models.Tag) *Tag {
	if tag == nil {
		return nil
	}
	return &Tag{
		ID:   tag.ID,
		Slug: tag.Slug,
		Name: tag.Name,
	}
}

func (t *Tag) LoadProjects(tag *models.Tag) *Tag {
	if tag == nil || tag.R == nil || len(tag.R.Projects) == 0 {
		return t
	}

	t.Projects = make([]*Project, len(tag.R.Projects))
	for i, project := range tag.R.Projects {
		t.Projects[i] = NewProject(project)
	}

	return t
}
