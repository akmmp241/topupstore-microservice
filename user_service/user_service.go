package main

import (
	"context"
	"database/sql"
	"errors"
	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/go-playground/validator/v10"
	"github.com/go-sql-driver/mysql"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"strings"
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
	internalAPI.Get("/:id", s.handleGetUser)
	internalAPI.Put("/:id", s.handleUpdateUser)
	internalAPI.Delete("/:id", s.handleDeleteUser)

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

	tx, err := s.DB.Begin()
	if err != nil {
		return err
	}
	defer shared.CommitOrRollback(tx, err)

	result, err := tx.ExecContext(s.Ctx, "INSERT INTO users (id, name, email, password, phone_number) VALUES (NULL, ?, ?, ?, ?)",
		registerRequest.Name, registerRequest.Email, registerRequest.Password, registerRequest.PhoneNumber)
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
	userID := c.Params("id")

	if userID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "User ID is required")
	}

	query := "SELECT id, name, email, phone_number, created_at, updated_at FROM users WHERE id = ?"
	rows, err := s.DB.QueryContext(s.Ctx, query, userID)
	if err != nil {
		slog.Error("Error occurred while querying user", "err", err)
		return err
	}
	defer rows.Close()

	var user User
	if !rows.Next() {
		return fiber.NewError(fiber.StatusNotFound, "User not found")
	}

	err = rows.Scan(&user.Id, &user.Name, &user.Email, &user.PhoneNumber, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		slog.Error("Error occurred while scanning user", "err", err)
		return err
	}

	return c.JSON(fiber.Map{
		"message": "User retrieved successfully",
		"data": fiber.Map{
			"id":           user.Id,
			"name":         user.Name,
			"email":        user.Email,
			"phone_number": user.PhoneNumber,
			"created_at":   user.CreatedAt,
			"updated_at":   user.UpdatedAt,
		},
		"errors": nil,
	})

}

func (s *UserService) handleUpdateUser(c *fiber.Ctx) error {
	return c.SendString("User updated successfully")
}

func (s *UserService) handleDeleteUser(c *fiber.Ctx) error {
	// Handle deleting user
	return c.SendString("User deleted successfully")
}
