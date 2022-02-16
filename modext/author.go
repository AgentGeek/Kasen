package modext

import "kasen/models"

type Author struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
	Name string `json:"name"`

	Projects []*Project `json:"-"`
}

func NewAuthor(author *models.Author) *Author {
	if author == nil {
		return nil
	}
	return &Author{
		ID:   author.ID,
		Slug: author.Slug,
		Name: author.Name,
	}
}

func (a *Author) LoadProjects(author *models.Author) *Author {
	if author == nil || author.R == nil {
		return a
	}

	if len(author.R.ArtistProjects) > 0 {
		a.Projects = make([]*Project, len(author.R.ArtistProjects))
		for i, project := range author.R.ArtistProjects {
			exists := false
			for _, p := range a.Projects {
				if p.ID == project.ID {
					exists = true
					break
				}
			}
			if !exists {
				a.Projects[i] = NewProject(project)
			}
		}
	}

	if len(author.R.AuthorProjects) > 0 {
		a.Projects = make([]*Project, len(author.R.AuthorProjects))
		for i, project := range author.R.AuthorProjects {
			exists := false
			for _, p := range a.Projects {
				if p.ID == project.ID {
					exists = true
					break
				}
			}
			if !exists {
				a.Projects[i] = NewProject(project)
			}
		}
	}

	return a
}
