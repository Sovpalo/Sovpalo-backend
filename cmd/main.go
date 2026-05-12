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
	"github.com/Sovpalo/sovpalo-backend/internal/telegram"
	"github.com/Sovpalo/sovpalo-backend/pkg/handler"
	"github.com/Sovpalo/sovpalo-backend/pkg/repository"
	"github.com/Sovpalo/sovpalo-backend/pkg/service"
)

func main() {
	cfg := config.Load()

	bootstrapCtx, bootstrapCancel := context.WithTimeout(context.Background(), 10*time.Second)
	telegramUsername := telegram.ResolveBotUsername(bootstrapCtx, cfg)
	bootstrapCancel()
	if telegramUsername != "" && cfg.TelegramBotUsername == "" {
		log.Printf("telegram bot username resolved from API: @%s", telegramUsername)
	}

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
	services := service.NewService(repos, cfg)
	handlers := handler.NewHandler(healthService, services, cfg, telegramUsername)

	botCtx, botCancel := context.WithCancel(context.Background())
	defer botCancel()
	if cfg.TelegramBotPollingEnabled() {
		bot := telegram.NewBot(cfg)
		go func() {
			if err := bot.Run(botCtx); err != nil && err != context.Canceled {
				log.Printf("telegram bot stopped: %v", err)
			}
		}()
	} else {
		log.Printf("telegram bot polling disabled: configure TELEGRAM_BOT_TOKEN and TELEGRAM_PUBLIC_BASE_URL or TELEGRAM_WEBAPP_URL")
	}

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
	botCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("error while shutting down server: %s", err.Error())
	}
}
