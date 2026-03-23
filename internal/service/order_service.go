package service

import (
	"context"
	"fmt"
	"net/http"

	"order-management-service/internal/dto"
	"order-management-service/internal/models"
	"order-management-service/internal/repository"
	"order-management-service/internal/utils"

	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MessageBroker defines the interface for publishing events to a message queue.
type MessageBroker interface {
	PublishOrderCreated(ctx context.Context, orderUUID string) error
	Close()
}

type orderSer struct {
	orderRepo   repository.OrderRepository
	productRepo repository.ProductRepository
	broker      MessageBroker
	logger      *zap.Logger
}

type OrderService interface {
	CreateOrder(ctx context.Context, userID uint, userUUID uuid.UUID, req *dto.CreateOrderRequest) (*dto.OrderResponse, *dto.AppError)
	GetOrder(ctx context.Context, uuid uuid.UUID) (*dto.OrderResponse, *dto.AppError)
	ListOrders(ctx context.Context, userID uint, status string) ([]dto.OrderResponse, *dto.AppError)
	CancelOrder(ctx context.Context, userUUID, uuid uuid.UUID) *dto.AppError
}

func NewOrderService(orderRepo repository.OrderRepository, productRepo repository.ProductRepository, broker MessageBroker, logger *zap.Logger) OrderService {
	return &orderSer{
		orderRepo:   orderRepo,
		productRepo: productRepo,
		broker:      broker,
		logger:      logger,
	}
}

func (s *orderSer) CreateOrder(ctx context.Context, userID uint, userUUID uuid.UUID, req *dto.CreateOrderRequest) (*dto.OrderResponse, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	s.logger.Info("Start OrderService.CreateOrder", zap.String("request_id", reqID))

	var total float64
	orderItems := make([]models.OrderItem, 0)

	for _, itemReq := range req.Items {
		productUUID, err := uuid.Parse(itemReq.ProductUUID)
		if err != nil {
			return nil, dto.NewAppError(dto.ErrCodeBadRequest, "Invalid product UUID", http.StatusBadRequest, err)
		}

		product, appErr := s.productRepo.FindByUUID(ctx, productUUID)
		if appErr != nil {
			s.logger.Error("Error OrderService.CreateOrder.ProductNotFound", zap.String("request_id", reqID), zap.String("product_uuid", itemReq.ProductUUID))
			return nil, appErr
		}

		if product.StockQuantity < itemReq.Quantity {
			s.logger.Warn("BadRequest OrderService.CreateOrder.InsufficientStock", zap.String("request_id", reqID), zap.String("product", product.Name))
			return nil, dto.NewAppError(dto.ErrCodeBadRequest, fmt.Sprintf("insufficient stock for product %s", product.Name), http.StatusBadRequest, nil)
		}

		subtotal := product.CurrentPrice * float64(itemReq.Quantity)
		total += subtotal

		orderItems = append(orderItems, models.OrderItem{
			ProductID:         product.ID,
			ProductName:       product.Name,
			Quantity:          itemReq.Quantity,
			UnitPriceSnapshot: product.CurrentPrice,
			Subtotal:          subtotal,
		})
	}

	order := &models.Order{
		UserID:      userID,
		Status:      models.StatusPending,
		TotalAmount: total,
		OrderItems:  orderItems,
	}

	createdOrder, appErr := s.orderRepo.Create(ctx, order)
	if appErr != nil {
		s.logger.Error("Error OrderService.CreateOrder.Repo", zap.String("request_id", reqID), zap.Error(appErr.Err))
		return nil, appErr
	}

	if s.broker != nil {
		go func() {
			_ = s.broker.PublishOrderCreated(context.Background(), createdOrder.UUID.String())
		}()
	}

	s.logger.Info("End OrderService.CreateOrder", zap.String("request_id", reqID))
	return s.mapOrderToResponse(createdOrder), nil
}

func (s *orderSer) GetOrder(ctx context.Context, uuid uuid.UUID) (*dto.OrderResponse, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	s.logger.Info("Start OrderService.GetOrder", zap.String("request_id", reqID))

	order, appErr := s.orderRepo.FindByUUID(ctx, uuid)
	if appErr != nil {
		s.logger.Error("Error OrderService.GetOrder.Repo", zap.String("request_id", reqID), zap.Error(appErr.Err))
		return nil, appErr
	}

	s.logger.Info("End OrderService.GetOrder", zap.String("request_id", reqID))
	return s.mapOrderToResponse(order), nil
}

func (s *orderSer) ListOrders(ctx context.Context, userID uint, status string) ([]dto.OrderResponse, *dto.AppError) {
	reqID := utils.GetRequestID(ctx)
	s.logger.Info("Start OrderService.ListOrders", zap.String("request_id", reqID))

	orders, appErr := s.orderRepo.FindAll(ctx, userID, status)
	if appErr != nil {
		s.logger.Error("Error OrderService.ListOrders.Repo", zap.String("request_id", reqID), zap.Error(appErr.Err))
		return nil, appErr
	}

	res := make([]dto.OrderResponse, 0, len(orders))
	for _, o := range orders {
		res = append(res, *s.mapOrderToResponse(&o))
	}

	s.logger.Info("End OrderService.ListOrders", zap.String("request_id", reqID))
	return res, nil
}

func (s *orderSer) CancelOrder(ctx context.Context, userUUID, uuid uuid.UUID) *dto.AppError {
	reqID := utils.GetRequestID(ctx)
	s.logger.Info("Start OrderService.CancelOrder", zap.String("request_id", reqID))

	order, appErr := s.orderRepo.FindByUUID(ctx, uuid)
	if appErr != nil {
		s.logger.Error("Error OrderService.CancelOrder.RepoFind", zap.String("request_id", reqID), zap.Error(appErr.Err))
		return appErr
	}

	// Condition: Only PENDING orders can be cancelled
	if order.Status != models.StatusPending {
		s.logger.Warn("CancelDenied OrderService.CancelOrder.InvalidStatus",
			zap.String("request_id", reqID),
			zap.String("order_uuid", uuid.String()),
			zap.String("current_status", string(order.Status)))

		return dto.NewAppError(dto.ErrCodeBadRequest,
			fmt.Sprintf("order cannot be cancelled because it is already in %s status", order.Status),
			http.StatusBadRequest, nil)
	}

	log := models.OrderEventLog{
		FromStatus:  order.Status,
		ToStatus:    models.StatusCancelled,
		Reason:      "Cancelled by user",
		TriggeredBy: userUUID.String(),
	}

	appErr = s.orderRepo.UpdateStatus(ctx, order.ID, models.StatusCancelled, log)
	if appErr != nil {
		s.logger.Error("Error OrderService.CancelOrder.RepoUpdate", zap.String("request_id", reqID), zap.Error(appErr.Err))
		return appErr
	}

	s.logger.Info("End OrderService.CancelOrder", zap.String("request_id", reqID))
	return nil
}

func (s *orderSer) mapOrderToResponse(order *models.Order) *dto.OrderResponse {
	items := make([]dto.OrderItemResponse, 0)
	for _, item := range order.OrderItems {
		items = append(items, dto.OrderItemResponse{
			ProductName:       item.ProductName,
			Quantity:          item.Quantity,
			UnitPriceSnapshot: item.UnitPriceSnapshot,
			Subtotal:          item.Subtotal,
		})
	}

	events := make([]dto.OrderEventResponse, 0)
	for _, e := range order.EventLogs {
		events = append(events, dto.OrderEventResponse{
			FromStatus:  string(e.FromStatus),
			ToStatus:    string(e.ToStatus),
			Reason:      e.Reason,
			TriggeredBy: e.TriggeredBy,
			CreatedOn:   utils.FormatRFC3339(e.CreatedOn),
		})
	}

	return &dto.OrderResponse{
		UUID:        order.UUID.String(),
		Status:      order.Status,
		TotalAmount: order.TotalAmount,
		CreatedOn:   utils.FormatRFC3339(order.CreatedOn),
		Items:       items,
		Events:      events,
	}
}
