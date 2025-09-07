package shared

import (
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
)

func DevOnlyMiddleware(ctx *fiber.Ctx) error {
	if isDev := os.Getenv("APP_ENV") != "production"; !isDev {
		slog.Error("Trying request to dev endpoint in production")
		return fiber.NewError(fiber.StatusServiceUnavailable, "dev endpoint is unavailable")
	}

	return ctx.Next()
}
