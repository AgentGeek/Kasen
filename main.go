package main

import (
	"os"

	"kasen/controllers"
	"kasen/controllers/api"
	"kasen/server"
	"kasen/services"

	"github.com/rs1703/logger"
)

func init() {
	logger.SetOutput("kasen.log")

	os.Setenv("MALLOC_ARENA_MAX", "2")
	os.MkdirAll(services.GetTempDir(), os.ModePerm)
	os.MkdirAll(services.GetChaptersSymlinksDir(), os.ModePerm)

	setup()
}

func main() {
	server.Init()

	controllers.Init()
	api.Init()

	server.Start()
}
