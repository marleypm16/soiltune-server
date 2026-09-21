package middleware

import (
	"crypto/sha256"
	"crypto/subtle"
	"strings"

	"github.com/gofiber/fiber/v3"
)

func RequireAPIKey(expectedKey string) fiber.Handler {
	expectedHash := sha256.Sum256([]byte(expectedKey))

	return func(c fiber.Ctx) error {
		scheme, providedKey, ok := strings.Cut(c.Get(fiber.HeaderAuthorization), " ")
		providedHash := sha256.Sum256([]byte(providedKey))
		if !ok || !strings.EqualFold(scheme, "Bearer") || providedKey == "" ||
			subtle.ConstantTimeCompare(providedHash[:], expectedHash[:]) != 1 {
			c.Set(fiber.HeaderWWWAuthenticate, `Bearer realm="soiltune-api"`)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized",
			})
		}

		return c.Next()
	}
}
