package server

import (
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"kasen/config"
	"kasen/modext"
	"kasen/services"

	"github.com/gin-gonic/gin"
)

type Context struct {
	*gin.Context

	sync.RWMutex
	MapData map[string]interface{}
}

func (c *Context) GetURL() string {
	u, _ := url.Parse(config.GetMeta().BaseURL)
	u.Path = c.Request.URL.Path
	u.RawQuery = c.Request.URL.RawQuery
	return u.String()
}

func (c *Context) preHTML(code *int) {
	if err, ok := c.GetData("error"); ok {
		err := strings.ToLower(err.(error).Error())
		if strings.Contains(err, "does not exist") || strings.Contains(err, "not found") {
			*code = http.StatusNotFound
		}
	}

	c.SetData("status", *code)
	c.SetData("statusText", http.StatusText(*code))

	if v, ok := c.MapData["name"]; !ok || len(v.(string)) == 0 {
		c.SetData("name", http.StatusText(*code))
	}

	meta := config.GetMeta()
	c.SetData("title", meta.Title)
	c.SetData("description", meta.Description)
	c.SetData("baseURL", meta.BaseURL)
	c.SetData("language", meta.Language)

	c.SetData("url", c.GetURL())
	c.SetData("query", c.Request.URL.Query())
}

func (c *Context) HTML(code int, name string) {
	c.preHTML(&code)
	renderTemplate(c, false, &RenderOptions{
		Status: code,
		Name:   name,
		Data:   c.MapData,
	})
}

func (c *Context) Cache(code int, name string) {
	if gin.Mode() == gin.DebugMode {
		c.HTML(code, name)
	} else {
		c.preHTML(&code)
		renderTemplate(c, true, &RenderOptions{
			Status: code,
			Name:   name,
			Key:    c.GetURL(),
			Data:   c.MapData,
		})
	}
}

func (c *Context) IsCached(name string) bool {
	_, ok := getTemplate(name, c.GetURL())
	return ok
}

func (c *Context) TryCache(name string) bool {
	if c.IsCached(name) {
		c.Cache(http.StatusOK, name)
		return true
	}
	return false
}

func (c *Context) ErrorJSON(code int, message string, err error) {
	c.JSON(code, gin.H{
		"error": gin.H{
			"message": message,
			"cause":   err.Error(),
		},
	})
}

func (c *Context) GetUser() *modext.User {
	if u, ok := c.Get("user"); ok {
		return u.(*modext.User)
	}

	uid, ok := c.GetUserID()
	if !ok {
		return nil
	}

	u, err := services.GetUser(uid)
	if err != nil {
		return nil
	}

	c.Set("user", u)
	c.SetData("user", u)

	return u
}

func (c *Context) GetUserID() (uid int64, ok bool) {
	if uid = c.GetInt64("uid"); uid <= 0 {
		if uid, ok = c.VerifySessionHeader(); !ok {
			if uid, ok = c.VerifySessionCookie(); !ok {
				return
			}
		}
	}
	return uid, uid > 0
}

func (c *Context) GetData(key string) (interface{}, bool) {
	c.RLock()
	defer c.RUnlock()

	v, exists := c.MapData[key]
	return v, exists
}

func (c *Context) SetData(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()

	if c.MapData == nil {
		c.MapData = make(map[string]interface{})
	}
	c.MapData[key] = value
}

func (c *Context) SetTokens(sessionToken *services.Token, refreshToken *services.Token) {
	if sessionToken == nil {
		c.SetCookie("session", "", nil)
	} else {
		c.SetCookie("session", sessionToken.String, sessionToken.ExprDate)
	}

	if refreshToken == nil {
		c.SetCookie("refresh", "", nil)
	} else {
		c.SetCookie("refresh", refreshToken.String, refreshToken.ExprDate)
	}
}

func (c *Context) SetCookie(name, value string, expires *time.Time) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		Path:     "/",
		SameSite: http.SameSiteLaxMode,
		Secure:   c.Request.TLS != nil || strings.HasPrefix(config.GetMeta().BaseURL, "https"),
		HttpOnly: true,
	}

	if expires == nil {
		cookie.MaxAge = -1
	} else {
		cookie.Expires = *expires
	}

	http.SetCookie(c.Writer, cookie)
}

func (c *Context) VerifySessionCookie() (uid int64, ok bool) {
	if _, ok := c.Get("VerifySessionCookie"); ok {
		uid = c.GetInt64("uid")
		return uid, uid > 0
	}

	var err error
	c.Set("VerifySessionCookie", 1)

	sessionToken, _ := c.Cookie("session")
	if len(sessionToken) > 0 {
		uid, err = services.VerifySessionToken(sessionToken)
	}

	refreshToken, _ := c.Cookie("refresh")
	if len(refreshToken) > 0 && (len(sessionToken) == 0 || err != nil) {
		var t *services.Token
		uid, t, err = services.RefreshToken(refreshToken)
		if err == nil {
			c.SetCookie("session", t.String, t.ExprDate)
		}
	}

	if uid == 0 || err != nil {
		c.SetTokens(nil, nil)
		return
	}

	c.Set("uid", uid)
	return uid, true
}

func (c *Context) VerifySessionHeader() (uid int64, ok bool) {
	if _, ok := c.Get("VerifySessionHeader"); ok {
		uid = c.GetInt64("uid")
		return uid, uid > 0
	}

	c.Set("VerifySessionHeader", 1)

	auth := c.GetHeader("Authorization")
	if len(auth) == 0 {
		return
	}

	token := strings.TrimPrefix(auth, "Bearer ")
	if len(token) == 0 {
		return
	}

	uid, err := services.VerifySessionToken(token)
	if err != nil {
		return
	}

	c.Set("uid", uid)
	return uid, true
}

func (c *Context) ParamInt(name string) (int, error) {
	return strconv.Atoi(c.Param(name))
}

func (c *Context) ParamInt64(name string) (int64, error) {
	return strconv.ParseInt(c.Param(name), 10, 64)
}
