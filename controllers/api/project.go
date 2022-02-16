package api

import (
	"net/http"

	"kasen/constants"
	"kasen/server"
	"kasen/services"

	"github.com/gin-gonic/gin"
)

func GetProjectCacheStats(c *server.Context) {
	c.JSON(http.StatusOK, services.GetProjectCacheStats())
}

type CheckProjectExistsQueries struct {
	ID    int64  `form:"id"`
	Slug  string `form:"slug"`
	Title string `form:"title"`
}

type CheckProjectExistsResponse struct {
	ID   int64  `json:"id"`
	Slug string `json:"slug"`
}

func CheckProjectExists(c *server.Context) {
	body := &CheckProjectExistsQueries{}
	c.BindQuery(body)

	var id int64
	var slug string
	if body.ID > 0 {
		id, slug = services.CheckProjectExists(body.ID)
	} else if len(body.Slug) > 0 {
		id, slug = services.CheckProjectExistsBySlug(body.Slug)
	} else if len(body.Title) > 0 {
		id, slug = services.CheckProjectExistsByTitle(body.Title)
	} else {
		c.Status(http.StatusBadRequest)
		return
	}

	if id > 0 {
		c.JSON(http.StatusOK, gin.H{"id": id, "slug": slug})
	} else {
		c.Status(http.StatusNoContent)
	}
}

func CreateProject(c *server.Context) {
	draft := &services.ProjectDraft{}
	c.BindJSON(draft)

	project, err := services.CreateProject(draft)
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to create project", err)
		return
	}
	c.JSON(http.StatusCreated, project)
}

func DeleteProject(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	if err = services.DeleteProject(id); err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to delete project", err)
		return
	}
	c.Status(http.StatusNoContent)
}

func GetProject(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil || id <= 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	opts := services.GetProjectOptions{}
	c.BindQuery(&opts)

	if opts.IncludesDrafts {
		user := c.GetUser()
		if user == nil || !user.HasPermissions(constants.PermsProject...) {
			c.Status(http.StatusForbidden)
			return
		}
	}

	result := services.GetProject(id, opts)
	if result.Err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get project", result.Err)
		return
	}
	c.JSON(http.StatusOK, result.Project)
}

func GetProjectMd(c *server.Context) {
	id := c.Param("id")
	if len(id) == 0 {
		c.Status(http.StatusBadRequest)
		return
	}

	project, err := services.GetProjectMd(id)
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get project metadata", err)
		return
	}
	c.JSON(http.StatusOK, project)
}

func GetProjects(c *server.Context) {
	opts := services.GetProjectsOptions{}
	c.BindQuery(&opts)

	if opts.IncludesDrafts {
		user := c.GetUser()
		if user == nil || !user.HasPermissions(constants.PermsProject...) {
			c.Status(http.StatusForbidden)
			return
		}
	}

	result := services.GetProjects(opts)
	if result.Err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get projects", result.Err)
		return
	}
	c.JSON(http.StatusOK, result)
}

func LockProject(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	project, err := services.LockProject(id)
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to lock project", err)
		return
	}
	c.JSON(http.StatusOK, project)
}

func PublishProject(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	project, err := services.PublishProject(id)
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to publish project", err)
		return
	}
	c.JSON(http.StatusOK, project)
}

func UnlockProject(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	project, err := services.UnlockProject(id)
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to unlock project", err)
		return
	}
	c.JSON(http.StatusOK, project)
}

func UnpublishProject(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	project, err := services.UnpublishProject(id)
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to unpublish project", err)
		return
	}
	c.JSON(http.StatusOK, project)
}

func UpdateProject(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	draft := &services.ProjectDraft{}
	c.BindJSON(draft)

	project, err := services.UpdateProject(id, draft)
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to update project", err)
		return
	}

	c.JSON(http.StatusOK, project)
}
