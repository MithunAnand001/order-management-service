package service

import (
	"context"
	"net/http"
	"time"

	"order-management-service/internal/config"
	"order-management-service/internal/dto"
	"order-management-service/internal/models"
	"order-management-service/internal/repository"
	"order-management-service/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
)

type userSer struct {
	repo   repository.UserRepository
	cfg    *config.Config
	logger *zap.Logger
}

type UserService interface {
	Register(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, *dto.AppError)
	Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, *dto.AppError)
	RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.TokenResponse, *dto.AppError)
}

func NewUserService(repo repository.UserRepository, cfg *config.Config, logger *zap.Logger) UserService {
	return &userSer{repo: repo, cfg: cfg, logger: logger}
}

func (s *userSer) Register(ctx context.Context, req *dto.CreateUserRequest) (*dto.UserResponse, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	s.logger.Info("Start UserService.Register", zap.String("request_id", reqID))

	// 1. Hash Password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, dto.NewInternalError(err)
	}

	// 2. Create User with Role Enum
	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
		Role:     models.UserRole(req.RoleCode),
	}

	createdUser, appErr := s.repo.Create(ctx, user)
	if appErr != nil {
		return nil, appErr
	}

	// 3. Create Initial Address if provided
	var addressRes []dto.AddressResponse
	if req.Address != nil {
		addr := &models.UserAddress{
			UserID:       createdUser.ID,
			AddressLine1: req.Address.AddressLine1,
			AddressLine2: req.Address.AddressLine2,
			City:         req.Address.City,
			State:        req.Address.State,
			PostalCode:   req.Address.PostalCode,
			Country:      req.Address.Country,
			IsCurrent:    true,
		}
		createdAddr, appErr := s.repo.CreateAddress(ctx, addr)
		if appErr == nil {
			addressRes = append(addressRes, dto.AddressResponse{
				UUID:         createdAddr.UUID.String(),
				AddressLine1: createdAddr.AddressLine1,
				AddressLine2: createdAddr.AddressLine2,
				City:         createdAddr.City,
				State:        createdAddr.State,
				PostalCode:   createdAddr.PostalCode,
				Country:      createdAddr.Country,
				IsCurrent:    createdAddr.IsCurrent,
			})
		}
	}

	s.logger.Info("End UserService.Register", zap.String("request_id", reqID))
	return &dto.UserResponse{
		UUID:      createdUser.UUID.String(),
		Name:      createdUser.Name,
		Email:     createdUser.Email,
		Role:      string(createdUser.Role),
		CreatedOn: utils.FormatRFC3339(createdUser.CreatedOn),
		IsActive:  createdUser.IsActive,
		Addresses: addressRes,
	}, nil
}

func (s *userSer) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	s.logger.Info("Start UserService.Login", zap.String("request_id", reqID))

	user, appErr := s.repo.FindByEmail(ctx, req.Email)
	if appErr != nil {
		return nil, appErr
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, dto.NewAppError(dto.ErrCodeUnauthorized, "Invalid credentials", http.StatusUnauthorized, err)
	}

	accessToken, err := s.generateToken(user.UUID, time.Minute*time.Duration(s.cfg.AccessTokenExp))
	if err != nil {
		return nil, dto.NewInternalError(err)
	}

	refreshToken, err := s.generateToken(user.UUID, time.Hour*time.Duration(s.cfg.RefreshTokenExp))
	if err != nil {
		return nil, dto.NewInternalError(err)
	}

	s.logger.Info("End UserService.Login", zap.String("request_id", reqID))
	return &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserResponse{
			UUID:      user.UUID.String(),
			Name:      user.Name,
			Email:     user.Email,
			Role:      string(user.Role),
			CreatedOn: utils.FormatRFC3339(user.CreatedOn),
			IsActive:  user.IsActive,
		},
	}, nil
}

func (s *userSer) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.TokenResponse, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	s.logger.Info("Start UserService.RefreshToken", zap.String("request_id", reqID))

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		return nil, dto.NewAppError(dto.ErrCodeUnauthorized, "Invalid or expired refresh token", http.StatusUnauthorized, err)
	}

	userUUIDStr := claims["uuid"].(string)
	userUUID, err := uuid.Parse(userUUIDStr)
	if err != nil {
		return nil, dto.NewAppError(dto.ErrCodeUnauthorized, "Invalid token claims", http.StatusUnauthorized, err)
	}

	user, appErr := s.repo.FindByUUID(ctx, userUUID)
	if appErr != nil {
		return nil, appErr
	}

	accessToken, err := s.generateToken(user.UUID, time.Minute*time.Duration(s.cfg.AccessTokenExp))
	if err != nil {
		return nil, dto.NewInternalError(err)
	}

	var newRefreshToken string
	expTime := time.Unix(int64(claims["exp"].(float64)), 0)
	if time.Until(expTime) < 6*time.Hour {
		newRefreshToken, err = s.generateToken(user.UUID, time.Hour*time.Duration(s.cfg.RefreshTokenExp))
		if err != nil {
			return nil, dto.NewInternalError(err)
		}
	}

	s.logger.Info("End UserService.RefreshToken", zap.String("request_id", reqID))
	return &dto.TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func (s *userSer) generateToken(userUUID uuid.UUID, duration time.Duration) (string, error) {
	now := utils.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uuid": userUUID.String(),
		"exp":  now.Add(duration).Unix(),
		"iat":  now.Unix(),
	})
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
