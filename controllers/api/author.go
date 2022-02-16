package api

import (
	"net/http"

	"kasen/modext"
	"kasen/server"
	"kasen/services"
)

func GetAuthor(c *server.Context) {
	queries := &GetEntityQueries{}
	c.BindQuery(queries)

	if queries.ID == 0 && len(queries.Slug) == 0 && len(queries.Name) == 0 {
		c.ErrorJSON(http.StatusBadRequest, "Must specify either id, slug or name", nil)
		return
	}

	var author *modext.Author
	var err error

	if queries.ID > 0 {
		author, err = services.GetAuthor(queries.ID)
	} else if len(queries.Slug) > 0 {
		author, err = services.GetAuthorBySlug(queries.Slug)
	} else {
		author, err = services.GetAuthorByName(queries.Name)
	}

	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get author", err)
		return
	}
	c.JSON(http.StatusOK, author)
}

func CreateAuthor(c *server.Context) {
	payload := &CreateEntityPayload{}
	c.BindJSON(payload)

	author, err := services.CreateAuthor(payload.Name)
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to create author", err)
		return
	}
	c.JSON(http.StatusOK, author)
}

func DeleteAuthor(c *server.Context) {
	if err := services.DeleteAuthorBySlugOrName(c.Param("identifier")); err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to delete author", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func GetAuthors(c *server.Context) {
	authors, err := services.GetAuthors()
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get authors", err)
		return
	}
	c.JSON(http.StatusOK, authors)
}
