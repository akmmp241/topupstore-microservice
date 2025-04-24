package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"strconv"
	"time"
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

	resp, err := CallUserService("/users", fiber.MethodPost, registerRequest)
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
			"message": string(resp.Body),
			"errors":  nil,
		})
	}

	// TODO: seng message to kafka for email verification

	// Handle user registration
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User registered successfully",
		"data":    nil,
		"errors":  nil,
	})
}

func (s *AuthService) Login(c *fiber.Ctx) error {

	loginRequest := &LoginRequest{}
	err := c.BodyParser(loginRequest)
	if err != nil {
		slog.Error("Error occurred while parsing request body", "err", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	err = s.Validator.Struct(loginRequest)
	if err != nil && errors.As(err, &validator.ValidationErrors{}) {
		return shared.NewFailedValidationError(*loginRequest, err.(validator.ValidationErrors))
	}

	url := fmt.Sprintf("/users?email=%s", loginRequest.Email)

	resp, err := CallUserService(url, fiber.MethodGet, nil)
	if err != nil || len(resp.Errs) > 0 {
		slog.Error("Error occurred while calling user service", "errs", resp.Errs)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"errors":  nil,
		})
	}

	if resp.StatusCode != fiber.StatusOK {
		slog.Error("User service returned non-200 status code", "code", resp.StatusCode)
		return c.Status(resp.StatusCode).JSON(fiber.Map{
			"message": string(resp.Body),
			"errors":  nil,
		})
	}

	globalResponse := &GlobalServiceResponse{}
	err = json.Unmarshal(resp.Body, globalResponse)
	if err != nil {
		slog.Error("Error occurred while unmarshalling response body", "err", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"errors":  nil,
		})
	}

	userResponse := globalResponse.Data.(GetUserResponse)

	err = bcrypt.CompareHashAndPassword([]byte(userResponse.Password), []byte(loginRequest.Password))
	if err != nil {
		slog.Error("Error occurred while comparing passwords", "err", err)
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"message": "Invalid credentials",
		})
	}

	// Generate JWT token
	expiry := time.Now().Add(time.Hour * 24)
	userId := strconv.Itoa(userResponse.Id)
	accessToken, err := shared.GenerateJWTForUser(userId, expiry)
	if err != nil {
		slog.Error("Error occurred while generating JWT token", "err", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"errors":  nil,
		})
	}

	// TODO: Send message to Kafka for email login

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"data":    fiber.Map{"access_token": accessToken},
		"errors":  nil,
	})
}
