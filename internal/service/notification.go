package service

import repo "job4j.ru/sharetrip-notification/internal/repository"

type NotificationService struct {
	repo *repo.PostgresNotificationRepository
}

func NewNotificationService(notificationRepo *repo.PostgresNotificationRepository) *NotificationService {
	return &NotificationService{
		repo: notificationRepo,
	}
}
