package main

import (
	"github.com/gofiber/fiber/v2"
	"log/slog"
)

type AuthService struct {
	Producer *KafkaProducer
}

func NewAuthService(p *KafkaProducer) *AuthService {
	return &AuthService{
		Producer: p,
	}
}

func (s *AuthService) RegisterRoutes(router fiber.Router) {
	router.Post("/register", s.handleRegister)
	router.Post("/login", s.Login)
}

func (s *AuthService) handleRegister(c *fiber.Ctx) error {
	slog.Info("Registering user")

	err := s.Producer.Write(
		[2]string{"key-1", "value-1"},
	)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	// Handle user registration
	return c.SendString("Registered successfully")
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	slog.Info("Logging in user")
	return c.SendString("Login successfully")
}
