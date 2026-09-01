package api

import "job4j.ru/sharetrip-notification/internal/service"

type Server struct {
	notifications *service.NotificationService
}

func NewServer(notifications *service.NotificationService) *Server {
	return &Server{
		notifications: notifications,
	}
}
