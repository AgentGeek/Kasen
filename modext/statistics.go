package modext

import (
	"sync"

	"kasen/models"
)

type Statistics struct {
	ID              int64 `json:"-"`
	ViewCount       int64 `json:"viewCount"`
	UniqueViewCount int64 `json:"-"`
	mutex           sync.Mutex
}

type StatisticsOwner interface {
	CacheIdentifier() string
	PrimaryKeyName() string
	PrimaryKey() int64
}

type StatisticsUpdater interface {
	IncreaseViewCount()
	IncreaseUniqueViewCount()
}

type StatisticsOwnerUpdater interface {
	StatisticsOwner
	StatisticsUpdater
}

func NewStatistics(statistics *models.Statistic) *Statistics {
	if statistics == nil {
		return nil
	}
	return &Statistics{
		ID:              statistics.ID,
		ViewCount:       statistics.ViewCount,
		UniqueViewCount: statistics.UniqueViewCount,
	}
}

func (stats *Statistics) GetMutex() *sync.Mutex {
	return &stats.mutex
}

func (stats *Statistics) IncreaseViewCount() {
	stats.mutex.Lock()
	stats.ViewCount++
	stats.mutex.Unlock()
}

func (stats *Statistics) IncreaseUniqueViewCount() {
	stats.mutex.Lock()
	stats.UniqueViewCount++
	stats.mutex.Unlock()
}
