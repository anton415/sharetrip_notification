package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"

	"job4j.ru/sharetrip-notification/internal/api"
	"job4j.ru/sharetrip-notification/internal/events"
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

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

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
	brokers := os.Getenv("KAFKA_BROKERS")
	if brokers == "" {
		brokers = "localhost:29092"
	}
	consumer := events.NewConsumer(strings.Split(brokers, ","), "trip.events", "notification-service")
	defer func() {
		if err := consumer.Close(); err != nil {
			logger.Error("close Kafka consumer", slog.Any("error", err))
		}
	}()
	server := api.NewServer(notificationService, logger)

	app := fiber.New(fiber.Config{DisableStartupMessage: true})
	server.RegisterRoutes(app)

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8081"
	}

	listener, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", addr, err)
	}

	group, ctx := errgroup.WithContext(ctx)
	group.Go(func() error {
		if err := consumer.Run(ctx, notificationService.HandleTripPublished); err != nil {
			if errors.Is(err, context.Canceled) && ctx.Err() != nil {
				return nil
			}
			return fmt.Errorf("consume trip events: %w", err)
		}
		return nil
	})
	group.Go(func() error {
		defer stop()
		logger.InfoContext(ctx, "notification service is listening", slog.String("address", listener.Addr().String()))
		if err := app.Listener(listener); err != nil && ctx.Err() == nil {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	})
	group.Go(func() error {
		<-ctx.Done()
		err := app.ShutdownWithTimeout(5 * time.Second)
		_ = listener.Close()
		return err
	})

	return group.Wait()
}
