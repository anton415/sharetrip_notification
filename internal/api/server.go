package api

import (
	"log/slog"

	"job4j.ru/sharetrip-notification/internal/service"
)

type Server struct {
	notifications *service.NotificationService
	logger        *slog.Logger
}

func NewServer(notifications *service.NotificationService, logger *slog.Logger) *Server {
	return &Server{
		notifications: notifications,
		logger:        logger,
	}
}
