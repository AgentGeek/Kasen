package controllers

import "kasen/server"

func Init() {
	var (
		GET  = server.GET
		POST = server.POST

		WithName            = server.WithName
		WithAuthorization   = server.WithAuthorization
		WithNoAuthorization = server.WithNoAuthorization
		WithRateLimit       = server.WithRateLimit
		WithRedirect        = server.WithRedirect
	)

	GET("/", WithName("Home"), Home)
	GET("/rss/*type", RSS)
	GET("/atom/*type", Atom)

	GET("/covers/:slug/:fileName", Cover)
	GET("/covers/:slug/:fileName/*width", Cover)

	GET("/projects/:id", Project)
	GET("/projects/:id/*slug", Project)
	GET("/projects", WithName("Browse Projects"), Projects)

	GET("/chapters/:id", Chapter)
	GET("/chapters/:id/*any", Chapter)
	GET("/chapters", WithName("Browse Chapters"), Chapters)

	GET("/pages/:id/:fileName", Page)
	GET("/pages/:id/:fileName/*width", Page)

	GET("/login",
		WithNoAuthorization(WithRedirect("/manage")),
		WithName("Login"),
		LoginPage)
	POST("/login",
		WithNoAuthorization(WithRedirect("/manage")),
		WithRateLimit("auth-login", "5-H"),
		WithName("Login"),
		Login,
	)

	GET("/register",
		WithNoAuthorization(WithRedirect("/manage")),
		WithName("Register"),
		RegisterPage)
	POST("/register",
		WithNoAuthorization(WithRedirect("/manage")),
		WithRateLimit("auth-register", "5-H"),
		WithName("Register"),
		Register)

	GET("/logout", Logout)

	GET("/manage",
		WithAuthorization(WithRedirect("/login")),
		WithName("Manage"),
		ManagePage)
	GET("/manage/*any",
		WithAuthorization(WithRedirect("/login")),
		WithName("Manage"),
		ManagePage)
}
