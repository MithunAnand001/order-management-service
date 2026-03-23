package controller

import (
	"net/http"

	"order-management-service/internal/dto"
	"order-management-service/internal/service"
	"order-management-service/internal/utils"

	"go.uber.org/zap"
)

type userCtrl struct {
	svc    service.UserService
	logger *zap.Logger
}

type UserController interface {
	Register(w http.ResponseWriter, r *http.Request)
	Login(w http.ResponseWriter, r *http.Request)
	RefreshToken(w http.ResponseWriter, r *http.Request)
}

func NewUserController(svc service.UserService, logger *zap.Logger) UserController {
	return &userCtrl{svc: svc, logger: logger}
}

func (c *userCtrl) Register(w http.ResponseWriter, r *http.Request) {
	reqID := utils.GetRequestID(r.Context())
	c.logger.Info("Start UserController.Register", zap.String("request_id", reqID), zap.String("method", r.Method))

	req, appErr := utils.DecodeAndValidate[dto.CreateUserRequest](r)
	if appErr != nil {
		c.logger.Error("Error UserController.Register.Validate", zap.String("request_id", reqID), zap.Error(appErr.Err))
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	res, appErr := c.svc.Register(r.Context(), req)
	if appErr != nil {
		c.logger.Error("Error UserController.Register.Service", zap.String("request_id", reqID), zap.Error(appErr.Err))
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	c.logger.Info("End UserController.Register", zap.String("request_id", reqID), zap.String("method", r.Method))
	utils.SendJSON(w, http.StatusCreated, utils.NewSuccessResponse(r.Context(), res, "User registered successfully", http.StatusCreated))
}

func (c *userCtrl) Login(w http.ResponseWriter, r *http.Request) {
	reqID := utils.GetRequestID(r.Context())
	c.logger.Info("Start UserController.Login", zap.String("request_id", reqID), zap.String("method", r.Method))

	req, appErr := utils.DecodeAndValidate[dto.LoginRequest](r)
	if appErr != nil {
		c.logger.Error("Error UserController.Login.Validate", zap.String("request_id", reqID), zap.Error(appErr.Err))
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	res, appErr := c.svc.Login(r.Context(), req)
	if appErr != nil {
		c.logger.Error("Error UserController.Login.Service", zap.String("request_id", reqID), zap.Error(appErr.Err))
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	c.logger.Info("End UserController.Login", zap.String("request_id", reqID), zap.String("method", r.Method))
	utils.SendJSON(w, http.StatusOK, utils.NewSuccessResponse(r.Context(), res, "Login successful", http.StatusOK))
}

func (c *userCtrl) RefreshToken(w http.ResponseWriter, r *http.Request) {
	reqID := utils.GetRequestID(r.Context())
	c.logger.Info("Start UserController.RefreshToken", zap.String("request_id", reqID), zap.String("method", r.Method))

	req, appErr := utils.DecodeAndValidate[dto.RefreshTokenRequest](r)
	if appErr != nil {
		c.logger.Error("Error UserController.RefreshToken.Validate", zap.String("request_id", reqID), zap.Error(appErr.Err))
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	res, appErr := c.svc.RefreshToken(r.Context(), req)
	if appErr != nil {
		c.logger.Error("Error UserController.RefreshToken.Service", zap.String("request_id", reqID), zap.Error(appErr.Err))
		utils.SendJSON(w, appErr.HTTPStatus, utils.NewErrorResponse(r.Context(), appErr))
		return
	}

	c.logger.Info("End UserController.RefreshToken", zap.String("request_id", reqID), zap.String("method", r.Method))
	utils.SendJSON(w, http.StatusOK, utils.NewSuccessResponse(r.Context(), res, "Token refreshed successfully", http.StatusOK))
}
