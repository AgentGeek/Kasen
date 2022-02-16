package api

import (
	"net/http"

	"kasen/server"
	"kasen/services"
)

func RemapSymlinks(c *server.Context) {
	go services.RemapSymlinks()
	c.Status(http.StatusOK)
}

func RefreshTemplates(c *server.Context) {
	go server.LoadTemplates()
	c.Status(http.StatusOK)
}
