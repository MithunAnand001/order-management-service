package models

type OrderStatus string

const (
	StatusPending        OrderStatus = "PENDING"
	StatusProcessing     OrderStatus = "PROCESSING"
	StatusOutForDelivery OrderStatus = "OUT_FOR_DELIVERY"
	StatusShipped        OrderStatus = "SHIPPED"
	StatusDelivered      OrderStatus = "DELIVERED"
	StatusCancelled      OrderStatus = "CANCELLED"
)

type UserRole string

const (
	RoleAdmin    UserRole = "ADMIN"
	RoleUser     UserRole = "USER"
	RoleDelivery UserRole = "DELIVERY"
)

type User struct {
	Base
	Role      UserRole      `gorm:"type:text;not null;default:'USER'" json:"role"`
	Name      string        `gorm:"not null" json:"name"`
	Email     string        `gorm:"uniqueIndex;not null" json:"email"`
	Password  string        `gorm:"not null" json:"-"`
	Addresses []UserAddress `gorm:"foreignKey:UserID" json:"addresses,omitempty"`
	Orders    []Order       `gorm:"foreignKey:UserID" json:"orders,omitempty"`
}

type UserAddress struct {
	Base
	UserID       uint   `gorm:"index;not null" json:"user_id"`
	AddressLine1 string `gorm:"not null" json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	City         string `gorm:"not null" json:"city"`
	State        string `gorm:"not null" json:"state"`
	PostalCode   string `gorm:"not null" json:"postal_code"`
	Country      string `gorm:"not null" json:"country"`
	IsCurrent    bool   `gorm:"index;default:false" json:"is_current"`
}

type Product struct {
	Base
	SKU           string  `gorm:"uniqueIndex;not null" json:"sku"`
	Name          string  `gorm:"index;not null" json:"name"`
	Description   string  `json:"description"`
	CurrentPrice  float64 `gorm:"type:decimal(10,2);not null" json:"current_price"`
	StockQuantity int     `gorm:"not null" json:"stock_quantity"`
}

type Order struct {
	Base
	UserID      uint            `gorm:"index;not null" json:"user_id"`
	AddressID   uint            `gorm:"index;not null" json:"address_id"`
	Address     UserAddress     `gorm:"foreignKey:AddressID" json:"address"`
	Status      OrderStatus     `gorm:"index;type:text;default:'PENDING'" json:"status"`
	TotalAmount float64         `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	OrderItems  []OrderItem     `gorm:"foreignKey:OrderID" json:"items"`
	EventLogs   []OrderEventLog `gorm:"foreignKey:OrderID" json:"events"`
}

type OrderItem struct {
	Base
	OrderID           uint    `gorm:"index;not null" json:"order_id"`
	ProductID         uint    `gorm:"index;not null" json:"product_id"`
	ProductName       string  `gorm:"not null" json:"product_name"`
	Quantity          int     `gorm:"not null" json:"quantity"`
	UnitPriceSnapshot float64 `gorm:"type:decimal(10,2);not null" json:"unit_price_snapshot"`
	Subtotal          float64 `gorm:"type:decimal(10,2);not null" json:"subtotal"`
}

type OrderEventLog struct {
	Base
	OrderID     uint        `gorm:"index;not null" json:"order_id"`
	FromStatus  OrderStatus `json:"from_status"`
	ToStatus    OrderStatus `json:"to_status"`
	Reason      string      `json:"reason"`
	TriggeredBy string      `json:"triggered_by"` // User UUID or "SYSTEM"
}
