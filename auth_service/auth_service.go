package main

import (
	"errors"
	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"log/slog"
)

type AuthService struct {
	Producer  *KafkaProducer
	Validator *validator.Validate
}

func NewAuthService(p *KafkaProducer, v *validator.Validate) *AuthService {
	return &AuthService{
		Producer:  p,
		Validator: v,
	}
}

func (s *AuthService) RegisterRoutes(router fiber.Router) {
	router.Post("/register", s.handleRegister)
	router.Post("/login", s.Login)
}

func (s *AuthService) handleRegister(c *fiber.Ctx) error {

	registerRequest := &RegisterRequest{}
	err := c.BodyParser(registerRequest)
	if err != nil {
		slog.Error("Error occurred while parsing request body", "err", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	err = s.Validator.Struct(registerRequest)
	if err != nil && errors.As(err, &validator.ValidationErrors{}) {
		return shared.NewFailedValidationError(*registerRequest, err.(validator.ValidationErrors))
	}

	resp, err := CallUserService("/user-service/create", fiber.MethodPost, registerRequest)
	if err != nil {
		slog.Error("Error occurred while calling user service", "err", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"errors":  nil,
		})
	}

	if len(resp.Errs) > 0 {
		slog.Error("Error occurred while calling user service", "errs", resp.Errs)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"errors":  nil,
		})
	}

	if resp.StatusCode != fiber.StatusCreated {
		slog.Error("User service returned non-200 status code", "code", resp.StatusCode)
		return c.Status(resp.StatusCode).JSON(fiber.Map{
			"message": "User service error",
			"errors":  nil,
		})
	}

	// Handle user registration
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"data":    nil,
		"errors":  nil,
	})
}

func (s *AuthService) Login(c *fiber.Ctx) error {
	slog.Info("Logging in user")
	return c.SendString("Login successfully")
}
