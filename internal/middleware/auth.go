package middleware

import (
	"context"
	"net/http"
	"strings"

	"order-management-service/internal/dto"
	"order-management-service/internal/repository"
	"order-management-service/internal/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

func AuthMiddleware(secret string, userRepo repository.UserRepository) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				utils.SendJSON(w, http.StatusUnauthorized, utils.NewErrorResponse(r.Context(), dto.NewAppError(dto.ErrCodeUnauthorized, "Authorization header required", http.StatusUnauthorized, nil)))
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				utils.SendJSON(w, http.StatusUnauthorized, utils.NewErrorResponse(r.Context(), dto.NewAppError(dto.ErrCodeUnauthorized, "Invalid authorization format", http.StatusUnauthorized, nil)))
				return
			}

			tokenString := parts[1]
			claims := jwt.MapClaims{}

			token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
				return []byte(secret), nil
			})

			if err != nil || !token.Valid {
				utils.SendJSON(w, http.StatusUnauthorized, utils.NewErrorResponse(r.Context(), dto.NewAppError(dto.ErrCodeUnauthorized, "Invalid or expired token", http.StatusUnauthorized, err)))
				return
			}

			userUUIDStr := claims["uuid"].(string)
			userUUID, err := uuid.Parse(userUUIDStr)
			if err != nil {
				utils.SendJSON(w, http.StatusUnauthorized, utils.NewErrorResponse(r.Context(), dto.NewAppError(dto.ErrCodeUnauthorized, "Invalid token identity", http.StatusUnauthorized, err)))
				return
			}

			// DB Lookup for verification and getting internal ID
			user, appErr := userRepo.FindByUUID(r.Context(), userUUID)
			if appErr != nil {
				utils.SendJSON(w, http.StatusUnauthorized, utils.NewErrorResponse(r.Context(), dto.NewAppError(dto.ErrCodeUnauthorized, "User verification failed", http.StatusUnauthorized, appErr.Err)))
				return
			}

			userClaims := &UserClaims{
				UserID: user.ID,
				UUID:   user.UUID,
				Email:  user.Email,
				Role:   user.Role,
			}

			ctx := context.WithValue(r.Context(), ClaimsKey, userClaims)
			ctx = context.WithValue(ctx, TokenKey, tokenString)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
