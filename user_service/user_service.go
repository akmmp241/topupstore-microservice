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
	internalAPI := router.Group("/user-service")
	internalAPI.Use(shared.JWTServiceMiddleware)
	internalAPI.Post("/create", s.handleCreateUser)
	internalAPI.Get("/get/:id", s.handleGetUser)
	internalAPI.Put("/update/:id", s.handleUpdateUser)
	internalAPI.Delete("/delete/:id", s.handleDeleteUser)

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

	// Handle user creation
	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "User created successfully",
		"data":    nil,
		"errors":  nil,
	})
}

func (s *UserService) handleGetUser(c *fiber.Ctx) error {
	slog.Info("Getting user")
	// Handle getting user
	return c.SendString("User retrieved successfully")
}

func (s *UserService) handleUpdateUser(c *fiber.Ctx) error {
	return c.SendString("User updated successfully")
}

func (s *UserService) handleDeleteUser(c *fiber.Ctx) error {
	// Handle deleting user
	return c.SendString("User deleted successfully")
}
