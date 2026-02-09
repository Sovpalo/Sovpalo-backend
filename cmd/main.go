package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Sovpalo/sovpalo-backend"
	"github.com/Sovpalo/sovpalo-backend/internal/config"
	"github.com/Sovpalo/sovpalo-backend/internal/db"
	"github.com/Sovpalo/sovpalo-backend/pkg/handler"
	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
	"github.com/Sovpalo/sovpalo-backend/pkg/service"
)

func main() {
	cfg := config.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	pool, err := db.NewPostgres(ctx, cfg)
	if err != nil {
		log.Fatalf("database connection error: %s", err.Error())
	}
	defer pool.Close()

	redisClient, err := db.NewRedis(ctx, cfg)
	if err != nil {
		log.Fatalf("redis connection error: %s", err.Error())
	}
	defer redisClient.Close()

	healthRepo := repository.NewCompositeHealthRepository(
		repository.NewPostgresHealthRepository(pool),
		repository.NewRedisHealthRepository(redisClient),
	)
	healthService := service.NewHealthService(healthRepo)
	repos := repository.NewRepository(pool, redisClient)
	services := service.NewService(repos)
	handlers := handler.NewHandler(healthService, services)

	srv := new(sovpalo.Server)
	go func() {
		log.Printf("server starting on :%s", cfg.Port)
		if err := srv.Run(cfg.Port, handlers.InitRoutes()); err != nil {
			log.Fatalf("error occured while running server: %s", err.Error())
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("error while shutting down server: %s", err.Error())
	}
}
