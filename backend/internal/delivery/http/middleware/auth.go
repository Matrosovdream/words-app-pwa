package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

const (
	CtxUserID    = "user_id"
	CtxUserEmail = "user_email"
	CtxUserRole  = "user_role"
)

// NewJWTMiddleware returns a Fiber middleware that validates the Bearer token
// and stashes the claims on the context.
func NewJWTMiddleware(secret string) fiber.Handler {
	key := []byte(secret)
	return func(ctx *fiber.Ctx) error {
		header := ctx.Get("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			return fiber.ErrUnauthorized
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")

		token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.ErrUnauthorized
			}
			return key, nil
		})
		if err != nil || !token.Valid {
			return fiber.ErrUnauthorized
		}
		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			return fiber.ErrUnauthorized
		}
		if sub, ok := claims["sub"].(string); ok {
			ctx.Locals(CtxUserID, sub)
		}
		if email, ok := claims["email"].(string); ok {
			ctx.Locals(CtxUserEmail, email)
		}
		if role, ok := claims["role"].(string); ok {
			ctx.Locals(CtxUserRole, role)
		}
		return ctx.Next()
	}
}

// GetUserID returns the user id stashed on the context by the JWT middleware.
func GetUserID(ctx *fiber.Ctx) string {
	if v, ok := ctx.Locals(CtxUserID).(string); ok {
		return v
	}
	return ""
}
