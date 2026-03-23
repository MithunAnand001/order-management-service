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

type ProductRepository interface {
	FindAll(ctx context.Context, search string, limit, offset int) ([]models.Product, int64, *dto.AppError)
	FindByUUID(ctx context.Context, uuid uuid.UUID) (*models.Product, *dto.AppError)
}

type productRepo struct {
	db     *gorm.DB
	logger *zap.Logger
}

func NewProductRepository(db *gorm.DB, logger *zap.Logger) ProductRepository {
	return &productRepo{db: db, logger: logger}
}

func (r *productRepo) FindAll(ctx context.Context, search string, limit, offset int) ([]models.Product, int64, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	r.logger.Info("Start ProductRepository.FindAll", zap.String("request_id", reqID))

	var products []models.Product
	var total int64
	query := r.db.WithContext(ctx).Model(&models.Product{}).Where("is_active = ?", true)

	if search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ? OR sku ILIKE ?", "%"+search+"%", "%"+search+"%", "%"+search+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		r.logger.Error("Error ProductRepository.FindAll.Count", zap.String("request_id", reqID), zap.Error(err))
		return nil, 0, dto.NewInternalError(err)
	}

	if err := query.Limit(limit).Offset(offset).Find(&products).Error; err != nil {
		r.logger.Error("Error ProductRepository.FindAll.Find", zap.String("request_id", reqID), zap.Error(err))
		return nil, 0, dto.NewInternalError(err)
	}

	r.logger.Info("End ProductRepository.FindAll", zap.String("request_id", reqID))
	return products, total, nil
}

func (r *productRepo) FindByUUID(ctx context.Context, uuid uuid.UUID) (*models.Product, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	r.logger.Info("Start ProductRepository.FindByUUID", zap.String("request_id", reqID))

	var product models.Product
	if err := r.db.WithContext(ctx).Where("uuid = ? AND is_active = ?", uuid, true).First(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("NotFound ProductRepository.FindByUUID", zap.String("request_id", reqID), zap.String("uuid", uuid.String()))
			return nil, dto.NewNotFoundError("Product not found")
		}
		r.logger.Error("Error ProductRepository.FindByUUID", zap.String("request_id", reqID), zap.Error(err))
		return nil, dto.NewInternalError(err)
	}

	r.logger.Info("End ProductRepository.FindByUUID", zap.String("request_id", reqID))
	return &product, nil
}
