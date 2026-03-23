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

type addressCtrl struct {
	svc    service.AddressService
	logger *zap.Logger
}

type AddressController interface {
	AddAddress(w http.ResponseWriter, r *http.Request)
	ListAddresses(w http.ResponseWriter, r *http.Request)
	SetCurrent(w http.ResponseWriter, r *http.Request)
}

func NewAddressController(svc service.AddressService, logger *zap.Logger) AddressController {
	return &addressCtrl{svc: svc, logger: logger}
}

func (c *addressCtrl) AddAddress(w http.ResponseWriter, r *http.Request) {
	reqID := utils.GetRequestID(r.Context())
	c.logger.Info("Start AddressController.AddAddress", zap.String("request_id", reqID))

	claims := middleware.GetClaims(r.Context())
	req, appErr := utils.DecodeAndValidate[dto.CreateAddressRequest](r)
	if appErr != nil {
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	res, appErr := c.svc.AddAddress(r.Context(), claims.UserID, req)
	if appErr != nil {
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	utils.SendJSON(w, http.StatusCreated, utils.NewSuccessResponse(r.Context(), res, "Address added successfully", http.StatusCreated))
}

func (c *addressCtrl) ListAddresses(w http.ResponseWriter, r *http.Request) {
	reqID := utils.GetRequestID(r.Context())
	c.logger.Info("Start AddressController.ListAddresses", zap.String("request_id", reqID))

	claims := middleware.GetClaims(r.Context())
	res, appErr := c.svc.ListAddresses(r.Context(), claims.UserID)
	if appErr != nil {
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	utils.SendJSON(w, http.StatusOK, utils.NewSuccessResponse(r.Context(), res, "Addresses retrieved", http.StatusOK))
}

func (c *addressCtrl) SetCurrent(w http.ResponseWriter, r *http.Request) {
	reqID := utils.GetRequestID(r.Context())
	c.logger.Info("Start AddressController.SetCurrent", zap.String("request_id", reqID))

	claims := middleware.GetClaims(r.Context())
	vars := mux.Vars(r)
	addressUUIDStr := vars["uuid"]
	addressUUID, err := uuid.Parse(addressUUIDStr)
	if err != nil {
		utils.SendJSON(w, http.StatusBadRequest, utils.NewErrorResponse(r.Context(), dto.NewAppError(dto.ErrCodeBadRequest, "Invalid address UUID", http.StatusBadRequest, err)))
		return
	}

	appErr := c.svc.SetCurrent(r.Context(), claims.UserID, addressUUID)
	if appErr != nil {
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	utils.SendJSON(w, http.StatusOK, utils.NewSuccessResponse(r.Context(), interface{}(nil), "Address set as current", http.StatusOK))
}
