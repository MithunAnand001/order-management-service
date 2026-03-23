package middleware

import (
	"context"

	"github.com/google/uuid"
)

type contextKey string

const (
	ClaimsKey contextKey = "user_claims"
	TokenKey  contextKey = "jwt_token"
)

type UserClaims struct {
	UserID uint
	UUID   uuid.UUID
	Email  string
}

func GetClaims(ctx context.Context) *UserClaims {
	if claims, ok := ctx.Value(ClaimsKey).(*UserClaims); ok {
		return claims
	}
	return nil
}

func GetToken(ctx context.Context) string {
	if token, ok := ctx.Value(TokenKey).(string); ok {
		return token
	}
	return ""
}
