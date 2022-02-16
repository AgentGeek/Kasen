package api

import (
	"net/http"

	"kasen/constants"
	"kasen/server"
	"kasen/services"
)

func GetChapterCacheStats(c *server.Context) {
	c.JSON(http.StatusOK, services.GetChapterCacheStats())
}

func CreateChapter(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	draft := services.ChapterDraft{}
	c.BindJSON(&draft)

	chapter, err := services.CreateChapter(id, draft, c.GetUser())
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to create chapter", err)
		return
	}
	c.JSON(http.StatusCreated, chapter)
}

func DeleteChapter(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err = services.DeleteChapter(id, c.GetUser()); err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to delete chapter", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func GetChapter(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	opts := services.GetChapterOptions{}
	c.BindQuery(&opts)

	if opts.IncludesDrafts {
		user := c.GetUser()
		if user == nil || !user.HasPermissions(constants.PermsChapter...) {
			c.Status(http.StatusForbidden)
			return
		}
	}

	result := services.GetChapter(id, opts)
	if result.Err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get chapter", result.Err)
		return
	}
	c.JSON(http.StatusOK, result.Chapter)
}

func GetChapterMd(c *server.Context) {
	id := c.Param("id")
	if len(id) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	chapter, err := services.GetChapterMd(id)
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get chapter metadata", err)
		return
	}
	c.JSON(http.StatusOK, chapter)
}

func GetChapters(c *server.Context) {
	opts := services.GetChaptersOptions{}
	c.BindQuery(&opts)

	if opts.IncludesDrafts {
		user := c.GetUser()
		if user == nil || !user.HasPermissions(constants.PermsChapter...) {
			c.Status(http.StatusForbidden)
			return
		}
	}

	result := services.GetChapters(opts)
	if result.Err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get chapters", result.Err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func GetChaptersByProject(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	opts := services.GetChaptersOptions{}
	c.BindQuery(&opts)

	if opts.IncludesDrafts {
		user := c.GetUser()
		if user == nil || !user.HasPermissions(constants.PermsChapter...) {
			c.Status(http.StatusForbidden)
			return
		}
	}
	opts.ProjectID = id

	result := services.GetChapters(opts)
	if result.Err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get chapters", result.Err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func LockChapter(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	chapter, err := services.LockChapter(id, c.GetUser())
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to lock chapter", err)
		return
	}
	c.JSON(http.StatusOK, chapter)
}

func PublishChapter(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	chapter, err := services.PublishChapter(id, c.GetUser())
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to publish chapter", err)
		return
	}
	c.JSON(http.StatusOK, chapter)
}

func UnlockChapter(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	chapter, err := services.UnlockChapter(id, c.GetUser())
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to unlock chapter", err)
		return
	}
	c.JSON(http.StatusOK, chapter)
}

func UnpublishChapter(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	chapter, err := services.UnpublishChapter(id, c.GetUser())
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to unpublish chapter", err)
		return
	}
	c.JSON(http.StatusOK, chapter)
}

func UpdateChapter(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	draft := &services.ChapterDraft{}
	c.BindJSON(draft)

	chapter, err := services.UpdateChapter(id, draft, c.GetUser())
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to update chapter", err)
		return
	}
	c.JSON(http.StatusOK, chapter)
}
