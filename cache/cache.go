package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"kasen/config"

	"github.com/bluele/gcache"
	"github.com/go-redis/redis/v8"
)

var (
	Redis          *redis.Client
	ProjectCache   *LRU
	CoverCache     *LRU
	ChapterCache   *LRU
	PagesCache     *LRU
	StatsCache     *LRU
	TemplatesCache *LRU
)

func init() {
	redisConfig := config.GetRedis()
	cacheConfig := config.GetCache()

	Redis = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", redisConfig.Host, redisConfig.Port),
		DB:       redisConfig.DB,
		Password: redisConfig.Passwd,
	})

	if _, err := Redis.Ping(context.Background()).Result(); err != nil {
		log.Fatalln(err)
	}

	ProjectCache = &LRU{gcache.New(512).LRU().Expiration(time.Duration(cacheConfig.DefaultTTL)).Build()}
	ChapterCache = &LRU{gcache.New(1024).LRU().Expiration(cacheConfig.DefaultTTL).Build()}
	CoverCache = &LRU{gcache.New(1024).LRU().Expiration(cacheConfig.DefaultTTL).Build()}
	PagesCache = &LRU{gcache.New(128).LRU().Expiration(cacheConfig.DefaultTTL).Build()}
	StatsCache = &LRU{gcache.New(4096).LRU().Expiration(cacheConfig.DefaultTTL).Build()}
	TemplatesCache = &LRU{gcache.New(512).LRU().Expiration(5 * time.Minute).Build()}
}
