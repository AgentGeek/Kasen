package api

import (
	"mime/multipart"
	"net/http"

	"kasen/server"
	"kasen/services"
)

func GetPagesCacheStats(c *server.Context) {
	c.JSON(http.StatusOK, services.GetPagesCacheStats())
}

func DeletePage(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	fileName := c.Param("fileName")
	if len(fileName) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	pages, err := services.DeletePage(id, fileName, c.GetUser())
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to delete page", err)
		return
	}
	c.JSON(http.StatusOK, pages)
}

func GetPages(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	result := services.GetPages(id)
	if result.Err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get pages", result.Err)
		return
	}
	c.JSON(http.StatusOK, result.Pages)
}

type UploadPagePayload struct {
	URL string `form:"url"`
}

func GetPagesMd(c *server.Context) {
	id := c.Param("id")
	if len(id) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	pages, err := services.GetPagesMd(id)
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get pages", err)
		return
	}
	c.JSON(http.StatusOK, pages)
}

func UploadPage(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	payload := &UploadPagePayload{}
	c.Bind(payload)

	var pages []string
	if len(payload.URL) > 0 {
		pages, err = services.UploadPageFromSource(id, payload.URL, c.GetUser())
	} else {
		var fh *multipart.FileHeader
		if fh, err = c.FormFile("data"); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		pages, err = services.UploadPageMultipart(id, fh, c.GetUser())
	}

	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to upload page", err)
		return
	}
	c.JSON(http.StatusOK, pages)
}
