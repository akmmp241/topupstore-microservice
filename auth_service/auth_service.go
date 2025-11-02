package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"

	"github.com/akmmp241/topupstore-microservice/shared"
	upb "github.com/akmmp241/topupstore-microservice/user-proto/v1"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthService struct {
	Producer          *KafkaProducer
	Validator         *validator.Validate
	RedisClient       *redis.Client
	UserServiceClient *upb.UserServiceClient
}

func NewAuthService(p *KafkaProducer, v *validator.Validate, r *redis.Client, u *upb.UserServiceClient) *AuthService {
	return &AuthService{
		Producer:          p,
		Validator:         v,
		RedisClient:       r,
		UserServiceClient: u,
	}
}

func (s *AuthService) RegisterRoutes(router fiber.Router) {
	router.Post("/register", s.handleRegister)
	router.Post("/login", s.Login)
	router.Get("/verify/:token", s.handleVerifyEmail)
	router.Post("/password", s.handleForgotPassword)
	router.Patch("/password/:reset_token", s.handleResetPassword)
	router.Get("/me", s.handleGetUser).Use(shared.JWTUserMiddleware)
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

	createUserReq := upb.CreateUserReq{
		Name:                   registerRequest.Name,
		Email:                  registerRequest.Email,
		Password:               registerRequest.Password,
		PhoneNumber:            registerRequest.PhoneNumber,
		EmailVerificationToken: registerRequest.EmailVerificationToken,
	}

	_, err = (*s.UserServiceClient).CreateUser(c.Context(), &createUserReq)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.AlreadyExists {
			slog.Error("User already exists", "err", err)
			return fiber.NewError(fiber.StatusConflict, st.Message())
		}

		slog.Error("Error occurred while calling user service create user", "err", err)
		return err
	}

	newRegistrationMsg := &NewRegistrationMessage{
		Email: registerRequest.Email,
		Name:  registerRequest.Name,
		VerificationUrl: fmt.Sprintf(
			"%s/api/auth/verify/%s",
			os.Getenv("APP_URL"),
			registerRequest.EmailVerificationToken,
		),
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

	getUserByEmailReq := upb.GetUserByEmailReq{
		Email: loginRequest.Email,
	}

	getUserRes, err := (*s.UserServiceClient).GetUserByEmail(c.Context(), &getUserByEmailReq)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return fiber.NewError(fiber.StatusNotFound, "User not found")
		}

		slog.Error("Error occurred while calling user service get user", "err", err)
		return err
	}
	user := getUserRes.GetUser()

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.GetPassword()),
		[]byte(loginRequest.Password),
	)
	if err != nil {
		slog.Error("Error occurred while comparing passwords", "err", err)
		return fiber.NewError(fiber.StatusUnauthorized, "Invalid credentials")
	}

	// Generate JWT token
	expiry := time.Now().Add(time.Hour * 24)
	userId := strconv.Itoa(int(user.GetId()))
	accessToken, err := shared.GenerateJWTForUser(userId, expiry)
	if err != nil {
		slog.Error("Error occurred while generating JWT token", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	newLoginMsg := &NewLoginMessage{
		Email:     user.GetEmail(),
		Name:      user.GetName(),
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

	verifyEmailReq := &upb.VerifyEmailReq{
		EmailVerificationToken: token,
	}

	_, err := (*s.UserServiceClient).VerifyEmail(c.Context(), verifyEmailReq)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			slog.Error("User not found", "err", err)
			return fiber.NewError(fiber.StatusNotFound, st.Message())
		}

		slog.Error("Error occurred while calling user service verify email", "err", err)
		return err
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
		return shared.NewFailedValidationError(
			*forgotPasswordRequest,
			err.(validator.ValidationErrors),
		)
	}

	getUserByEmailReq := upb.GetUserByEmailReq{
		Email: forgotPasswordRequest.Email,
	}

	getUserRes, err := (*s.UserServiceClient).GetUserByEmail(c.Context(), &getUserByEmailReq)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			return c.JSON(fiber.Map{
				"message": "Password reset instructions sent to email",
				"data":    nil,
				"errors":  nil,
			})
		}

		slog.Error("Error occurred while calling user service get user", "err", err)
		return err
	}
	user := getUserRes.GetUser()

	resetToken := uuid.NewString()
	expiration := time.Hour

	key := fmt.Sprintf("forgot-password:%s", resetToken)
	err = s.RedisClient.SetEx(c.Context(), key, user.GetEmail(), expiration).Err()
	if err != nil {
		slog.Error("Error occurred while setting Redis key", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	forgotPasswordMsg := &ForgotPasswordMessage{
		Email:     user.GetEmail(),
		Name:      user.GetName(),
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
		return shared.NewFailedValidationError(
			*resetPasswordRequest,
			err.(validator.ValidationErrors),
		)
	}

	// checks both password and confirm_password's integrity
	if resetPasswordRequest.Password != resetPasswordRequest.PasswordConfirmation {
		slog.Error("New Password and password confirmation must be matched", "err", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// checks reset token integrity by accessing redis
	key := fmt.Sprintf("forgot-password:%s", resetPasswordRequest.ResetToken)
	userEmail, err := s.RedisClient.Get(c.Context(), key).Result()

	if userEmail == "" || err != nil {
		slog.Error("Reset token is not valid", "err", err)
		return fiber.NewError(fiber.StatusBadRequest, "Reset token is not valid")
	}

	// calls user service
	resetPasswordByEmailReq := upb.ResetPasswordByEmailReq{
		Email:    userEmail,
		Password: resetPasswordRequest.Password,
	}

	_, err = (*s.UserServiceClient).ResetPasswordByEmail(c.Context(), &resetPasswordByEmailReq)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			slog.Error("User not found", "err", err)
			return fiber.NewError(fiber.StatusNotFound, st.Message())
		}

		slog.Error("Error occurred while calling user service reset password", "err", err)
		return err
	}

	// Delete the reset token from Redis
	s.RedisClient.Del(c.Context(), key)

	return c.JSON(fiber.Map{
		"message": "Password changed successfully",
		"data":    nil,
		"errors":  nil,
	})
}

func (s *AuthService) handleGetUser(c *fiber.Ctx) error {
	userId, err := shared.GetUserIdFromToken(c)
	if err != nil {
		slog.Error("Error occurred while getting user ID from token", "err", err)
		return err
	}

	getUserRes, err := (*s.UserServiceClient).GetUserById(c.Context(), &upb.GetUserByIdReq{
		Id: userId,
	})
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			slog.Error("User not found", "err", err)
			return fiber.NewError(fiber.StatusNotFound, st.Message())
		}

		slog.Error("Error occurred while calling user service get user", "err", err)
		return err
	}

	getResp := &GetResponse{
		Id:              int(getUserRes.GetUser().Id),
		Name:            getUserRes.GetUser().Name,
		Email:           getUserRes.GetUser().Email,
		PhoneNumber:     getUserRes.GetUser().PhoneNumber,
		EmailVerifiedAt: getUserRes.GetUser().EmailVerifiedAt.AsTime(),
		CreatedAt:       getUserRes.GetUser().CreatedAt.AsTime(),
		UpdatedAt:       getUserRes.GetUser().UpdatedAt.AsTime(),
	}

	return c.JSON(fiber.Map{
		"message": "User retrieved successfully",
		"data":    getResp,
		"errors":  nil,
	})
}
