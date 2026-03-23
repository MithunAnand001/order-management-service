package repository

import (
	"context"
	"errors"
	"net/http"

	"order-management-service/internal/dto"
	"order-management-service/internal/models"
	"order-management-service/internal/utils"

	"go.uber.org/zap"
	"gorm.io/gorm"
)

type userRepo struct {
	db     *gorm.DB
	logger *zap.Logger
}

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, *dto.AppError)
	FindByEmail(ctx context.Context, email string) (*models.User, *dto.AppError)
	FindByUUID(ctx context.Context, uuid string) (*models.User, *dto.AppError)
}

func NewUserRepository(db *gorm.DB, logger *zap.Logger) UserRepository {
	return &userRepo{db: db, logger: logger}
}

func (r *userRepo) Create(ctx context.Context, user *models.User) (*models.User, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	r.logger.Info("Start UserRepository.Create", zap.String("request_id", reqID))

	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		r.logger.Error("Error UserRepository.Create", zap.String("request_id", reqID), zap.Error(err))
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return nil, dto.NewAppError(dto.ErrCodeConflict, "User already exists", http.StatusConflict, err)
		}
		return nil, dto.NewInternalError(err)
	}

	r.logger.Info("End UserRepository.Create", zap.String("request_id", reqID))
	return user, nil
}

func (r *userRepo) FindByEmail(ctx context.Context, email string) (*models.User, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	r.logger.Info("Start UserRepository.FindByEmail", zap.String("request_id", reqID))

	var user models.User
	if err := r.db.WithContext(ctx).Where("email = ? AND is_active = ?", email, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("NotFound UserRepository.FindByEmail", zap.String("request_id", reqID), zap.String("email", email))
			return nil, dto.NewNotFoundError("User not found")
		}
		r.logger.Error("Error UserRepository.FindByEmail", zap.String("request_id", reqID), zap.Error(err))
		return nil, dto.NewInternalError(err)
	}

	r.logger.Info("End UserRepository.FindByEmail", zap.String("request_id", reqID))
	return &user, nil
}

func (r *userRepo) FindByUUID(ctx context.Context, uuid string) (*models.User, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	r.logger.Info("Start UserRepository.FindByUUID", zap.String("request_id", reqID))

	var user models.User
	if err := r.db.WithContext(ctx).Where("uuid = ? AND is_active = ?", uuid, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			r.logger.Warn("NotFound UserRepository.FindByUUID", zap.String("request_id", reqID), zap.String("uuid", uuid))
			return nil, dto.NewNotFoundError("User not found")
		}
		r.logger.Error("Error UserRepository.FindByUUID", zap.String("request_id", reqID), zap.Error(err))
		return nil, dto.NewInternalError(err)
	}

	r.logger.Info("End UserRepository.FindByUUID", zap.String("request_id", reqID))
	return &user, nil
}
