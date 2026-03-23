package service

import (
	"context"

	"order-management-service/internal/dto"
	"order-management-service/internal/models"
	"order-management-service/internal/repository"
	"order-management-service/internal/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AddressService interface {
	AddAddress(ctx context.Context, userID uint, req *dto.CreateAddressRequest) (*dto.AddressResponse, *dto.AppError)
	ListAddresses(ctx context.Context, userID uint) ([]dto.AddressResponse, *dto.AppError)
	SetCurrent(ctx context.Context, userID uint, addressUUID uuid.UUID) *dto.AppError
}

type addressSer struct {
	repo   repository.UserRepository
	logger *zap.Logger
}

func NewAddressService(repo repository.UserRepository, logger *zap.Logger) AddressService {
	return &addressSer{repo: repo, logger: logger}
}

func (s *addressSer) AddAddress(ctx context.Context, userID uint, req *dto.CreateAddressRequest) (*dto.AddressResponse, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	s.logger.Info("Start AddressService.AddAddress", zap.String("request_id", reqID), zap.Uint("user_id", userID))

	addr := &models.UserAddress{
		UserID:       userID,
		AddressLine1: req.AddressLine1,
		AddressLine2: req.AddressLine2,
		City:         req.City,
		State:        req.State,
		PostalCode:   req.PostalCode,
		Country:      req.Country,
		IsCurrent:    req.IsCurrent,
	}

	createdAddr, appErr := s.repo.CreateAddress(ctx, addr)
	if appErr != nil {
		return nil, appErr
	}

	s.logger.Info("End AddressService.AddAddress", zap.String("request_id", reqID))
	return &dto.AddressResponse{
		UUID:         createdAddr.UUID.String(),
		AddressLine1: createdAddr.AddressLine1,
		AddressLine2: createdAddr.AddressLine2,
		City:         createdAddr.City,
		State:        createdAddr.State,
		PostalCode:   createdAddr.PostalCode,
		Country:      createdAddr.Country,
		IsCurrent:    createdAddr.IsCurrent,
	}, nil
}

func (s *addressSer) ListAddresses(ctx context.Context, userID uint) ([]dto.AddressResponse, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	s.logger.Info("Start AddressService.ListAddresses", zap.String("request_id", reqID), zap.Uint("user_id", userID))

	addrs, appErr := s.repo.ListAddresses(ctx, userID)
	if appErr != nil {
		return nil, appErr
	}

	res := make([]dto.AddressResponse, 0, len(addrs))
	for _, a := range addrs {
		res = append(res, dto.AddressResponse{
			UUID:         a.UUID.String(),
			AddressLine1: a.AddressLine1,
			AddressLine2: a.AddressLine2,
			City:         a.City,
			State:        a.State,
			PostalCode:   a.PostalCode,
			Country:      a.Country,
			IsCurrent:    a.IsCurrent,
		})
	}

	s.logger.Info("End AddressService.ListAddresses", zap.String("request_id", reqID))
	return res, nil
}

func (s *addressSer) SetCurrent(ctx context.Context, userID uint, addressUUID uuid.UUID) *dto.AppError {
	reqID := utils.GetRequestID(ctx)
	s.logger.Info("Start AddressService.SetCurrent", zap.String("request_id", reqID), zap.Uint("user_id", userID), zap.String("address_uuid", addressUUID.String()))

	addr, appErr := s.repo.FindAddressByUUID(ctx, addressUUID)
	if appErr != nil {
		return appErr
	}

	if addr.UserID != userID {
		return dto.NewNotFoundError("Address not found for this user")
	}

	s.logger.Info("End AddressService.SetCurrent", zap.String("request_id", reqID))
	return s.repo.SetCurrentAddress(ctx, userID, addr.ID)
}
