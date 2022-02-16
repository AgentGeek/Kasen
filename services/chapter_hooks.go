package services

import (
	"kasen/models"

	"github.com/volatiletech/null/v8"
	. "github.com/volatiletech/sqlboiler/v4/boil"
)

func chapterAfterUpdateHook(c *models.Chapter) {
	go refreshTemplatesCache()

	refreshProjectChaptersCache(c.ProjectID)
	refreshChapterCache(c.ID)
	refreshChaptersCache()
}

func init() {
	// Create statistics after insert
	models.AddChapterHook(AfterInsertHook, func(e Executor, c *models.Chapter) error {
		if c.ID <= 0 {
			return nil
		}
		s := models.Statistic{ChapterID: null.Int64From(c.ID)}
		return s.Insert(e, Infer())
	})
}
