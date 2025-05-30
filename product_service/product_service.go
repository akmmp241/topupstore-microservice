package main

import (
	"context"
	"database/sql"
	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"log/slog"
	"strconv"
)

type ProductService struct {
	validate *validator.Validate
	DB       *sql.DB
	Ctx      context.Context
}

func NewProductService(validate *validator.Validate, DB *sql.DB) *ProductService {
	return &ProductService{validate: validate, DB: DB, Ctx: context.Background()}
}

func (p *ProductService) RegisterRoutes(route fiber.Router) {
	route.Get("/categories", p.handleGetCategories)
	route.Get("/categories/:id", p.handleGetCategoryByID)
	route.Get("/categories/:id/operators", p.handleGetOperatorsByCategoryID)
	route.Get("/operators", p.handleGetOperators)
	route.Get("/operators/:id", p.handleGetOperatorByID)
	route.Get("/operators/:id/product-types", p.handleGetProductTypesByOperatorID)
	route.Get("/product-types", p.handleGetProductTypes)
	route.Get("/product-types/:id", p.handleGetProductTypeByID)
	route.Get("/product-types/:id/products", p.handleGetProductsByProductTypeID)
	route.Get("/products", p.handleGetProducts)
	route.Get("/products/:id", p.handleGetProductByID)
}

func (p *ProductService) handleGetCategories(c *fiber.Ctx) error {
	afterStr := c.Query("after")
	limitStr := c.Query("limit")

	var afterID int
	var err error
	if afterStr != "" {
		afterID, err = strconv.Atoi(afterStr)
		if err != nil {
			slog.Error("Invalid 'after' parameter", "error", err)
			return fiber.NewError(fiber.StatusBadRequest, "Invalid 'after' parameter")
		}
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 100 {
		limit = 10
	}

	tx, err := p.DB.Begin()
	if err != nil {
		slog.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer shared.CommitOrRollback(tx, err)

	query := "SELECT id, ref_id, name, created_at, updated_at FROM categories WHERE id > ? ORDER BY id LIMIT ?"

	rows, err := p.DB.QueryContext(p.Ctx, query, afterID, limit)
	if err != nil {
		slog.Error("Failed to query categories", "error", err)
		return err
	}
	defer rows.Close()

	var categories []Category
	var lastID int
	for rows.Next() {
		var category Category
		if err := rows.Scan(&category.Id, &category.RefId, &category.Name, &category.CreatedAt, &category.UpdatedAt); err != nil {
			slog.Error("Failed to scan category row", "error", err)
			return err
		}
		categories = append(categories, category)
		lastID = category.Id
	}

	// Set the next page cursor
	var nextCursor *int
	if len(categories) == limit {
		nextCursor = &lastID
	}

	return c.JSON(fiber.Map{
		"message": "Categories retrieved successfully",
		"data": fiber.Map{
			"categories":  categories,
			"next_cursor": nextCursor,
		},
		"errors": nil,
	})
}

func (p *ProductService) handleGetCategoryByID(c *fiber.Ctx) error {
	// Implementation for getting a category by ID
	return c.SendString("Get category by ID")
}

func (p *ProductService) handleGetOperatorsByCategoryID(c *fiber.Ctx) error {
	// Implementation for getting operators by category ID
	return c.SendString("Get operators by category ID")
}

func (p *ProductService) handleGetOperators(c *fiber.Ctx) error {
	// Implementation for getting all operators
	return c.SendString("Get all operators")
}

func (p *ProductService) handleGetOperatorByID(c *fiber.Ctx) error {
	// Implementation for getting an operator by ID
	return c.SendString("Get operator by ID")
}

func (p *ProductService) handleGetProductTypesByOperatorID(c *fiber.Ctx) error {
	// Implementation for getting product types by operator ID
	return c.SendString("Get product types by operator ID")
}

func (p *ProductService) handleGetProductTypes(c *fiber.Ctx) error {
	// Implementation for getting all product types
	return c.SendString("Get all product types")
}

func (p *ProductService) handleGetProductTypeByID(c *fiber.Ctx) error {
	// Implementation for getting a product type by ID
	return c.SendString("Get product type by ID")
}

func (p *ProductService) handleGetProductsByProductTypeID(c *fiber.Ctx) error {
	// Implementation for getting products by product type ID
	return c.SendString("Get products by product type ID")
}

func (p *ProductService) handleGetProducts(c *fiber.Ctx) error {
	// Implementation for getting all products
	return c.SendString("Get all products")
}

func (p *ProductService) handleGetProductByID(c *fiber.Ctx) error {
	// Implementation for getting a product by ID
	return c.SendString("Get product by ID")
}
