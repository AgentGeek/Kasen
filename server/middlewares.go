package server

import (
	"fmt"
	"net/http"

	"kasen/cache"

	"github.com/rs1703/logger"
	"github.com/ulule/limiter/v3"
	mgin "github.com/ulule/limiter/v3/drivers/middleware/gin"
	"github.com/ulule/limiter/v3/drivers/store/redis"
)

func WithName(name string) Handler {
	return func(c *Context) {
		c.SetData("name", name)
		c.Next()
	}
}

func WithRedirect(relativePath string) Handler {
	return func(c *Context) {
		c.Redirect(http.StatusFound, relativePath)
	}
}

func WithAuthorization(onError Handler) Handler {
	return func(c *Context) {
		if c.GetUser() == nil {
			if onError == nil {
				c.AbortWithStatus(http.StatusForbidden)
			} else {
				onError(c)
			}
			c.Abort()
			return
		}
		c.Next()
	}
}

func WithNoAuthorization(onError Handler) Handler {
	return func(c *Context) {
		if c.GetUser() != nil {
			if onError == nil {
				c.AbortWithStatus(http.StatusForbidden)
			} else {
				onError(c)
			}
			c.Abort()
			return
		}
		c.Next()
	}
}

func WithPermissions(permissions ...string) Handler {
	return func(c *Context) {
		user := c.GetUser()
		if user == nil || !user.HasPermissions(permissions...) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}

var limiters = make(map[string]Handler)

func WithRateLimit(prefix, formatted string) Handler {
	prefix = fmt.Sprintf("limiter-%s", prefix)

	handler, ok := limiters[prefix]
	if !ok {
		rate, err := limiter.NewRateFromFormatted(formatted)
		if err != nil {
			logger.Err.Fatalln(err)
		}

		store, err := redis.NewStoreWithOptions(cache.Redis, limiter.StoreOptions{
			Prefix: prefix,
		})
		if err != nil {
			logger.Err.Fatalln(err)
		}

		instance := limiter.New(store, rate, limiter.WithTrustForwardHeader(true))
		m := mgin.NewMiddleware(instance)

		handler = func(c *Context) {
			if u := c.GetUser(); u == nil || len(u.Permissions) <= 2 {
				m(c.Context)
			}
		}
		limiters[prefix] = handler
	}
	return handler
}
