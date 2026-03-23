package service

import (
	"context"
	"fmt"

	"order-management-service/internal/repository"
	"order-management-service/internal/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type OrderActivityService interface {
	HandleOrderCreatedActivity(ctx context.Context, orderUUID uuid.UUID) error
}

type orderActivitySer struct {
	userRepo  repository.UserRepository
	orderRepo repository.OrderRepository
	commSvc   CommunicationService
	logger    *zap.Logger
}

func NewOrderActivityService(userRepo repository.UserRepository, orderRepo repository.OrderRepository, commSvc CommunicationService, logger *zap.Logger) OrderActivityService {
	return &orderActivitySer{
		userRepo:  userRepo,
		orderRepo: orderRepo,
		commSvc:   commSvc,
		logger:    logger,
	}
}

func (s *orderActivitySer) HandleOrderCreatedActivity(ctx context.Context, orderUUID uuid.UUID) error {
	reqID := utils.GetRequestID(ctx)
	s.logger.Info("Start OrderActivityService.HandleOrderCreatedActivity", zap.String("request_id", reqID), zap.String("order_uuid", orderUUID.String()))

	// 1. Fetch Order Details
	order, appErr := s.orderRepo.FindByUUID(ctx, orderUUID)
	if appErr != nil {
		s.logger.Error("Error OrderActivityService.HandleOrderCreatedActivity.OrderFind", zap.String("request_id", reqID), zap.Error(appErr.Err))
		return fmt.Errorf("failed to find order: %v", appErr.Err)
	}

	// 2. Fetch User Details using order.UserID
	user, appErr := s.userRepo.FindByID(ctx, order.UserID)
	if appErr != nil {
		s.logger.Error("Error OrderActivityService.HandleOrderCreatedActivity.UserFind", zap.String("request_id", reqID), zap.Error(appErr.Err))
		return fmt.Errorf("failed to find user: %v", appErr.Err)
	}

	// 3. Delegate to Communication Service
	err := s.commSvc.SendOrderConfirmationEmail(ctx, user.Name, user.Email, orderUUID.String(), order.TotalAmount)
	if err != nil {
		return err // Retry will be handled by the consumer
	}

	s.logger.Info("End OrderActivityService.HandleOrderCreatedActivity", zap.String("request_id", reqID))
	return nil
}
