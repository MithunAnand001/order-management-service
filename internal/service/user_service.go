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

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("Error UserService.Register.Hash", zap.String("request_id", reqID), zap.Error(err))
		return nil, dto.NewInternalError(err)
	}

	user := &models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: string(hashedPassword),
	}

	createdUser, appErr := s.repo.Create(ctx, user)
	if appErr != nil {
		s.logger.Error("Error UserService.Register.Repo", zap.String("request_id", reqID), zap.Error(appErr.Err))
		return nil, appErr
	}

	res := &dto.UserResponse{
		UUID:      createdUser.UUID.String(),
		Name:      createdUser.Name,
		Email:     createdUser.Email,
		CreatedOn: utils.FormatRFC3339(createdUser.CreatedOn),
		IsActive:  createdUser.IsActive,
	}

	s.logger.Info("End UserService.Register", zap.String("request_id", reqID))
	return res, nil
}

func (s *userSer) Login(ctx context.Context, req *dto.LoginRequest) (*dto.LoginResponse, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	s.logger.Info("Start UserService.Login", zap.String("request_id", reqID))

	user, appErr := s.repo.FindByEmail(ctx, req.Email)
	if appErr != nil {
		s.logger.Error("Error UserService.Login.Repo", zap.String("request_id", reqID), zap.Error(appErr.Err))
		return nil, appErr
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.logger.Warn("Unauthorized UserService.Login", zap.String("request_id", reqID), zap.String("email", req.Email))
		return nil, dto.NewAppError(dto.ErrCodeUnauthorized, "Invalid credentials", http.StatusUnauthorized, err)
	}

	accessToken, err := s.generateToken(user.UUID.String(), time.Minute*time.Duration(s.cfg.AccessTokenExp))
	if err != nil {
		return nil, dto.NewInternalError(err)
	}

	refreshToken, err := s.generateToken(user.UUID.String(), time.Hour*time.Duration(s.cfg.RefreshTokenExp))
	if err != nil {
		return nil, dto.NewInternalError(err)
	}

	res := &dto.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User: dto.UserResponse{
			UUID:      user.UUID.String(),
			Name:      user.Name,
			Email:     user.Email,
			CreatedOn: utils.FormatRFC3339(user.CreatedOn),
			IsActive:  user.IsActive,
		},
	}

	s.logger.Info("End UserService.Login", zap.String("request_id", reqID))
	return res, nil
}

func (s *userSer) RefreshToken(ctx context.Context, req *dto.RefreshTokenRequest) (*dto.TokenResponse, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	s.logger.Info("Start UserService.RefreshToken", zap.String("request_id", reqID))

	claims := jwt.MapClaims{}
	token, err := jwt.ParseWithClaims(req.RefreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.JWTSecret), nil
	})

	if err != nil || !token.Valid {
		s.logger.Warn("Unauthorized UserService.RefreshToken", zap.String("request_id", reqID))
		return nil, dto.NewAppError(dto.ErrCodeUnauthorized, "Invalid or expired refresh token", http.StatusUnauthorized, err)
	}

	userUUID := claims["uuid"].(string)
	user, appErr := s.repo.FindByUUID(ctx, userUUID)
	if appErr != nil {
		return nil, appErr
	}

	accessToken, err := s.generateToken(user.UUID.String(), time.Minute*time.Duration(s.cfg.AccessTokenExp))
	if err != nil {
		return nil, dto.NewInternalError(err)
	}

	var newRefreshToken string
	expTime := time.Unix(int64(claims["exp"].(float64)), 0)
	if time.Until(expTime) < 6*time.Hour {
		newRefreshToken, err = s.generateToken(user.UUID.String(), time.Hour*time.Duration(s.cfg.RefreshTokenExp))
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

func (s *userSer) generateToken(uuid string, duration time.Duration) (string, error) {
	now := utils.Now()
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"uuid": uuid,
		"exp":  now.Add(duration).Unix(),
		"iat":  now.Unix(),
	})
	return token.SignedString([]byte(s.cfg.JWTSecret))
}
