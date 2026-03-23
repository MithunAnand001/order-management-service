package repository

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"order-management-service/internal/dto"
	"order-management-service/internal/models"
	"order-management-service/internal/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type OrderRepository interface {
	CreateWithStock(ctx context.Context, order *models.Order) (*models.Order, *dto.AppError)
	FindByID(ctx context.Context, id uint) (*models.Order, *dto.AppError)
	FindByUUID(ctx context.Context, uuid uuid.UUID) (*models.Order, *dto.AppError)
	FindAll(ctx context.Context, userID uint, status string) ([]models.Order, *dto.AppError)
	UpdateStatusWithStock(ctx context.Context, id uint, status models.OrderStatus, log models.OrderEventLog) *dto.AppError
}

type orderRepo struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewOrderRepository(db *gorm.DB, logger *zap.Logger) OrderRepository {
	return &orderRepo{db: db, logger: logger}
}

func (r *orderRepo) CreateWithStock(ctx context.Context, order *models.Order) (*models.Order, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	r.logger.Info("Start OrderRepository.CreateWithStock", zap.String("request_id", reqID))

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Deduct Stock for each item
		for _, item := range order.OrderItems {
			result := tx.Model(&models.Product{}).
				Where("id = ? AND stock_quantity >= ?", item.ProductID, item.Quantity).
				Update("stock_quantity", gorm.Expr("stock_quantity - ?", item.Quantity))

			if result.Error != nil {
				return result.Error
			}
			if result.RowsAffected == 0 {
				return fmt.Errorf("insufficient stock for product ID %d", item.ProductID)
			}
		}

		// 2. Create Order
		if err := tx.Create(order).Error; err != nil {
			return err
		}

		// 3. Initial Event Log
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
		r.logger.Error("Error OrderRepository.CreateWithStock", zap.String("request_id", reqID), zap.Error(err))
		if err.Error()[:12] == "insufficient" {
			return nil, dto.NewAppError(dto.ErrCodeBadRequest, err.Error(), http.StatusBadRequest, err)
		}
		return nil, dto.NewInternalError(err)
	}

	r.logger.Info("End OrderRepository.CreateWithStock", zap.String("request_id", reqID))
	return order, nil
}

func (r *orderRepo) UpdateStatusWithStock(ctx context.Context, id uint, status models.OrderStatus, log models.OrderEventLog) *dto.AppError {
	reqID := utils.GetRequestID(ctx)
	r.logger.Info("Start OrderRepository.UpdateStatusWithStock", zap.String("request_id", reqID))

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Update status
		if err := tx.Model(&models.Order{}).Where("id = ?", id).Update("status", status).Error; err != nil {
			return err
		}

		// 2. If CANCELLED, replenish stock
		if status == models.StatusCancelled {
			var items []models.OrderItem
			if err := tx.Where("order_id = ?", id).Find(&items).Error; err != nil {
				return err
			}
			for _, item := range items {
				if err := tx.Model(&models.Product{}).Where("id = ?", item.ProductID).
					Update("stock_quantity", gorm.Expr("stock_quantity + ?", item.Quantity)).Error; err != nil {
					return err
				}
			}
		}

		// 3. Log event
		log.OrderID = id
		if err := tx.Create(&log).Error; err != nil {
			return err
		}
		return nil
	})

	if err != nil {
		r.logger.Error("Error OrderRepository.UpdateStatusWithStock", zap.String("request_id", reqID), zap.Error(err))
		return dto.NewInternalError(err)
	}

	r.logger.Info("End OrderRepository.UpdateStatusWithStock", zap.String("request_id", reqID))
	return nil
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
	if err := r.db.WithContext(ctx).Preload("OrderItems").Preload("EventLogs").Where("uuid = ?", uuid).First(&order).Error; err != nil {
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
