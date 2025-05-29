package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type AuthService struct {
	Producer    *KafkaProducer
	Validator   *validator.Validate
	RedisClient *redis.Client
}

func NewAuthService(p *KafkaProducer, v *validator.Validate, r *redis.Client) *AuthService {
	return &AuthService{
		Producer:    p,
		Validator:   v,
		RedisClient: r,
	}
}

func (s *AuthService) RegisterRoutes(router fiber.Router) {
	router.Post("/register", s.handleRegister)
	router.Post("/login", s.Login)
	router.Get("/verify/:token", s.handleVerifyEmail)
	router.Post("/password", s.handleForgotPassword)
	router.Patch("/password/:reset_token", s.handleResetPassword)
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

	registerRequest.EmailVerificationToken = uuid.NewString()

	resp, err := CallUserService("/users", fiber.MethodPost, registerRequest)
	if err != nil || len(resp.Errs) > 0 {
		slog.Error("Error occurred while calling user service", "errs", resp.Errs)
		slog.Error("Error occurred while calling user service", "err", err)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"message": "Internal Server Error",
			"errors":  nil,
		})
	}

	if resp.StatusCode != fiber.StatusCreated {
		slog.Error("User service returned non-200 status code", "code", resp.StatusCode)
		return fiber.NewError(resp.StatusCode, string(resp.Body))
	}

	newRegistrationMsg := &NewRegistrationMessage{
		Email:           registerRequest.Email,
		Name:            registerRequest.Name,
		VerificationUrl: fmt.Sprintf("%s/api/auth/verify/%s", os.Getenv("APP_URL"), registerRequest.EmailVerificationToken),
	}

	newRegistrationMsgBytes, err := json.Marshal(newRegistrationMsg)

	msg := [2]string{"user-registration", string(newRegistrationMsgBytes)}

	if err := s.Producer.Write(c.Context(), "user-registration", msg); err != nil {
		slog.Error("Error occurred while sending message to Kafka", "err", err)
	}

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
		slog.Error("Error occurred while calling user service", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	if resp.StatusCode != fiber.StatusOK {
		slog.Error("User service returned non-200 status code", "code", resp.StatusCode)
		return fiber.NewError(resp.StatusCode, string(resp.Body))
	}

	getUserResponse := &GetUserResponse{}
	err = json.Unmarshal(resp.Body, getUserResponse)
	if err != nil {
		slog.Error("Error occurred while unmarshalling response body", "err", err)
		return err
	}
	userResponse := getUserResponse.Data

	err = bcrypt.CompareHashAndPassword([]byte(userResponse.Password), []byte(loginRequest.Password))
	if err != nil {
		slog.Error("Error occurred while comparing passwords", "err", err)
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	// Generate JWT token
	expiry := time.Now().Add(time.Hour * 24)
	userId := strconv.Itoa(userResponse.Id)
	accessToken, err := shared.GenerateJWTForUser(userId, expiry)
	if err != nil {
		slog.Error("Error occurred while generating JWT token", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	newLoginMsg := &NewLoginMessage{
		Email:     userResponse.Email,
		Name:      userResponse.Name,
		LoginTime: time.Now(),
		IpAddress: c.IP(),
		Device:    c.Get("User-Agent"),
	}
	newLoginMsgBytes, err := json.Marshal(newLoginMsg)

	msg := [2]string{"new-login", string(newLoginMsgBytes)}

	err = s.Producer.Write(c.Context(), "user-login", msg)
	if err != nil {
		slog.Error("Error occurred while sending message to Kafka", "err", err)
	}

	return c.JSON(fiber.Map{
		"message": "Login successful",
		"data":    fiber.Map{"access_token": accessToken},
		"errors":  nil,
	})
}

func (s *AuthService) handleVerifyEmail(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid token")
	}

	url := fmt.Sprintf("/users/verify/%s", token)

	resp, err := CallUserService(url, fiber.MethodPatch, nil)
	if err != nil || len(resp.Errs) > 0 {
		slog.Error("Error occurred while calling user service", "errs", resp.Errs)
		slog.Error("Error occurred while calling user service", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	if resp.StatusCode != fiber.StatusOK {
		slog.Error("User service returned non-200 status code", "code", resp.StatusCode)
		return fiber.NewError(resp.StatusCode, string(resp.Body))
	}

	return c.JSON(fiber.Map{
		"message": "Email verified successfully",
		"data":    nil,
		"errors":  nil,
	})
}

func (s *AuthService) handleForgotPassword(c *fiber.Ctx) error {
	forgotPasswordRequest := &ForgotPasswordRequest{}
	err := c.BodyParser(forgotPasswordRequest)
	if err != nil {
		slog.Error("Error occurred while parsing request body", "err", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	err = s.Validator.Struct(forgotPasswordRequest)
	if err != nil && errors.As(err, &validator.ValidationErrors{}) {
		return shared.NewFailedValidationError(*forgotPasswordRequest, err.(validator.ValidationErrors))
	}

	url := fmt.Sprintf("/users?email=%s", forgotPasswordRequest.Email)
	resp, err := CallUserService(url, fiber.MethodGet, nil)
	if err != nil || len(resp.Errs) > 0 {
		slog.Error("Error occurred while calling user service", "errs", resp.Errs)
		slog.Error("Error occurred while calling user service", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	if resp.StatusCode == fiber.StatusNotFound {
		return c.JSON(fiber.Map{
			"message": "Password reset instructions sent to email",
			"data":    nil,
			"errors":  nil,
		})
	}

	if resp.StatusCode != fiber.StatusOK {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	getUserResponse := &GetUserResponse{}
	err = json.Unmarshal(resp.Body, getUserResponse)
	if err != nil {
		slog.Error("Error occurred while unmarshalling response body", "err", err)
		return err
	}

	resetToken := uuid.NewString()
	expiration := time.Hour

	key := fmt.Sprintf("forgot-password:%s", resetToken)
	err = s.RedisClient.SetEx(c.Context(), key, getUserResponse.Data.Email, expiration).Err()
	if err != nil {
		slog.Error("Error occurred while setting Redis key", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	forgotPasswordMsg := &ForgotPasswordMessage{
		Email:     getUserResponse.Data.Email,
		Name:      getUserResponse.Data.Name,
		ResetUrl:  fmt.Sprintf("%s/reset-password/%s", os.Getenv("APP_URL"), resetToken),
		ExpiresAt: time.Now().Add(expiration),
	}

	forgotPasswordMsgBytes, err := json.Marshal(forgotPasswordMsg)
	if err != nil {
		slog.Error("Error occurred while marshalling message", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	msg := [2]string{"forgot-password", string(forgotPasswordMsgBytes)}
	if err := s.Producer.Write(c.Context(), "forget-password", msg); err != nil {
		slog.Error("Error occurred while sending message to Kafka", "err", err)
	}

	return c.JSON(fiber.Map{
		"message": "Password reset instructions sent to email",
		"data":    nil,
		"errors":  nil,
	})
}

func (s *AuthService) handleResetPassword(c *fiber.Ctx) error {
	// parse body n take the reset_token
	resetPasswordRequest := &ResetPasswordRequest{}
	err := c.BodyParser(resetPasswordRequest)
	resetPasswordRequest.ResetToken = c.Params("reset_token")

	if err != nil {
		slog.Error("Error occurred while parsing request body", "err", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// to validate, rules can be seen in a dto file
	err = s.Validator.Struct(resetPasswordRequest)

	if err != nil && errors.As(err, &validator.ValidationErrors{}) {
		return shared.NewFailedValidationError(*resetPasswordRequest, err.(validator.ValidationErrors))
	}

	// checks both password and confirm_password's integrity
	if resetPasswordRequest.Password != resetPasswordRequest.PasswordConfirmation {
		slog.Error("New Password and password confirmation must be matched", "err", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// checks reset token integrity by accessing redis
	key := fmt.Sprintf("forgot-password:%s", resetPasswordRequest.ResetToken)
	val, err := s.RedisClient.Get(c.Context(), key).Result()
	if err != nil {
		return err
	}
	if val == "" {
		slog.Error("Reset token is not valid", "err", err)
		return fiber.NewError(fiber.StatusBadRequest, "Reset token is not valid")
	}

	updateRequest := &UpdateUserRequest{
		Password: resetPasswordRequest.Password,
	}

	// calls user service
	url := fmt.Sprintf("/users?email=%s", val)
	resp, err := CallUserService(url, fiber.MethodPut, updateRequest)
	if err != nil || len(resp.Errs) > 0 {
		slog.Error("Error occurred while calling user service", "errs", resp.Errs)
		slog.Error("Error occurred while calling user service", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	if resp.StatusCode != fiber.StatusOK {
		slog.Error("User service returned non-200 status code", "code", resp.StatusCode)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	// Delete the reset token from Redis
	s.RedisClient.Del(c.Context(), key)

	return c.JSON(fiber.Map{
		"message": "Password changed successfully",
		"data":    nil,
		"errors":  nil,
	})
}
