package api

import (
	"mime/multipart"
	"net/http"

	"kasen/constants"
	"kasen/modext"
	"kasen/server"
	"kasen/services"
)

func GetCoverCacheStats(c *server.Context) {
	c.JSON(http.StatusOK, services.GetCoverCacheStats())
}

func DeleteCover(c *server.Context) {
	_, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	cid, err := c.ParamInt64("cid")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := services.DeleteCover(cid); err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to delete cover", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func GetCover(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	result := services.GetCover(id)
	if result.Err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get cover", result.Err)
		return
	}
	c.JSON(http.StatusOK, result.Cover)
}

func GetCovers(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	result := services.GetCovers(id)
	if result.Err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get covers", result.Err)
		return
	}
	c.JSON(http.StatusOK, result.Covers)
}

func SetCover(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	cid, err := c.ParamInt64("cid")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err := services.SetCover(id, cid); err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to set project cover", err)
		return
	}
	c.Status(http.StatusNoContent)
}

type UploadCoverPayload struct {
	URL            string `form:"url"`
	IsInitialCover bool   `form:"isInitialCover"`
	SetAsMainCover bool   `form:"setAsMainCover"`
}

func UploadCover(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	payload := &UploadCoverPayload{}
	c.Bind(payload)

	var cover *modext.Cover
	if len(payload.URL) > 0 {
		cover, err = services.UploadCoverFromSource(id, payload.URL, c.GetUser())
	} else {
		var fh *multipart.FileHeader
		if fh, err = c.FormFile("data"); err != nil {
			c.Status(http.StatusBadRequest)
			return
		}
		cover, err = services.UploadCoverMultipart(id, fh, c.GetUser())
	}

	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to upload cover", err)
		return
	}

	if payload.IsInitialCover || (payload.SetAsMainCover && c.GetUser().HasPermissions(constants.PermSetCover)) {
		if err := services.SetCover(id, cover.ID); err != nil {
			c.ErrorJSON(http.StatusInternalServerError, "Failed to set project main cover", err)
			return
		}
	}
	c.JSON(http.StatusOK, cover)
}
