package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	. "kasen/cache"
	. "kasen/database"

	"kasen/errs"
	"kasen/models"
	"kasen/modext"

	"github.com/go-redis/redis/v8"
	"github.com/rs1703/logger"
	"github.com/volatiletech/sqlboiler/v4/boil"
	. "github.com/volatiletech/sqlboiler/v4/queries/qm"
)

// GetStatsCacheStats gets the cache stats of the statistics LRU cache.
func GetStatsCacheStats() *CacheStats {
	return StatsCache.GetStats()
}

// GetProjectStatsResult represents the result of GetProjectStats.
type GetProjectStatsResult struct {
	Stats *modext.ProjectStats `json:"data,omitempty"`
	Err   error                `json:"error,omitempty"`
}

// This function simply calls GetProjectStatsEx with the global Read connection.
func GetProjectStats(id int64) *GetProjectStatsResult {
	return GetProjectStatsEx(ReadDB, id)
}

// GetProjectStatsEx returns the statistics of the given project.
//
// The returned value will be cached in the LRU cache if
// Stats or Err is not nil.
func GetProjectStatsEx(e boil.Executor, id int64) (result *GetProjectStatsResult) {
	if c, err := StatsCache.GetWithPrefix("p", id); err == nil {
		return c.(*GetProjectStatsResult)
	}

	defer logger.Track()()

	result = &GetProjectStatsResult{}
	defer func() {
		if result.Stats != nil || result.Err != nil {
			StatsCache.RemoveWithPrefix("p", id)
			StatsCache.SetWithPrefix("p", id, result, 0)
		}
	}()

	stats, err := models.Statistics(Where("project_id = ?", id)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			result.Err = errs.ErrChapterNotFound
			return
		}
		logger.Err.Println(err)
		result.Err = err
		return
	}

	result.Stats = modext.NewProjectStats(stats)
	return
}

// GetProjectStatsResult represents the result of GetChapterStatsEx.
type GetChapterStatsResult struct {
	Stats *modext.ChapterStats `json:"data,omitempty"`
	Err   error                `json:"error,omitempty"`
}

// This function simply calls GetChapterStatsEx with the global Read connection.
func GetChapterStats(id int64) *GetChapterStatsResult {
	return GetChapterStatsEx(ReadDB, id)
}

// GetChapterStatsEx returns the statistics of the given project.
//
// The returned value will be cached in the LRU cache if
// Stats or Err is not nil.
func GetChapterStatsEx(e boil.Executor, id int64) (result *GetChapterStatsResult) {
	if c, err := StatsCache.GetWithPrefix("c", id); err == nil {
		return c.(*GetChapterStatsResult)
	}

	defer logger.Track()()

	result = &GetChapterStatsResult{}
	defer func() {
		if result.Stats != nil || result.Err != nil {
			StatsCache.RemoveWithPrefix("c", id)
			StatsCache.SetWithPrefix("c", id, result, 0)
		}
	}()

	stats, err := models.Statistics(Where("chapter_id = ?", id)).One(e)
	if err != nil {
		if err == sql.ErrNoRows {
			result.Err = errs.ErrChapterNotFound
			return
		}
		logger.Err.Println(err)
		result.Err = err
		return
	}

	result.Stats = modext.NewChapterStats(stats)
	return
}

// IncreaseViewCount increases the view count and unique count of the given
// project or chapter.
//
// It will update the view count on a per-minute basis,
// And it will only update the unique view count if the ip address is unique.
func IncreaseViewCount(s modext.StatisticsOwnerUpdater, ip string) error {
	var vInc, uvInc int

	identifier := s.CacheIdentifier()
	pkName := s.PrimaryKeyName()
	pk := s.PrimaryKey()

	vk := fmt.Sprintf("%s%dv%s", identifier, pk, ip)
	if _, err := Redis.Get(context.Background(), vk).Result(); err == redis.Nil {
		vInc++
		s.IncreaseViewCount()
	}

	uvk := fmt.Sprintf("%s%duv%s", identifier, pk, ip)
	if _, err := Redis.Get(context.Background(), uvk).Result(); err == redis.Nil {
		uvInc++
		s.IncreaseUniqueViewCount()
	}

	if vInc > 0 || uvInc > 0 {
		if _, err := WriteDB.Exec(
			fmt.Sprintf(
				`UPDATE statistics
					SET view_count = view_count + $1,	unique_view_count = unique_view_count + $2
				WHERE %s = $3`, pkName), vInc, uvInc, pk); err != nil {
			return err
		}

		if vInc > 0 {
			Redis.Set(context.Background(), vk, 1, 1*time.Minute)
		}
		if uvInc > 0 {
			Redis.Set(context.Background(), uvk, 1, 0)
		}

	}
	return nil
}
