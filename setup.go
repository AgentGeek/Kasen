package main

import (
	"fmt"
	"log"

	"kasen/config"
	"kasen/constants"
	"kasen/services"

	"github.com/google/uuid"
)

func fatalln(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func setup() {
	if config.GetInitialized() {
		return
	}

	setupUser()
	generateTags()

	config.SetInitialized(true)
	if err := config.Save(); err != nil {
		log.Fatalln(err)
	}

	fmt.Println("\nSetup completed")
}

func setupUser() {
	if users, err := services.GetUsers(); len(users) > 0 && err == nil {
		return
	}

	fmt.Println("\n=== Creating initial account ===")
	fmt.Println("These could be changed later")

	defaultName := "Kasen"
	defaultEmail := "admin@example.com"
	defaultPassword := uuid.NewString()

	for {
		params := services.CreateUserOptions{}

		fmt.Printf("\nName (default: %s)\n> ", defaultName)
		fmt.Scanln(&params.Name)
		if len(params.Name) == 0 {
			params.Name = defaultName
		}

		fmt.Printf("\nEmail (default: %s)\n> ", defaultEmail)
		fmt.Scanln(&params.Email)
		if len(params.Email) == 0 {
			params.Email = defaultEmail
		}

		fmt.Printf("\nPassword (default randomized: %s)\n> ", defaultPassword)
		fmt.Scanln(&params.RawPassword)
		if len(params.RawPassword) == 0 {
			params.RawPassword = defaultPassword
		}

		user, err := services.CreateUser(params)
		if err != nil {
			log.Println("unable to create account:", err.Error())
			continue
		}

		if _, err := services.UpdateUserPermissions(user, constants.Perms); err != nil {
			log.Fatalln(err)
		}
		break
	}
}
