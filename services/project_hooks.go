package services

import (
	"kasen/models"

	"github.com/gosimple/slug"
	"github.com/volatiletech/null/v8"
	. "github.com/volatiletech/sqlboiler/v4/boil"
)

func projectAfterUpdateHook(p *models.Project) {
	go refreshTemplatesCache()

	refreshProjectCache(p.ID)
	refreshProjectsCache()
}

func projectAfterPublishStateUpdateHook(p *models.Project) {
	refreshProjectChaptersCache(p.ID)
	refreshChaptersCache()
}

func init() {
	// makeSlug transforms title to slug before insert, update or upsert
	makeSlug := func(e Executor, p *models.Project) error {
		if p != nil && len(p.Title) > 0 {
			p.Slug = slug.Make(p.Title)
		}
		return nil
	}

	models.AddProjectHook(BeforeInsertHook, makeSlug)
	models.AddProjectHook(BeforeUpdateHook, makeSlug)
	models.AddProjectHook(BeforeUpsertHook, makeSlug)

	// Create statistics after insert
	models.AddProjectHook(AfterInsertHook, func(e Executor, p *models.Project) error {
		if p.ID <= 0 {
			return nil
		}
		s := models.Statistic{ProjectID: null.Int64From(p.ID)}
		return s.Insert(e, Infer())
	})
}
