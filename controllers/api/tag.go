package api

import (
	"net/http"

	"kasen/modext"
	"kasen/server"
	"kasen/services"
)

func GetTag(c *server.Context) {
	body := &GetEntityQueries{}
	c.BindQuery(body)

	if body.ID == 0 && len(body.Slug) == 0 && len(body.Name) == 0 {
		c.ErrorJSON(http.StatusBadRequest, "Must specify either id, slug or name", nil)
		return
	}

	var tag *modext.Tag
	var err error

	if body.ID > 0 {
		tag, err = services.GetTag(body.ID)
	} else if len(body.Slug) > 0 {
		tag, err = services.GetTagBySlug(body.Slug)
	} else {
		tag, err = services.GetTagByName(body.Name)
	}

	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get tag", err)
		return
	}

	c.JSON(http.StatusOK, tag)
}

func CreateTag(c *server.Context) {
	payload := &CreateEntityPayload{}
	c.BindJSON(payload)

	tag, err := services.CreateTag(payload.Name)
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to create tag", err)
		return
	}
	c.JSON(http.StatusOK, tag)
}

func DeleteTag(c *server.Context) {
	if err := services.DeleteTagBySlugOrName(c.Param("identifier")); err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to delete tag", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func GetTags(c *server.Context) {
	tags, err := services.GetTags()
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get tags", err)
		return
	}
	c.JSON(http.StatusOK, tags)
}
