package controllers

import (
	"net/http"

	"kasen/modext"
	"kasen/server"
	"kasen/services"
)

func Home(c *server.Context) {
	templateName := "home.html"
	if c.TryCache(templateName) {
		return
	}

	pResult := services.GetProjects(services.GetProjectsOptions{
		Preloads: []string{
			services.ProjectRels.Cover,
			services.ProjectRels.Tags,
		},
		Limit: 12,
	})

	var projects []*modext.Project
	if len(pResult.Projects) > 0 {
		projects = append(projects, pResult.Projects...)
		if len(projects) > 6 {
			if remainder := len(projects) % 6; remainder > 0 {
				projects = projects[:len(projects)-remainder]
			}
		}
	}
	c.SetData("projects", projects)

	cResult := services.GetChapters(services.GetChaptersOptions{
		Limit: 12,
		Preloads: []string{
			services.ChapterRels.Uploader,
			services.ChapterRels.ScanlationGroups,
		},

		GetThumbnail: true,
		GetTags:      true,
	})
	var chapters []*modext.Chapter
	if len(cResult.Chapters) > 0 {
		chapters = append(chapters, cResult.Chapters...)
		if len(chapters) > 3 {
			if remainder := len(chapters) % 3; remainder > 0 {
				chapters = chapters[:len(chapters)-remainder]
			}
		}
	}
	c.SetData("chapters", chapters)
	c.Cache(http.StatusOK, templateName)
}
