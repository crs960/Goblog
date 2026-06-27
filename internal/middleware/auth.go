package middleware

import (
	"strings"

	"goblog/internal/auth"
	"goblog/internal/httpx"

	"github.com/gofiber/fiber/v2"
)

func JWT(secret string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		header := c.Get("Authorization")
		if header == "" {
			return httpx.Unauthorized("token nao informado")
		}

		tokenString := strings.TrimPrefix(header, "Bearer ")
		if tokenString == header {
			return httpx.Unauthorized("use o formato Bearer token")
		}

		claims, err := auth.ParseToken(tokenString, secret)
		if err != nil {
			return httpx.Unauthorized("token invalido")
		}

		c.Locals("userID", claims.UserID)
		return c.Next()
	}
}

func UserID(c *fiber.Ctx) string {
	userID, _ := c.Locals("userID").(string)
	return userID
}
