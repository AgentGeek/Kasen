package modext

import "kasen/models"

type ChapterStats struct {
	*Statistics
	ChapterID int64 `json:"-"`
}

func NewChapterStats(statistics *models.Statistic) *ChapterStats {
	if statistics == nil {
		return nil
	}
	return &ChapterStats{
		Statistics: NewStatistics(statistics),
		ChapterID:  statistics.ChapterID.Int64,
	}
}

func (ChapterStats) CacheIdentifier() string {
	return "c"
}

func (ChapterStats) PrimaryKeyName() string {
	return "chapter_id"
}

func (c *ChapterStats) PrimaryKey() int64 {
	return c.ChapterID
}
