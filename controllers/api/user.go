package api

import (
	"net/http"

	"kasen/errs"
	"kasen/server"
	"kasen/services"
)

type User struct {
	ID          int64    `json:"id"`
	Name        string   `json:"name"`
	Email       string   `json:"email"`
	Permissions []string `json:"permissions,omitempty"`
}

type DeleteUserPayload struct {
	Password string `json:"password"`
}

func DeleteUser(c *server.Context) {
	payload := DeleteUserPayload{}
	c.BindJSON(&payload)

	user := c.GetUser()
	if err := user.CheckPassword(payload.Password); err != nil {
		c.ErrorJSON(http.StatusUnauthorized, "Failed to delete user", errs.ErrInvalidCredentials)
		return
	}

	if err := services.DeleteUser(user); err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to delete user", err)
		return
	}

	c.SetCookie("refresh", "", nil)
	c.SetCookie("session", "", nil)

	c.Status(http.StatusNoContent)
}

func DeleteUserById(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	user, err := services.GetUser(id)
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get user", err)
		return
	}

	if err := services.DeleteUser(user); err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to delete user", err)
		return
	}

	c.Status(http.StatusNoContent)
}

func GetUser(c *server.Context) {
	user := c.GetUser()
	c.JSON(200, User{
		ID:          user.ID,
		Name:        user.Name,
		Email:       user.Email,
		Permissions: user.Permissions,
	})
}

func GetUsers(c *server.Context) {
	users, err := services.GetUsers()
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get users", err)
		return
	}
	c.JSON(http.StatusOK, users)
}

type UpdateUserNamePayload struct {
	Name string `json:"name"`
}

func UpdateUserName(c *server.Context) {
	payload := UpdateUserNamePayload{}
	c.BindJSON(&payload)

	if err := services.UpdateUserName(c.GetUser(), payload.Name); err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to update user name", err)
	}
	c.Status(http.StatusNoContent)
}

func UpdateUserNameById(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	payload := UpdateUserNamePayload{}
	c.BindJSON(&payload)

	user, err := services.GetUser(id)
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get user", err)
		return
	}

	if err := services.UpdateUserName(user, payload.Name); err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to update user name", err)
		return
	}

	c.Status(http.StatusNoContent)
}

func UpdateUserPassword(c *server.Context) {
	payload := services.UpdateUserPasswordOptions{}
	c.BindJSON(&payload)

	if err := services.UpdateUserPassword(c.GetUser(), payload); err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to update user password", err)
	}
	c.Status(http.StatusNoContent)
}

func UpdateUserPasswordById(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	payload := services.UpdateUserPasswordOptions{}
	c.BindJSON(&payload)

	user, err := services.GetUser(id)
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get user", err)
		return
	}

	if err := services.UpdateUserPassword(user, payload); err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to update user password", err)
		return
	}

	c.Status(http.StatusNoContent)
}

type UpdateUserPermissionsPayload struct {
	Permissions []string `json:"permissions"`
}

func UpdateUserPermissions(c *server.Context) {
	id, err := c.ParamInt64("id")
	if err != nil {
		c.Status(http.StatusBadRequest)
		return
	}

	payload := UpdateUserPermissionsPayload{}
	c.BindJSON(&payload)

	user, err := services.GetUser(id)
	if err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to get user", err)
		return
	}

	if _, err := services.UpdateUserPermissions(user, payload.Permissions); err != nil {
		c.ErrorJSON(http.StatusInternalServerError, "Failed to update user permissions", err)
		return
	}
	c.Status(http.StatusNoContent)
}
