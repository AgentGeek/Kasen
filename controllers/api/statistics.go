package api

import (
	"net/http"

	"kasen/server"
	"kasen/services"
)

func GetProjectStats(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	result := services.GetProjectStats(id)
	if result.Err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get project stats", result.Err)
		return
	}

	c.JSON(http.StatusOK, result.Stats)
}

func GetChapterStats(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	result := services.GetChapterStats(id)
	if result.Err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get chapter stats", result.Err)
		return
	}

	c.JSON(http.StatusOK, result.Stats)
}
