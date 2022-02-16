package controllers

import (
	"net/http"
	"strings"

	"kasen/server"
	"kasen/services"

	"github.com/gorilla/feeds"
)

func createFeed(c *server.Context, isAtom bool) {
	t := strings.TrimRight(strings.TrimLeft(c.Param("type"), "/"), "/")
	if !strings.EqualFold(t, "projects") && !strings.EqualFold(t, "chapters") {
		c.Status(http.StatusBadRequest)
		return
	}

	var feed *feeds.Feed
	if strings.EqualFold(t, "projects") {
		feed = services.CreateProjectFeed()
	} else {
		feed = services.CreateChapterFeed()
	}

	var str string
	if isAtom {
		str, _ = feed.ToAtom()
	} else {
		str, _ = feed.ToRss()
	}
	c.Data(http.StatusOK, "application/xml", []byte(str))

}

func RSS(c *server.Context) {
	createFeed(c, false)
}

func Atom(c *server.Context) {
	createFeed(c, true)
}
