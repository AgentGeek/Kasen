package services

import (
	"fmt"
	"strconv"
	"time"

	"kasen/config"

	"github.com/gorilla/feeds"
)

func CreateProjectFeed() *feeds.Feed {
	meta := config.GetMeta()

	feed := &feeds.Feed{
		Title:       fmt.Sprintf("%s Project RSS", meta.BaseURL),
		Link:        &feeds.Link{Href: JoinURL(meta.BaseURL, "/rss/projects")},
		Description: fmt.Sprintf("RSS feed for %s projects", meta.Title),
	}

	r := GetProjects(GetProjectsOptions{})
	for _, p := range r.Projects {
		feed.Items = append(feed.Items, &feeds.Item{
			Id:    strconv.Itoa(int(p.ID)),
			Title: p.Title,
			Link: &feeds.Link{
				Href: JoinURL(meta.BaseURL, fmt.Sprintf("projects/%d/%s", p.ID, p.Slug)),
			},
			Created: time.Unix(p.CreatedAt, 0).UTC(),
			Updated: time.Unix(p.UpdatedAt, 0).UTC(),
		})
	}
	return feed
}
