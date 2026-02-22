package main

import (
	"audiodrive/internal/controller"
	"audiodrive/internal/database"
	"audiodrive/internal/repo"
	"audiodrive/internal/router"
	"log"
	"net/http"
)

func main() {

	db, err := database.SetupPostgres()
	if err != nil {
		log.Fatalf("Failed to setup postgres: %s", err.Error())
	}

	database.RunMigrations(db)

	userRepo := repo.NewUserRepo(db)
	objectRepo := repo.NewAudioObjectRepo(db)

	r := router.NewHandler(
		controller.NewUserController(userRepo),
		controller.NewAudioObjectController(objectRepo),
	)

	log.Println("Server listening on :8080")
	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal(err)
	}
}
