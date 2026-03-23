package repository

import (
	"context"
	"errors"
	"net/http"

	"order-management-service/internal/dto"
	"order-management-service/internal/models"
	"order-management-service/internal/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, *dto.AppError)
	FindByEmail(ctx context.Context, email string) (*models.User, *dto.AppError)
	FindByUUID(ctx context.Context, uuid uuid.UUID) (*models.User, *dto.AppError)
	FindByID(ctx context.Context, id uint) (*models.User, *dto.AppError)
	
	// Address Methods
	CreateAddress(ctx context.Context, addr *models.UserAddress) (*models.UserAddress, *dto.AppError)
	ListAddresses(ctx context.Context, userID uint) ([]models.UserAddress, *dto.AppError)
	FindAddressByUUID(ctx context.Context, uuid uuid.UUID) (*models.UserAddress, *dto.AppError)
	SetCurrentAddress(ctx context.Context, userID uint, addressID uint) *dto.AppError
}

type userRepo struct {
	db     *gorm.DB
	logger *zap.Logger
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
			return nil, dto.NewNotFoundError("User not found")
		}
		return nil, dto.NewInternalError(err)
	}

	return &user, nil
}

func (r *userRepo) FindByUUID(ctx context.Context, uuid uuid.UUID) (*models.User, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	r.logger.Info("Start UserRepository.FindByUUID", zap.String("request_id", reqID))

	var user models.User
	if err := r.db.WithContext(ctx).Where("uuid = ? AND is_active = ?", uuid, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dto.NewNotFoundError("User not found")
		}
		return nil, dto.NewInternalError(err)
	}

	return &user, nil
}

func (r *userRepo) FindByID(ctx context.Context, id uint) (*models.User, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	r.logger.Info("Start UserRepository.FindByID", zap.String("request_id", reqID))

	var user models.User
	if err := r.db.WithContext(ctx).Where("id = ? AND is_active = ?", id, true).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dto.NewNotFoundError("User not found")
		}
		return nil, dto.NewInternalError(err)
	}

	return &user, nil
}

func (r *userRepo) CreateAddress(ctx context.Context, addr *models.UserAddress) (*models.UserAddress, *dto.AppError) {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if addr.IsCurrent {
			if err := tx.Model(&models.UserAddress{}).Where("user_id = ?", addr.UserID).Update("is_current", false).Error; err != nil {
				return err
			}
		}
		return tx.Create(addr).Error
	})
	if err != nil {
		return nil, dto.NewInternalError(err)
	}
	return addr, nil
}

func (r *userRepo) ListAddresses(ctx context.Context, userID uint) ([]models.UserAddress, *dto.AppError) {
	var addrs []models.UserAddress
	if err := r.db.WithContext(ctx).Where("user_id = ? AND is_active = ?", userID, true).Order("is_current DESC, created_on DESC").Find(&addrs).Error; err != nil {
		return nil, dto.NewInternalError(err)
	}
	return addrs, nil
}

func (r *userRepo) FindAddressByUUID(ctx context.Context, uuid uuid.UUID) (*models.UserAddress, *dto.AppError) {
	var addr models.UserAddress
	if err := r.db.WithContext(ctx).Where("uuid = ? AND is_active = ?", uuid, true).First(&addr).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, dto.NewNotFoundError("Address not found")
		}
		return nil, dto.NewInternalError(err)
	}
	return &addr, nil
}

func (r *userRepo) SetCurrentAddress(ctx context.Context, userID uint, addressID uint) *dto.AppError {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Model(&models.UserAddress{}).Where("user_id = ?", userID).Update("is_current", false).Error; err != nil {
			return err
		}
		if err := tx.Model(&models.UserAddress{}).Where("id = ?", addressID).Update("is_current", true).Error; err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return dto.NewInternalError(err)
	}
	return nil
}
