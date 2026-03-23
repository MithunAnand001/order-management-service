package dto

type ApiError struct {
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
	Code    string `json:"code"`
}

type ApiResponse[T any] struct {
	Code      int        `json:"code"`
	Message   string     `json:"message"`
	Data      T          `json:"data"`
	Timestamp string     `json:"timestamp"`
	RequestID string     `json:"request_id"`
	Errors    []ApiError `json:"errors"`
}
