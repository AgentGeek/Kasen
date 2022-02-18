package main

import (
	"log"
	"os"

	"kasen/controllers"
	"kasen/controllers/api"
	"kasen/server"
	"kasen/services"
)

func init() {
	os.Setenv("MALLOC_ARENA_MAX", "2")

	if err := services.MkdirAll(services.GetTempDir()); err != nil {
		log.Fatalln(err)
	}

	if err := services.MkdirAll(services.GetChaptersSymlinksDir()); err != nil {
		log.Fatalln(err)
	}

	setup()
}

func main() {
	services.RemapSymlinks()

	server.Init()
	controllers.Init()
	api.Init()

	server.Start()
}
