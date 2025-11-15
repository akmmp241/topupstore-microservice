package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"math"
	"strconv"

	ipb "github.com/akmmp241/topupstore-microservice/indexer-proto/v1"
	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type Map map[string]interface{}

type ProductService struct {
	validate       *validator.Validate
	DB             *sql.DB
	EsClient       *elasticsearch.Client
	Ctx            context.Context
	IndexerService *ipb.IndexerServiceClient
}

func NewProductService(validate *validator.Validate, DB *sql.DB, IndexService *ipb.IndexerServiceClient, esClient *elasticsearch.Client) *ProductService {
	return &ProductService{validate: validate, DB: DB, Ctx: context.Background(), IndexerService: IndexService, EsClient: esClient}
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

	route.Get("/products-index", shared.DevOnlyMiddleware, p.handleProductIndexingToES)
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
	idStr := c.Params("id")
	if idStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Category ID is required")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error("Invalid category ID", "error", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid category ID")
	}

	tx, err := p.DB.Begin()
	if err != nil {
		slog.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer shared.CommitOrRollback(tx, err)

	query := "SELECT id, ref_id, name, created_at, updated_at FROM categories WHERE id = ?"
	row := p.DB.QueryRowContext(p.Ctx, query, id)

	var category Category
	if err := row.Scan(&category.Id, &category.RefId, &category.Name, &category.CreatedAt, &category.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Category not found")
		}
		slog.Error("Failed to scan category row", "error", err)
		return err
	}

	return c.JSON(fiber.Map{
		"message": "Category retrieved successfully",
		"data":    category,
		"errors":  nil,
	})
}

func (p *ProductService) handleGetOperatorsByCategoryID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Category ID is required")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error("Invalid category ID", "error", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid category ID")
	}

	tx, err := p.DB.Begin()
	if err != nil {
		slog.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer shared.CommitOrRollback(tx, err)

	query := "SELECT id, ref_id, category_id, name, slug, image_url, description, created_at, updated_at FROM operators WHERE category_id = ?"
	rows, err := p.DB.QueryContext(p.Ctx, query, id)
	if err != nil {
		slog.Error("Failed to query operators", "error", err)
		return err
	}
	defer rows.Close()

	var operators []Operator
	for rows.Next() {
		var operator Operator
		if err := rows.Scan(&operator.Id, &operator.RefId, &operator.CategoryId, &operator.Name, &operator.Slug, &operator.ImageUrl, &operator.Description, &operator.CreatedAt, &operator.UpdatedAt); err != nil {
			slog.Error("Failed to scan operator row", "error", err)
			return err
		}
		operators = append(operators, operator)
	}

	return c.JSON(fiber.Map{
		"message": "Operators retrieved successfully",
		"data": fiber.Map{
			"operators": operators,
		},
		"errors": nil,
	})
}

func (p *ProductService) handleGetOperators(c *fiber.Ctx) error {
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

	query := "SELECT id, ref_id, category_id, name, slug, image_url, description, created_at, updated_at FROM operators WHERE id > ? ORDER BY id LIMIT ?"

	rows, err := p.DB.QueryContext(p.Ctx, query, afterID, limit)
	if err != nil {
		slog.Error("Failed to query operators", "error", err)
		return err
	}
	defer rows.Close()

	var operators []Operator
	var lastID int
	for rows.Next() {
		var operator Operator
		if err := rows.Scan(&operator.Id, &operator.RefId, &operator.CategoryId, &operator.Name, &operator.Slug, &operator.ImageUrl, &operator.Description, &operator.CreatedAt, &operator.UpdatedAt); err != nil {
			slog.Error("Failed to scan operator row", "error", err)
			return err
		}
		operators = append(operators, operator)
		lastID = operator.Id
	}

	// Set the next page cursor
	var nextCursor *int
	if len(operators) == limit {
		nextCursor = &lastID
	}

	return c.JSON(fiber.Map{
		"message": "Operators retrieved successfully",
		"data": fiber.Map{
			"operators":   operators,
			"next_cursor": nextCursor,
		},
		"errors": nil,
	})
}

func (p *ProductService) handleGetOperatorByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Operator ID is required")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error("Invalid operator ID", "error", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid operator ID")
	}

	tx, err := p.DB.Begin()
	if err != nil {
		slog.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer shared.CommitOrRollback(tx, err)

	query := "SELECT id, ref_id, category_id, name, slug, image_url, description, created_at, updated_at FROM operators WHERE id = ?"
	row := p.DB.QueryRowContext(p.Ctx, query, id)

	var operator Operator
	if err := row.Scan(&operator.Id, &operator.RefId, &operator.CategoryId, &operator.Name, &operator.Slug, &operator.ImageUrl, &operator.Description, &operator.CreatedAt, &operator.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Operator not found")
		}
		slog.Error("Failed to scan operator row", "error", err)
		return err
	}

	return c.JSON(fiber.Map{
		"message": "Operator retrieved successfully",
		"data":    operator,
		"errors":  nil,
	})
}

func (p *ProductService) handleGetProductTypesByOperatorID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Operator ID is required")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error("Invalid operator ID", "error", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid operator ID")
	}

	tx, err := p.DB.Begin()
	if err != nil {
		slog.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer shared.CommitOrRollback(tx, err)

	query := "SELECT id, ref_id, operator_id, name, format_form, created_at, updated_at FROM product_types WHERE operator_id = ?"
	rows, err := p.DB.QueryContext(p.Ctx, query, id)
	if err != nil {
		slog.Error("Failed to query product types", "error", err)
		return err
	}
	defer rows.Close()

	var productTypes []ProductType
	for rows.Next() {
		var productType ProductType
		if err := rows.Scan(&productType.Id, &productType.RefId, &productType.OperatorId, &productType.Name, &productType.FormatForm, &productType.CreatedAt, &productType.UpdatedAt); err != nil {
			slog.Error("Failed to scan product type row", "error", err)
			return err
		}
		productTypes = append(productTypes, productType)
	}

	return c.JSON(fiber.Map{
		"message": "Product types retrieved successfully",
		"data": fiber.Map{
			"product_types": productTypes,
		},
		"errors": nil,
	})
}

func (p *ProductService) handleGetProductTypes(c *fiber.Ctx) error {
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

	query := "SELECT id, ref_id, operator_id, name, format_form, created_at, updated_at FROM product_types WHERE id > ? ORDER BY id LIMIT ?"

	rows, err := p.DB.QueryContext(p.Ctx, query, afterID, limit)
	if err != nil {
		slog.Error("Failed to query product types", "error", err)
		return err
	}
	defer rows.Close()

	var productTypes []ProductType
	var lastID int
	for rows.Next() {
		var productType ProductType
		if err := rows.Scan(&productType.Id, &productType.RefId, &productType.OperatorId, &productType.Name, &productType.FormatForm, &productType.CreatedAt, &productType.UpdatedAt); err != nil {
			slog.Error("Failed to scan product type row", "error", err)
			return err
		}
		productTypes = append(productTypes, productType)
		lastID = productType.Id
	}

	// Set the next page cursor
	var nextCursor *int
	if len(productTypes) == limit {
		nextCursor = &lastID
	}

	return c.JSON(fiber.Map{
		"message": "Product types retrieved successfully",
		"data": fiber.Map{
			"product_types": productTypes,
			"next_cursor":   nextCursor,
		},
		"errors": nil,
	})
}

func (p *ProductService) handleGetProductTypeByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Product type ID is required")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error("Invalid product type ID", "error", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid product type ID")
	}

	tx, err := p.DB.Begin()
	if err != nil {
		slog.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer shared.CommitOrRollback(tx, err)

	query := "SELECT id, ref_id, operator_id, name, format_form, created_at, updated_at FROM product_types WHERE id = ?"
	row := p.DB.QueryRowContext(p.Ctx, query, id)

	var productType ProductType
	if err := row.Scan(&productType.Id, &productType.RefId, &productType.OperatorId, &productType.Name, &productType.FormatForm, &productType.CreatedAt, &productType.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Product type not found")
		}
		slog.Error("Failed to scan product type row", "error", err)
		return err
	}

	return c.JSON(fiber.Map{
		"message": "Product type retrieved successfully",
		"data":    productType,
		"errors":  nil,
	})
}

func (p *ProductService) handleGetProductsByProductTypeID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Product type ID is required")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error("Invalid product type ID", "error", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid product type ID")
	}

	tx, err := p.DB.Begin()
	if err != nil {
		slog.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer shared.CommitOrRollback(tx, err)

	query := "SELECT id, ref_id, product_type_id, name, description, image_url, created_at, updated_at FROM products WHERE product_type_id = ?"
	rows, err := p.DB.QueryContext(p.Ctx, query, id)
	if err != nil {
		slog.Error("Failed to query products", "error", err)
		return err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var product Product
		if err := rows.Scan(&product.Id, &product.RefId, &product.ProductTypeId, &product.Name, &product.Description, &product.ImageUrl, &product.CreatedAt, &product.UpdatedAt); err != nil {
			slog.Error("Failed to scan product row", "error", err)
			return err
		}
		products = append(products, product)
	}

	return c.JSON(fiber.Map{
		"message": "Products retrieved successfully",
		"data": fiber.Map{
			"products": products,
		},
		"errors": nil,
	})
}

func (p *ProductService) handleGetProducts(c *fiber.Ctx) error {
	query := c.Query("q", "")
	pageStr := c.Query("page", "1")
	sizeStr := c.Query("size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size, err := strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 {
		size = 10
	}

	from := (page - 1) * size

	esRes, err := p.searchProducts(c.Context(), query, from, size)
	if err != nil {
		return err
	}

	var products []*ProductSearch
	for _, hit := range esRes.Hits.Hits {
		products = append(products, &hit.Source)
	}

	totalHits := esRes.Hits.Total.Value
	totalPages := int(math.Ceil(float64(totalHits) / float64(size)))

	return c.JSON(fiber.Map{
		"message": "Products retrieved successfully",
		"data": fiber.Map{
			"products": products,
			"meta": fiber.Map{
				"total_hits":   totalHits,
				"total_pages":  totalPages,
				"current_page": page,
				"size":         size,
			},
		},
		"errors": nil,
	})
}

func (p *ProductService) handleGetProductByID(c *fiber.Ctx) error {
	idStr := c.Params("id")
	if idStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Product ID is required")
	}

	id, err := strconv.Atoi(idStr)
	if err != nil {
		slog.Error("Invalid product ID", "error", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid product ID")
	}

	tx, err := p.DB.Begin()
	if err != nil {
		slog.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer shared.CommitOrRollback(tx, err)

	query := "SELECT id, ref_id, product_type_id, name, description, image_url, price, created_at, updated_at FROM products WHERE id = ?"
	row := p.DB.QueryRowContext(p.Ctx, query, id)

	var product Product
	if err := row.Scan(&product.Id, &product.RefId, &product.ProductTypeId, &product.Name, &product.Description, &product.ImageUrl, &product.Price, &product.CreatedAt, &product.UpdatedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Product not found")
		}
		slog.Error("Failed to scan product row", "error", err)
		return err
	}

	return c.JSON(fiber.Map{
		"message": "Product retrieved successfully",
		"data":    product,
		"errors":  nil,
	})
}

func (p *ProductService) handleProductIndexingToES(c *fiber.Ctx) error {
	tx, err := p.DB.Begin()
	if err != nil {
		slog.Error("Failed to begin transaction", "error", err)
		return err
	}
	defer shared.CommitOrRollback(tx, err)

	stream, err := (*p.IndexerService).BulkIndexProducts(p.Ctx)
	if err != nil {
		slog.Error("Failed to bulk index products", "error", err)
		return err
	}

	query := `
		SELECT
			p.id, p.name, p.image_url, p.price, p.description,
			pt.id as type_id, pt.name as type_name,
			o.id as operator_id, o.name as operator_name, o.slug as operator_slug, o.image_url as operator_image_url,
			c.id as category_id, c.name as category_name
		FROM products p
			JOIN product_types pt ON p.product_type_id = pt.id
			JOIN operators o ON pt.operator_id = o.id
			JOIN categories c ON o.category_id = c.id;
`

	rows, err := p.DB.QueryContext(p.Ctx, query)
	if err != nil {
		slog.Error("Failed to query products", "error", err)
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var product ipb.Product

		// for avoiding nil pointer dereference
		product.ProductType = &ipb.ProductType{}
		product.Operator = &ipb.Operator{}
		product.Category = &ipb.Category{}

		err = rows.Scan(
			&product.Id, &product.Name, &product.ImageUrl, &product.Price, &product.Description,
			&product.ProductType.Id, &product.ProductType.Name,
			&product.Operator.Id, &product.Operator.Name, &product.Operator.Slug, &product.Operator.ImageUrl,
			&product.Category.Id, &product.Category.Name,
		)
		if err != nil {
			slog.Error("Failed to scan product row", "error", err)
			return err
		}

		if err := stream.Send(&product); err != nil {
			slog.Error("Failed to send product to stream", "error", err)
			return err
		}
	}

	summary, err := stream.CloseAndRecv()
	if err != nil {
		slog.Error("Failed to close stream", "error", err)
		return err
	}

	slog.Info("Indexing products to ES completed", "total-indexed", summary.TotalIndexed, "total-failed", summary.TotalFailed)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Indexing products to ES completed",
		"data": fiber.Map{
			"total_indexed": summary.TotalIndexed,
			"total_failed":  summary.TotalFailed,
		},
		"errors": nil,
	})
}

func (p *ProductService) searchProducts(ctx context.Context, query string, from, size int) (*EsResponse, error) {
	var response *EsResponse
	var reqReader io.Reader = nil
	var searchQuery = Map{}

	if query != "" {
		searchQuery = Map{
			"query": Map{
				"multi_match": Map{
					"query":     query,
					"fields":    []string{"name", "description"},
					"fuzziness": "AUTO",
				},
			},
		}
	}

	searchQuery["from"] = from
	searchQuery["size"] = size

	reqBuf, err := json.Marshal(searchQuery)
	if err != nil {
		slog.Error("Error occurred while marshaling request", "error", err)
		return nil, err
	}

	reqReader = bytes.NewReader(reqBuf)

	searchReq := esapi.SearchRequest{
		Index:          []string{"products"},
		Body:           reqReader,
		TrackTotalHits: true,
	}

	res, err := searchReq.Do(ctx, p.EsClient)
	if err != nil {
		slog.Error("Error occurred while searching products", "error", err)
		return nil, err
	}

	if res.IsError() {
		slog.Error("Error occurred while searching products", "error", res.String())
		return nil, errors.New(res.String())
	}

	if err := json.NewDecoder(res.Body).Decode(&response); err != nil {
		slog.Error("Error occurred while decoding response", "error", err)
		return nil, err
	}

	return response, nil
}
