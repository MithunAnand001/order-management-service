package controller

import (
	"net/http"
	"strconv"

	"order-management-service/internal/service"
	"order-management-service/internal/utils"

	"go.uber.org/zap"
)

type productCtrl struct {
	svc    service.ProductService
	logger *zap.Logger
}

type ProductController interface {
	ListProducts(w http.ResponseWriter, r *http.Request)
}

func NewProductController(svc service.ProductService, logger *zap.Logger) ProductController {
	return &productCtrl{svc: svc, logger: logger}
}

func (c *productCtrl) ListProducts(w http.ResponseWriter, r *http.Request) {
	reqID := utils.GetRequestID(r.Context())
	c.logger.Info("Start ProductController.ListProducts", zap.String("request_id", reqID), zap.String("method", r.Method))

	search := r.URL.Query().Get("search")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit, _ := strconv.Atoi(limitStr)
	if limit <= 0 {
		limit = 20
	}
	offset, _ := strconv.Atoi(offsetStr)

	res, appErr := c.svc.GetProducts(r.Context(), search, limit, offset)
	if appErr != nil {
		c.logger.Error("Error ProductController.ListProducts.Service", zap.String("request_id", reqID), zap.Error(appErr.Err))
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	c.logger.Info("End ProductController.ListProducts", zap.String("request_id", reqID), zap.String("method", r.Method))
	utils.SendJSON(w, http.StatusOK, utils.NewSuccessResponse(r.Context(), res, "Products retrieved", http.StatusOK))
}
