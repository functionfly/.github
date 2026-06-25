package employee

import (
	"github.com/functionfly/functionfly/internal/api/handlers/chat"
	"github.com/functionfly/functionfly/internal/notification"
	"github.com/functionfly/functionfly/internal/storage"
	"github.com/sirupsen/logrus"
)

type Handler struct {
	repo            storage.Repository
	notificationSvc *notification.Service
	aiClient        *chat.AIServiceClient
	log             *logrus.Logger
}

func NewHandler(repo storage.Repository, notificationSvc *notification.Service, log *logrus.Logger) *Handler {
	return &Handler{
		repo:            repo,
		notificationSvc: notificationSvc,
		log:             log,
	}
}

func NewHandlerWithAI(repo storage.Repository, notificationSvc *notification.Service, aiClient *chat.AIServiceClient, log *logrus.Logger) *Handler {
	return &Handler{
		repo:            repo,
		notificationSvc: notificationSvc,
		aiClient:        aiClient,
		log:             log,
	}
}
