package main

import (
	"github.com/Sovpalo/sovpalo-backend"
	"github.com/Sovpalo/sovpalo-backend/pkg/handler"
	"log"
)

func main() {
	handlers := new(handler.Handler)
	srv := new(sovpalo.Server)
	log.Println("server starting on :8000")
	if err := srv.Run("8000", handlers.InitRoutes()); err != nil {
		log.Fatalf("error occured while running server: %s", err.Error())
	}
}
