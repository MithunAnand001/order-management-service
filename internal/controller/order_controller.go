package controller

import (
	"net/http"

	"order-management-service/internal/dto"
	"order-management-service/internal/middleware"
	"order-management-service/internal/service"
	"order-management-service/internal/utils"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type orderCtrl struct {
	svc    service.OrderService
	logger *zap.Logger
}

type OrderController interface {
	CreateOrder(w http.ResponseWriter, r *http.Request)
	GetOrder(w http.ResponseWriter, r *http.Request)
	ListOrders(w http.ResponseWriter, r *http.Request)
	CancelOrder(w http.ResponseWriter, r *http.Request)
	UpdateStatus(w http.ResponseWriter, r *http.Request)
}

func NewOrderController(svc service.OrderService, logger *zap.Logger) OrderController {
	return &orderCtrl{svc: svc, logger: logger}
}

func (c *orderCtrl) CreateOrder(w http.ResponseWriter, r *http.Request) {
	reqID := utils.GetRequestID(r.Context())
	c.logger.Info("Start OrderController.CreateOrder", zap.String("request_id", reqID), zap.String("method", r.Method))

	claims := middleware.GetClaims(r.Context())
	if claims == nil {
		c.logger.Warn("Unauthorized OrderController.CreateOrder", zap.String("request_id", reqID))
		utils.SendJSON(w, http.StatusUnauthorized, utils.NewErrorResponse(r.Context(), dto.NewAppError(dto.ErrCodeUnauthorized, "Unauthorized", http.StatusUnauthorized, nil)))
		return
	}

	req, appErr := utils.DecodeAndValidate[dto.CreateOrderRequest](r)
	if appErr != nil {
		c.logger.Error("Error OrderController.CreateOrder.Validate", zap.String("request_id", reqID), zap.Error(appErr.Err))
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	res, appErr := c.svc.CreateOrder(r.Context(), claims.UserID, claims.UUID, req)
	if appErr != nil {
		c.logger.Error("Error OrderController.CreateOrder.Service", zap.String("request_id", reqID), zap.Error(appErr.Err))
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	c.logger.Info("End OrderController.CreateOrder", zap.String("request_id", reqID), zap.String("method", r.Method))
	utils.SendJSON(w, http.StatusCreated, utils.NewSuccessResponse(r.Context(), res, "Order created successfully", http.StatusCreated))
}

func (c *orderCtrl) GetOrder(w http.ResponseWriter, r *http.Request) {
	reqID := utils.GetRequestID(r.Context())
	c.logger.Info("Start OrderController.GetOrder", zap.String("request_id", reqID), zap.String("method", r.Method))

	vars := mux.Vars(r)
	orderUUIDStr := vars["uuid"]
	orderUUID, err := uuid.Parse(orderUUIDStr)
	if err != nil {
		c.logger.Error("Error OrderController.GetOrder.ParseUUID", zap.String("request_id", reqID), zap.Error(err))
		utils.SendJSON(w, http.StatusBadRequest, utils.NewErrorResponse(r.Context(), dto.NewAppError(dto.ErrCodeBadRequest, "Invalid order UUID", http.StatusBadRequest, err)))
		return
	}

	res, appErr := c.svc.GetOrder(r.Context(), orderUUID)
	if appErr != nil {
		c.logger.Error("Error OrderController.GetOrder.Service", zap.String("request_id", reqID), zap.Error(appErr.Err))
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	c.logger.Info("End OrderController.GetOrder", zap.String("request_id", reqID), zap.String("method", r.Method))
	utils.SendJSON(w, http.StatusOK, utils.NewSuccessResponse(r.Context(), res, "Order retrieved", http.StatusOK))
}

func (c *orderCtrl) ListOrders(w http.ResponseWriter, r *http.Request) {
	reqID := utils.GetRequestID(r.Context())
	c.logger.Info("Start OrderController.ListOrders", zap.String("request_id", reqID), zap.String("method", r.Method))

	claims := middleware.GetClaims(r.Context())
	status := r.URL.Query().Get("status")

	res, appErr := c.svc.ListOrders(r.Context(), claims.UserID, status)
	if appErr != nil {
		c.logger.Error("Error OrderController.ListOrders.Service", zap.String("request_id", reqID), zap.Error(appErr.Err))
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	c.logger.Info("End OrderController.ListOrders", zap.String("request_id", reqID), zap.String("method", r.Method))
	utils.SendJSON(w, http.StatusOK, utils.NewSuccessResponse(r.Context(), res, "Orders retrieved", http.StatusOK))
}

func (c *orderCtrl) CancelOrder(w http.ResponseWriter, r *http.Request) {
	reqID := utils.GetRequestID(r.Context())
	c.logger.Info("Start OrderController.CancelOrder", zap.String("request_id", reqID), zap.String("method", r.Method))

	claims := middleware.GetClaims(r.Context())
	vars := mux.Vars(r)
	orderUUIDStr := vars["uuid"]
	orderUUID, err := uuid.Parse(orderUUIDStr)
	if err != nil {
		c.logger.Error("Error OrderController.CancelOrder.ParseUUID", zap.String("request_id", reqID), zap.Error(err))
		utils.SendJSON(w, http.StatusBadRequest, utils.NewErrorResponse(r.Context(), dto.NewAppError(dto.ErrCodeBadRequest, "Invalid order UUID", http.StatusBadRequest, err)))
		return
	}

	appErr := c.svc.CancelOrder(r.Context(), claims.UUID, orderUUID)
	if appErr != nil {
		c.logger.Error("Error OrderController.CancelOrder.Service", zap.String("request_id", reqID), zap.Error(appErr.Err))
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	c.logger.Info("End OrderController.CancelOrder", zap.String("request_id", reqID), zap.String("method", r.Method))
	utils.SendJSON(w, http.StatusOK, utils.NewSuccessResponse(r.Context(), interface{}(nil), "Order cancelled successfully", http.StatusOK))
}

func (c *orderCtrl) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	reqID := utils.GetRequestID(r.Context())
	c.logger.Info("Start OrderController.UpdateStatus", zap.String("request_id", reqID), zap.String("method", r.Method))

	claims := middleware.GetClaims(r.Context())
	vars := mux.Vars(r)
	orderUUIDStr := vars["uuid"]
	orderUUID, err := uuid.Parse(orderUUIDStr)
	if err != nil {
		c.logger.Error("Error OrderController.UpdateStatus.ParseUUID", zap.String("request_id", reqID), zap.Error(err))
		utils.SendJSON(w, http.StatusBadRequest, utils.NewErrorResponse(r.Context(), dto.NewAppError(dto.ErrCodeBadRequest, "Invalid order UUID", http.StatusBadRequest, err)))
		return
	}

	req, appErr := utils.DecodeAndValidate[dto.UpdateOrderStatusRequest](r)
	if appErr != nil {
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	appErr = c.svc.UpdateOrderStatus(r.Context(), claims, orderUUID, req)
	if appErr != nil {
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	c.logger.Info("End OrderController.UpdateStatus", zap.String("request_id", reqID), zap.String("method", r.Method))
	utils.SendJSON(w, http.StatusOK, utils.NewSuccessResponse(r.Context(), interface{}(nil), "Order status updated", http.StatusOK))
}
