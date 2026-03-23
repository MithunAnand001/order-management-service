package dto

import "order-management-service/internal/models"

type CreateOrderRequest struct {
	Items []OrderItemRequest `json:"items" validate:"required,min=1,dive"`
}

type OrderItemRequest struct {
	ProductUUID string `json:"product_uuid" validate:"required,uuid"`
	Quantity    int    `json:"quantity" validate:"required,min=1"`
}

type OrderResponse struct {
	UUID        string               `json:"uuid"`
	Status      models.OrderStatus   `json:"status"`
	TotalAmount float64              `json:"total_amount"`
	CreatedOn   string               `json:"created_on"`
	Items       []OrderItemResponse  `json:"items,omitempty"`
	Events      []OrderEventResponse `json:"events,omitempty"`
}

type OrderItemResponse struct {
	ProductUUID       string  `json:"product_uuid"`
	ProductName       string  `json:"product_name"`
	Quantity          int     `json:"quantity"`
	UnitPriceSnapshot float64 `json:"unit_price_snapshot"`
	Subtotal          float64 `json:"subtotal"`
}

type OrderEventResponse struct {
	FromStatus  string `json:"from_status"`
	ToStatus    string `json:"to_status"`
	Reason      string `json:"reason"`
	TriggeredBy string `json:"triggered_by"`
	CreatedOn   string `json:"created_on"`
}

type ProductResponse struct {
	UUID           string  `json:"uuid"`
	SKU            string  `json:"sku"`
	Name           string  `json:"name"`
	Description    string  `json:"description"`
	CurrentPrice   float64 `json:"current_price"`
	StockQuantity  int     `json:"stock_quantity"`
}

type ProductListResponse struct {
	Items  []ProductResponse `json:"items"`
	Total  int64             `json:"total"`
	Limit  int               `json:"limit"`
	Offset int               `json:"offset"`
}
