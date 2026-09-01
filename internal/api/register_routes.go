package api

import "github.com/gofiber/fiber/v2"

func (s *Server) RegisterRoutes(router fiber.Router) {
	router.Post("/notifications", s.createNotification)
	router.Get("/notifications/:id", s.getNotificationByID)
}
