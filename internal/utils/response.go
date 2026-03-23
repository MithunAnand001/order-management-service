package utils

import (
	"context"
	"encoding/json"
	"net/http"

	"order-management-service/internal/dto"
)

type contextKey string

const RequestIDKey contextKey = "request_id"

func GetRequestID(ctx context.Context) string {
	if reqID, ok := ctx.Value(RequestIDKey).(string); ok {
		return reqID
	}
	return ""
}

func SendJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(payload)
}

func NewSuccessResponse[T any](ctx context.Context, data T, message string, code int) dto.ApiResponse[T] {
	return dto.ApiResponse[T]{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: FormatRFC3339(Now()),
		RequestID: GetRequestID(ctx),
		Errors:    []dto.ApiError{},
	}
}

func NewErrorResponse(ctx context.Context, appErr *dto.AppError) dto.ApiResponse[interface{}] {
	return dto.ApiResponse[interface{}]{
		Code:      appErr.HTTPStatus,
		Message:   appErr.Message,
		Data:      nil,
		Timestamp: FormatRFC3339(Now()),
		RequestID: GetRequestID(ctx),
		Errors: []dto.ApiError{
			{
				Message: appErr.Message,
				Code:    string(appErr.Code),
			},
		},
	}
}

func NewApiResponse[T any](ctx context.Context, code int, message string, data T, errs []dto.ApiError) dto.ApiResponse[T] {
	return dto.ApiResponse[T]{
		Code:      code,
		Message:   message,
		Data:      data,
		Timestamp: FormatRFC3339(Now()),
		RequestID: GetRequestID(ctx),
		Errors:    errs,
	}
}
