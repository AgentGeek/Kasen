package services

import (
	"fmt"
	"strconv"
	"time"

	"kasen/config"

	"github.com/gorilla/feeds"
)

func CreateChapterFeed() *feeds.Feed {
	meta := config.GetMeta()

	feed := &feeds.Feed{
		Title:       fmt.Sprintf("%s Chapter RSS", meta.Title),
		Link:        &feeds.Link{Href: JoinURL(meta.BaseURL, "/rss/chapters")},
		Description: fmt.Sprintf("RSS feed for %s chapters", meta.Title),
	}

	r := GetChapters(GetChaptersOptions{
		Preloads: []string{ChapterRels.Project},
	})
	for _, c := range r.Chapters {
		feed.Items = append(feed.Items, &feeds.Item{
			Id:    strconv.Itoa(int(c.ID)),
			Title: fmt.Sprintf("%s - %s", c.Project.Title, FormatChapter(c)),
			Link: &feeds.Link{
				Href: JoinURL(meta.BaseURL, fmt.Sprintf("chapters/%d", c.ID)),
			},
			Created: time.Unix(c.CreatedAt, 0).UTC(),
			Updated: time.Unix(c.UpdatedAt, 0).UTC(),
		})
	}
	return feed
}
