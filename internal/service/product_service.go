package service

import (
	"context"

	"order-management-service/internal/dto"
	"order-management-service/internal/repository"
	"order-management-service/internal/utils"

	"go.uber.org/zap"
)

type productSer struct {
	repo   repository.ProductRepository
	logger *zap.Logger
}

type ProductService interface {
	GetProducts(ctx context.Context, search string, limit, offset int) (*dto.ProductListResponse, *dto.AppError)
}

func NewProductService(repo repository.ProductRepository, logger *zap.Logger) ProductService {
	return &productSer{repo: repo, logger: logger}
}

func (s *productSer) GetProducts(ctx context.Context, search string, limit, offset int) (*dto.ProductListResponse, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	s.logger.Info("Start ProductService.GetProducts", zap.String("request_id", reqID))

	products, total, appErr := s.repo.FindAll(ctx, search, limit, offset)
	if appErr != nil {
		s.logger.Error("Error ProductService.GetProducts.Repo", zap.String("request_id", reqID), zap.Error(appErr.Err))
		return nil, appErr
	}

	resItems := make([]dto.ProductResponse, 0, len(products))
	for _, p := range products {
		resItems = append(resItems, dto.ProductResponse{
			UUID:           p.UUID.String(),
			SKU:            p.SKU,
			Name:           p.Name,
			Description:    p.Description,
			CurrentPrice:   p.CurrentPrice,
			StockQuantity:  p.StockQuantity,
		})
	}

	s.logger.Info("End ProductService.GetProducts", zap.String("request_id", reqID))
	return &dto.ProductListResponse{
		Items:  resItems,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}
