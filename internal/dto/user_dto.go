package dto

type CreateUserRequest struct {
	Name     string                `json:"name" validate:"required,min=2,max=50"`
	Email    string                `json:"email" validate:"required,email"`
	Password string                `json:"password" validate:"required,min=8"`
	RoleCode string                `json:"role_code" validate:"required"` // ADMIN, USER, DELIVERY
	Address  *CreateAddressRequest `json:"address,omitempty"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	UUID      string            `json:"uuid"`
	Name      string            `json:"name"`
	Email     string            `json:"email"`
	Role      string            `json:"role"`
	CreatedOn string            `json:"created_on"`
	IsActive  bool              `json:"is_active"`
	Addresses []AddressResponse `json:"addresses,omitempty"`
}

type CreateAddressRequest struct {
	AddressLine1 string `json:"address_line1" validate:"required"`
	AddressLine2 string `json:"address_line2"`
	City         string `json:"city" validate:"required"`
	State        string `json:"state" validate:"required"`
	PostalCode   string `json:"postal_code" validate:"required"`
	Country      string `json:"country" validate:"required"`
	IsCurrent    bool   `json:"is_current"`
}

type AddressResponse struct {
	UUID         string `json:"uuid"`
	AddressLine1 string `json:"address_line1"`
	AddressLine2 string `json:"address_line2"`
	City         string `json:"city"`
	State        string `json:"state"`
	PostalCode   string `json:"postal_code"`
	Country      string `json:"country"`
	IsCurrent    bool   `json:"is_current"`
}

type LoginResponse struct {
	AccessToken  string       `json:"access_token"`
	RefreshToken string       `json:"refresh_token"`
	User         UserResponse `json:"user"`
}

type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
}
