package services

import (
	"kasen/models"

	"github.com/gosimple/slug"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

func init() {
	nameToSlug := func(e boil.Executor, t *models.Tag) error {
		if len(t.Name) > 0 {
			t.Slug = slug.Make(t.Name)
		}
		return nil
	}

	models.AddTagHook(boil.BeforeInsertHook, nameToSlug)
	models.AddTagHook(boil.BeforeUpdateHook, nameToSlug)
	models.AddTagHook(boil.BeforeUpsertHook, nameToSlug)
}
