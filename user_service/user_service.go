package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"strings"
	"time"
)

type UserService struct {
	Validator *validator.Validate
	DB        *sql.DB
	Ctx       context.Context
}

func NewUserService(validator *validator.Validate, db *sql.DB) *UserService {
	return &UserService{Validator: validator, DB: db, Ctx: context.Background()}
}

func (s *UserService) RegisterRoutes(router fiber.Router) {
	internalAPI := router.Group("/users")
	internalAPI.Use(shared.JWTServiceMiddleware)
	internalAPI.Post("/", s.handleCreateUser)
	internalAPI.Get("/", s.handleGetUser)
	internalAPI.Put("/", s.handleUpdateUser)
	internalAPI.Delete("/:id", s.handleDeleteUser)
	internalAPI.Patch("/verify/:token", s.handleVerifyEmail)

	router.Get("/me", s.handleGetUser)
}

func (s *UserService) handleCreateUser(c *fiber.Ctx) error {
	registerRequest := &RegisterRequest{}
	err := c.BodyParser(registerRequest)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	err = s.Validator.Struct(registerRequest)
	if err != nil && errors.As(err, &validator.ValidationErrors{}) {
		return shared.NewFailedValidationError(*registerRequest, err.(validator.ValidationErrors))
	}

	// Hash the password
	password, err := bcrypt.GenerateFromPassword([]byte(registerRequest.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	registerRequest.Password = string(password)

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer shared.CommitOrRollback(tx, err)

	result, err := tx.ExecContext(s.Ctx, "INSERT INTO users (id, name, email, password,phone_number, email_verification_token) VALUES (NULL, ?, ?, ?, ?, ?)",
		registerRequest.Name, registerRequest.Email, registerRequest.Password, registerRequest.PhoneNumber, registerRequest.EmailVerificationToken)
	if err != nil {
		if mysqlErr, ok := err.(*mysql.MySQLError); ok {
			slog.Error("MySQL error occurred", "code", mysqlErr.Number, "message", mysqlErr.Message)
			// Check for duplicate entry error
			if mysqlErr.Number == 1062 {
				errMsg := ""
				// Parse the error message to find which constraint was violated
				if strings.Contains(mysqlErr.Message, "email") {
					errMsg = "email already exists"
				} else if strings.Contains(mysqlErr.Message, "phone_number") {
					errMsg = "phone number already exists"
				}
				return fiber.NewError(fiber.StatusConflict, "Duplicate entry. "+errMsg)
			}
		}
		slog.Info("Error occurred while inserting user", "err", err)
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		slog.Info("No rows affected while inserting user", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create user")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"data":    nil,
		"errors":  nil,
	})
}

func (s *UserService) handleGetUser(c *fiber.Ctx) error {
	userID := c.Query("id")
	userEmail := c.Query("email")

	if userEmail == "" && userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "User ID or email is required")
	} else if userEmail != "" && userID != "" {
		return fiber.NewError(fiber.StatusBadRequest, "Only one of user ID or email should be provided")
	}

	target := userID
	column := "id"
	if userEmail != "" {
		column = "email"
		target = userEmail
	}
	query := fmt.Sprintf("SELECT id, name, email, password, phone_number, email_verification_token, email_verified_at, created_at, updated_at FROM users WHERE %s = ?", column)

	rows, err := s.DB.QueryContext(s.Ctx, query, target)
	if err != nil {
		slog.Debug("Error occurred while querying user", "err", err)
		return err
	}
	defer rows.Close()

	var user User
	if !rows.Next() {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	var emailVerificationToken sql.NullString
	var emailVerifiedAt sql.NullTime
	err = rows.Scan(&user.Id, &user.Name, &user.Email, &user.Password, &user.PhoneNumber, &emailVerificationToken, &emailVerifiedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		slog.Debug("Error occurred while scanning user", "err", err)
		return err
	}

	if emailVerificationToken.Valid {
		user.EmailVerificationToken = emailVerificationToken.String
	}
	if emailVerifiedAt.Valid {
		user.EmailVerifiedAt = emailVerifiedAt.Time
	}

	return c.JSON(fiber.Map{
		"message": "User retrieved successfully",
		"data":    user,
		"errors":  nil,
	})

}

func (s *UserService) handleUpdateUser(c *fiber.Ctx) error {
	// Initialize model
	user := &User{}

	// Changed the handler's flexibility, adding more than one identifier.
	userId := c.Query("id")
	userEmail := c.Query("email")

	if userEmail == "" && userId == "" {
		return fiber.NewError(fiber.StatusBadRequest, "User ID or email is required")
	} else if userEmail != "" && userId != "" {
		return fiber.NewError(fiber.StatusBadRequest, "Only one of user ID or email should be provided")
	}

	target := userId
	column := "id"
	if userEmail != "" {
		column = "email"
		target = userEmail
	}

	// parse request, converting it from ResetPasswordRequest to User model
	err := c.BodyParser(user)
	if err != nil {
		slog.Error("Error occurred while parsing request body", "err", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	// Encrypt pass akwakwka
	password, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.Password = string(password)

	// Transaction stuffs
	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer shared.CommitOrRollback(tx, err)

	// Build n exec query
	query := spew.Sprintf("UPDATE users SET email = ?, password = ?, updated_at = CURRENT_TIMESTAMP WHERE %s = ?", column)

	result, err := tx.ExecContext(s.Ctx, query, user.Email, user.Password, target)

	if err != nil {
		slog.Info("Internal server error", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update user")
	}

	// Checks update status
	rowsAffected, err := result.RowsAffected()
	if err != nil || rowsAffected == 0 {
		slog.Info("No rows affected while updating user", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to update user")
	}

	return c.JSON(fiber.Map{
		"message": "User updated successfully",
		"data":    nil,
		"errors":  nil,
	})
}

func (s *UserService) handleDeleteUser(c *fiber.Ctx) error {
	// Handle deleting user
	return c.SendString("User deleted successfully")
}

func (s *UserService) handleVerifyEmail(c *fiber.Ctx) error {
	token := c.Params("token")
	if token == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid token")
	}

	query := "SELECT id, name, email, password, phone_number, email_verification_token, email_verified_at, created_at, updated_at FROM users WHERE email_verification_token = ?"

	rows, err := s.DB.QueryContext(s.Ctx, query, token)
	if err != nil {
		slog.Debug("Error occurred while querying user", "err", err)
		return err
	}
	defer rows.Close()

	var user User
	if !rows.Next() {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	var emailVerifiedAt sql.NullTime
	err = rows.Scan(&user.Id, &user.Name, &user.Email, &user.Password, &user.PhoneNumber, &user.EmailVerificationToken, &emailVerifiedAt, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		slog.Debug("Error occurred while scanning user", "err", err)
		return err
	}

	if emailVerifiedAt.Valid {
		user.EmailVerifiedAt = emailVerifiedAt.Time
	}

	if user.EmailVerificationToken == token {
		// Update the email verification token and set email verified at
		user.EmailVerificationToken = ""
		user.EmailVerifiedAt = time.Now()

		_, err = s.DB.ExecContext(s.Ctx, "UPDATE users SET email_verification_token = ?, email_verified_at = ? WHERE id = ?", user.EmailVerificationToken, user.EmailVerifiedAt, user.Id)
		if err != nil {
			slog.Debug("Error occurred while updating user", "err", err)
			return err
		}
	} else {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid token")
	}

	return c.JSON(fiber.Map{
		"message": "Email verified",
		"data":    nil,
		"errors":  nil,
	})
}
