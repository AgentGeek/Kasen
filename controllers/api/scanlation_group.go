package api

import (
	"net/http"

	"kasen/modext"
	"kasen/server"
	"kasen/services"
)

func CreateScanlationGroup(c *server.Context) {
	payload := &CreateEntityPayload{}
	c.BindJSON(payload)

	scanlationGroup, err := services.CreateScanlationGroup(payload.Name)
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to create scanlation group", err)
		return
	}
	c.JSON(http.StatusOK, scanlationGroup)
}

func DeleteScanlationGroup(c *server.Context) {
	if err := services.DeleteScanlationGroupBySlugOrName(c.Param("identifier")); err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to delete scanlation group", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func GetScanlationGroup(c *server.Context) {
	queries := &GetEntityQueries{}
	c.BindQuery(queries)

	if queries.ID == 0 && len(queries.Slug) == 0 && len(queries.Name) == 0 {
		c.ErrorJSON(http.StatusBadRequest, "Must specify either id, slug or name", nil)
		return
	}

	var scanlationGroup *modext.ScanlationGroup
	var err error

	if queries.ID > 0 {
		scanlationGroup, err = services.GetScanlationGroup(queries.ID)
	} else if len(queries.Slug) > 0 {
		scanlationGroup, err = services.GetScanlationGroupBySlug(queries.Slug)
	} else {
		scanlationGroup, err = services.GetScanlationGroupByName(queries.Name)
	}

	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get scanlation group", err)
		return
	}
	c.JSON(http.StatusOK, scanlationGroup)
}

func GetScanlationGroups(c *server.Context) {
	scanlationGroups, err := services.GetScanlationGroups()
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get scanlation groups", err)
		return
	}
	c.JSON(http.StatusOK, scanlationGroups)
}
