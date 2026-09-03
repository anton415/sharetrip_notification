package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"

	"job4j.ru/sharetrip-notification/internal/api"
	repo "job4j.ru/sharetrip-notification/internal/repository"
	"job4j.ru/sharetrip-notification/internal/service"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	if err := run(logger); err != nil {
		logger.Error("notification service stopped", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(logger *slog.Logger) error {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		return errors.New("DATABASE_URL is required")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		return fmt.Errorf("create PostgreSQL connection pool: %w", err)
	}
	defer pool.Close()

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	err = pool.Ping(pingCtx)
	cancel()
	if err != nil {
		return fmt.Errorf("ping PostgreSQL: %w", err)
	}

	logger.InfoContext(ctx, "connected to PostgreSQL")

	notificationRepo := repo.NewPostgresNotificationRepository(pool)
	notificationService := service.NewNotificationService(notificationRepo)
	server := api.NewServer(notificationService, logger)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	server.RegisterRoutes(app)

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	logger.InfoContext(ctx, "notification service is listening", slog.String("address", addr))
	if err := app.Listen(addr); err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	return nil
}
