package controllers

import (
	"net/http"

	"kasen/config"
	"kasen/server"
	"kasen/services"
)

func ManagePage(c *server.Context) {
	if c.GetUser() == nil {
		c.Redirect(http.StatusFound, "/login")
		return
	}
	c.Cache(http.StatusOK, "manage.html")
}

func LoginPage(c *server.Context) {
	if c.GetUser() != nil {
		c.Redirect(http.StatusFound, "/manage")
		return
	}
	c.HTML(http.StatusOK, "login.html")
}

type LoginRequest struct {
	Email       string `form:"email"`
	RawPassword string `form:"password"`
}

func Login(c *server.Context) {
	if c.GetUser() != nil {
		c.Redirect(http.StatusFound, "/manage")
		return
	}

	payload := &LoginRequest{}
	c.Bind(payload)

	rt, st, err := services.Login(services.LoginOptions{
		Email:       payload.Email,
		RawPassword: payload.RawPassword,
	})
	if err != nil {
		c.SetData("error", err)
		c.HTML(http.StatusInternalServerError, "login.html")
		return
	}

	c.SetTokens(st, rt)
	c.Redirect(http.StatusFound, "/manage")
}

func RegisterPage(c *server.Context) {
	if config.GetService().DisableRegistration {
		c.Redirect(http.StatusFound, "/login")
		return
	} else if c.GetUser() != nil {
		c.Redirect(http.StatusFound, "/manage")
		return
	}
	c.HTML(http.StatusOK, "register.html")
}

type RegisterRequest struct {
	Name        string `form:"name"`
	Email       string `form:"email"`
	RawPassword string `form:"password"`
}

func Register(c *server.Context) {
	if config.GetService().DisableRegistration {
		c.Redirect(http.StatusFound, "/login")
		return
	} else if c.GetUser() != nil {
		c.Redirect(http.StatusFound, "/manage")
		return
	}

	payload := &RegisterRequest{}
	c.Bind(payload)

	rt, st, err := services.Register(services.CreateUserOptions{
		Name:        payload.Name,
		Email:       payload.Email,
		RawPassword: payload.RawPassword,
	})
	if err != nil {
		c.SetData("error", err)
		c.HTML(http.StatusInternalServerError, "register.html")
		return
	}

	c.SetTokens(st, rt)
	c.Redirect(http.StatusFound, "/manage")
}

func Logout(c *server.Context) {
	st, _ := c.Cookie("session")
	rt, _ := c.Cookie("refresh")

	services.Logout(st, rt)

	c.SetTokens(nil, nil)
	c.Redirect(http.StatusFound, "/")
}
