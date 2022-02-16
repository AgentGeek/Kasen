package api

import (
	"net/http"

	"kasen/config"
	"kasen/server"
	"kasen/services"
)

func GetServiceConfig(c *server.Context) {
	c.JSON(http.StatusOK, services.GetServiceConfig())
}

func UpdateMeta(c *server.Context) {
	payload := &config.Meta{}
	c.BindJSON(payload)

	if err := services.UpdateMeta(payload); err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to update meta", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func UpdateServiceConfig(c *server.Context) {
	payload := &config.Service{}
	c.BindJSON(payload)

	if err := services.UpdateServiceConfig(payload); err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to update service config", err)
		return
	}
	c.Status(http.StatusNoContent)
}
