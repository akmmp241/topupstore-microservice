package main

import (
	"log/slog"
	"os"

	"github.com/gofiber/fiber/v2"
)

func WebhookTokenMiddleware(c *fiber.Ctx) error {
	callbackHeader := os.Getenv("XENDIT_CALLBACK_TOKEN_HEADER")
	if callbackHeader == "" {
		slog.Error("Missing configuration: xendit callback token header")
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	callbackToken := os.Getenv("XENDIT_CALLBACK_TOKEN")
	if callbackToken == "" {
		slog.Error("Missing configuration: xendit callback token")
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	if token := c.Get(callbackHeader); token != callbackToken {
		slog.Error("Callback Token Mismatch")
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid Token")
	}

	return c.Next()
}
