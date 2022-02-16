package services

import (
	"kasen/models"

	"github.com/gosimple/slug"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func init() {
	// makeSlug transforms title to slug before insert, update or upsert
	makeSlug := func(e boil.Executor, a *models.ScanlationGroup) error {
		if a != nil && len(a.Name) > 0 {
			a.Slug = slug.Make(a.Name)
		}
		return nil
	}

	models.AddScanlationGroupHook(boil.BeforeInsertHook, makeSlug)
	models.AddScanlationGroupHook(boil.BeforeUpdateHook, makeSlug)
	models.AddScanlationGroupHook(boil.BeforeUpsertHook, makeSlug)
}
