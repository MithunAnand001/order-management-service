package utils

import (
	"encoding/json"
	"net/http"

	"order-management-service/internal/dto"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func DecodeAndValidate[T any](r *http.Request) (*T, *dto.AppError) {
	var payload T
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		return nil, dto.NewAppError(dto.ErrCodeBadRequest, "Invalid JSON payload", http.StatusBadRequest, err)
	}

	if err := validate.Struct(payload); err != nil {
		return nil, dto.NewAppError(dto.ErrCodeValidation, "Validation failed", http.StatusBadRequest, err)
	}

	return &payload, nil
}
