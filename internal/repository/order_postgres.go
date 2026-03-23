package repository

import (
	"context"
	"errors"

	"order-management-service/internal/dto"
	"order-management-service/internal/models"
	"order-management-service/internal/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type orderRepo struct {
	db     *gorm.DB
	logger *zap.Logger
}

type OrderRepository interface {
	Create(ctx context.Context, order *models.Order) (*models.Order, *dto.AppError)
	FindByID(ctx context.Context, id uint) (*models.Order, *dto.AppError)
	FindByUUID(ctx context.Context, uuid uuid.UUID) (*models.Order, *dto.AppError)
	FindAll(ctx context.Context, userID uint, status string) ([]models.Order, *dto.AppError)
	UpdateStatus(ctx context.Context, id uint, status models.OrderStatus, log models.OrderEventLog) *dto.AppError
	FindPendingOrders(ctx context.Context) ([]models.Order, *dto.AppError)
}

func NewOrderRepository(db *gorm.DB, logger *zap.Logger) OrderRepository {
	return &orderRepo{db: db, logger: logger}
}

func (r *orderRepo) Create(ctx context.Context, order *models.Order) (*models.Order, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	r.logger.Info("Start OrderRepository.Create", zap.String("request_id", reqID))

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		log := models.OrderEventLog{
			OrderID:     order.ID,
			ToStatus:    models.StatusPending,
			Reason:      "Order created",
			TriggeredBy: "SYSTEM",
		}
		if err := tx.Create(&log).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		r.logger.Error("Error OrderRepository.Create", zap.String("request_id", reqID), zap.Error(err))
		return nil, dto.NewInternalError(err)
	}

	r.logger.Info("End OrderRepository.Create", zap.String("request_id", reqID))
	return order, nil
}

func (r *orderRepo) FindByID(ctx context.Context, id uint) (*models.Order, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	r.logger.Info("Start OrderRepository.FindByID", zap.String("request_id", reqID))

	var order models.Order
	if err := r.db.WithContext(ctx).Preload("OrderItems").Preload("EventLogs").First(&order, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("NotFound OrderRepository.FindByID", zap.String("request_id", reqID), zap.Uint("id", id))
			return nil, dto.NewNotFoundError("Order not found")
		}
		r.logger.Error("Error OrderRepository.FindByID", zap.String("request_id", reqID), zap.Error(err))
		return nil, dto.NewInternalError(err)
	}

	r.logger.Info("End OrderRepository.FindByID", zap.String("request_id", reqID))
	return &order, nil
}

func (r *orderRepo) FindByUUID(ctx context.Context, uuid uuid.UUID) (*models.Order, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	r.logger.Info("Start OrderRepository.FindByUUID", zap.String("request_id", reqID))

	var order models.Order
	if err := r.db.WithContext(ctx).
		Preload("OrderItems").
		Preload("EventLogs").
		Where("uuid = ?", uuid).
		First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("NotFound OrderRepository.FindByUUID", zap.String("request_id", reqID), zap.String("uuid", uuid.String()))
			return nil, dto.NewNotFoundError("Order not found")
		}
		r.logger.Error("Error OrderRepository.FindByUUID", zap.String("request_id", reqID), zap.Error(err))
		return nil, dto.NewInternalError(err)
	}

	r.logger.Info("End OrderRepository.FindByUUID", zap.String("request_id", reqID))
	return &order, nil
}

func (r *orderRepo) FindAll(ctx context.Context, userID uint, status string) ([]models.Order, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	r.logger.Info("Start OrderRepository.FindAll", zap.String("request_id", reqID))

	var orders []models.Order
	query := r.db.WithContext(ctx).Where("user_id = ?", userID)
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if err := query.Find(&orders).Error; err != nil {
		r.logger.Error("Error OrderRepository.FindAll", zap.String("request_id", reqID), zap.Error(err))
		return nil, dto.NewInternalError(err)
	}

	r.logger.Info("End OrderRepository.FindAll", zap.String("request_id", reqID))
	return orders, nil
}

func (r *orderRepo) UpdateStatus(ctx context.Context, id uint, status models.OrderStatus, log models.OrderEventLog) *dto.AppError {
	reqID := utils.GetRequestID(ctx)
	r.logger.Info("Start OrderRepository.UpdateStatus", zap.String("request_id", reqID))

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.Order{}).Where("id = ?", id).Update("status", status).Error; err != nil {
			return err
		}
		log.OrderID = id
		if err := tx.Create(&log).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		r.logger.Error("Error OrderRepository.UpdateStatus", zap.String("request_id", reqID), zap.Error(err))
		return dto.NewInternalError(err)
	}

	r.logger.Info("End OrderRepository.UpdateStatus", zap.String("request_id", reqID))
	return nil
}

func (r *orderRepo) FindPendingOrders(ctx context.Context) ([]models.Order, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	r.logger.Info("Start OrderRepository.FindPendingOrders", zap.String("request_id", reqID))

	var orders []models.Order
	if err := r.db.WithContext(ctx).Where("status = ?", models.StatusPending).Find(&orders).Error; err != nil {
		r.logger.Error("Error OrderRepository.FindPendingOrders", zap.String("request_id", reqID), zap.Error(err))
		return nil, dto.NewInternalError(err)
	}

	r.logger.Info("End OrderRepository.FindPendingOrders", zap.String("request_id", reqID))
	return orders, nil
}
