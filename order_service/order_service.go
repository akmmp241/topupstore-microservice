package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	urllib "net/url"
	"os"
	"time"

	ppb "github.com/akmmp241/topupstore-microservice/payment-proto/v1"
	prpb "github.com/akmmp241/topupstore-microservice/product-proto/v1"
	"github.com/akmmp241/topupstore-microservice/shared"
	upb "github.com/akmmp241/topupstore-microservice/user-proto/v1"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	AppServiceCharge            int     = 1000
	EwalletServiceCharge        float64 = 0.04
	VirtualAccountServiceCharge int     = 4000
	QrisServiceCharge           float64 = 0.007
)

const OrderTopic = "order-mail-service"

const (
	NewOrder     = "new-order"
	SuccessOrder = "order-succeeded"
	FailedOrder  = "order-failed"
)

type OrderService struct {
	DB             *sql.DB
	Validate       *validator.Validate
	Ctx            context.Context
	Producer       *KafkaProducer
	PaymentService *ppb.PaymentServiceClient
	ProductService *prpb.ProductServiceClient
	UserService    *upb.UserServiceClient
}

func NewOrderService(
	DB *sql.DB,
	validate *validator.Validate,
	producer *KafkaProducer,
	PaymentService *ppb.PaymentServiceClient,
	ProductService *prpb.ProductServiceClient,
	UserService *upb.UserServiceClient,
) *OrderService {
	return &OrderService{DB: DB, Validate: validate, Ctx: context.Background(), Producer: producer, PaymentService: PaymentService, ProductService: ProductService, UserService: UserService}
}

func (o *OrderService) RegisterRoutes(app fiber.Router) {
	app.Get("/orders", o.handleGetOrders)
	app.Get("/orders/:id", o.handleGetOrderById)
	app.Post("/orders", o.handleCreateOrders)

	app.Post("/orders/:id/simulate", shared.DevOnlyMiddleware, o.handleSimulatePayment)

	app.Use(WebhookTokenMiddleware)
	app.Post("/webhook/orders/succeeded", o.handleOrderSucceededWebhook)
	app.Post("/webhook/orders/failed", o.handleOrderFailedWebhook)
}

func (o *OrderService) handleGetOrders(c *fiber.Ctx) error {
	tx, err := o.DB.Begin()
	if err != nil {
		slog.Error("Error occurred while starting transaction", "err", err)
		return err
	}
	defer shared.CommitOrRollback(tx, nil)

	query := `SELECT id, buyer_id, buyer_email, buyer_phone, product_id, product_name, destination, server_id, total_product_amount, service_charge, total_amount, status, failure_code, created_at, updated_at FROM orders`

	rows, err := tx.QueryContext(o.Ctx, query)
	if err != nil {
		slog.Error("Error occurred while querying orders", "err", err)
		return err
	}
	defer rows.Close()

	var orders []Order
	for rows.Next() {
		var order Order
		err := rows.Scan(
			&order.Id,
			&order.BuyerId,
			&order.BuyerEmail,
			&order.BuyerPhone,
			&order.ProductId,
			&order.ProductName,
			&order.Destination,
			&order.ServerId,
			&order.TotalProductAmount,
			&order.ServiceCharge,
			&order.TotalAmount,
			&order.Status,
			&order.FailureCode,
			&order.CreatedAt,
			&order.UpdatedAt,
		)
		if err != nil {
			slog.Error("Error occurred while scanning order row", "err", err)
			return err
		}
		orders = append(orders, order)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Orders retrieved successfully",
		"data":    orders,
		"errors":  nil,
	})
}

func (o *OrderService) handleGetOrderById(c *fiber.Ctx) error {
	orderId := c.Params("id")
	if orderId == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Order ID is required")
	}

	tx, err := o.DB.Begin()
	if err != nil {
		slog.Error("Error occurred while starting transaction", "err", err)
		return err
	}
	defer shared.CommitOrRollback(tx, nil)

	query := `SELECT id, payment_reference_id, buyer_id, buyer_email, buyer_phone, product_id, product_name, destination, server_id, total_product_amount, service_charge, total_amount, status, failure_code, created_at, updated_at FROM orders WHERE id = ?`

	row := tx.QueryRowContext(o.Ctx, query, orderId)

	var order Order
	err = row.Scan(
		&order.Id,
		&order.PaymentReferenceId,
		&order.BuyerId,
		&order.BuyerEmail,
		&order.BuyerPhone,
		&order.ProductId,
		&order.ProductName,
		&order.Destination,

		&order.ServerId,
		&order.TotalProductAmount,
		&order.ServiceCharge,
		&order.TotalAmount,
		&order.Status,
		&order.FailureCode,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Order not found")
		}
		slog.Error("Error occurred while scanning order row", "err", err)
		return err
	}

	getPaymentIdReq := ppb.GetPaymentByIdReq{
		PaymentId: order.PaymentReferenceId,
	}

	getPaymentByIdRes, err := (*o.PaymentService).GetPaymentById(c.Context(), &getPaymentIdReq)
	if err != nil {
		st, ok := status.FromError(err)
		if ok && st.Code() == codes.NotFound {
			slog.Info("payment not found", "payment-reference-id", order.PaymentReferenceId)
			return fiber.NewError(fiber.StatusNotFound, "Order not found")
		}

		slog.Error("Error occurred while calling payment service", "err", err)
		return err
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"message": "Order retrieved successfully",
		"data":    getPaymentByIdRes,
		"errors":  nil,
	})
}

func (o *OrderService) handleCreateOrders(c *fiber.Ctx) error {
	orderRequest := &CreateOrderRequest{}
	orderData := &Order{}

	err := c.BodyParser(orderRequest)
	if err != nil {
		slog.Error("Error occurred while parsing request body", "err", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request")
	}

	err = o.Validate.Struct(orderRequest)
	if err != nil && errors.As(err, &validator.ValidationErrors{}) {
		slog.Error("Validation Error")
		return shared.NewFailedValidationError(*orderRequest, err.(validator.ValidationErrors))
	}

	orderData.Id = uuid.NewString()
	orderData.Destination = orderRequest.Destination
	orderData.ServerId = orderRequest.ServerId
	orderData.BuyerEmail = orderRequest.BuyerEmail

	userServiceErrChan := make(chan error, 1)
	productServiceErrChan := make(chan error, 1)
	defer close(userServiceErrChan)
	defer close(productServiceErrChan)

	var user *upb.User
	var product *prpb.Product

	// get user if logged in
	userChannel := make(chan *upb.User, 1)
	go func() {
		defer close(userChannel)

		if c.Get("Authorization") == "" {
			userServiceErrChan <- nil
			userChannel <- nil
			return
		}

		userId, err := shared.GetUserIdFromToken(c)
		if err != nil {
			userServiceErrChan <- err
			userChannel <- nil
			return
		}

		getUserByIdReq := upb.GetUserByIdReq{Id: userId}

		getUserByIdRes, err := (*o.UserService).GetUserById(c.Context(), &getUserByIdReq)
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.NotFound {
				slog.Info("product not found", "product-id", orderRequest.ProductId)
				userServiceErrChan <- fiber.NewError(fiber.StatusNotFound, st.Message())
				userChannel <- nil
				return
			}

			slog.Error("Error occurred while calling product service", "err", err)
			userServiceErrChan <- err
			userChannel <- nil
			return
		}

		userServiceErrChan <- nil
		userChannel <- getUserByIdRes.GetUser()
	}()

	// get data product
	productChannel := make(chan *prpb.Product, 1)
	go func() {
		defer close(productChannel)

		getProductByIdReq := prpb.GetProductByIdReq{
			ProductId: int32(orderRequest.ProductId),
		}

		getProductByIdRes, err := (*o.ProductService).GetProductById(c.Context(), &getProductByIdReq)
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.NotFound {
				slog.Info("product not found", "product-id", orderRequest.ProductId)
				productServiceErrChan <- fiber.NewError(fiber.StatusNotFound, "Product not found")
				productChannel <- nil
				return
			}

			slog.Error("Error occurred while calling product service", "err", err)
			productServiceErrChan <- err
			productChannel <- nil
			return
		}

		productServiceErrChan <- nil
		productChannel <- getProductByIdRes.Product
	}()

	// wait for user and product service to finish
	if err = <-userServiceErrChan; err != nil {
		return err
	}

	user = <-userChannel
	if user != nil {
		orderData.BuyerId = int(user.Id)
	}

	if err = <-productServiceErrChan; err != nil {
		return err
	}

	product = <-productChannel
	orderData.ProductId = int(product.Id)
	orderData.ProductName = product.Name
	orderData.TotalProductAmount = int(product.Price)

	// set payment method
	paymentMethod, err := getPaymentMethodDetails(orderRequest.PaymentMethod, int(product.Price))
	if err != nil {
		slog.Error("Invalid Channel Code", "err", err)
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	orderData.ServiceCharge = paymentMethod.ServiceCharge
	orderData.TotalAmount = paymentMethod.TotalAmount

	// call create payment
	paymentServiceErrChan := make(chan error, 1)
	defer close(paymentServiceErrChan)
	createPaymentResChan := make(chan *ppb.CreatePaymentRes, 1)
	go func(pm *Order) {
		defer close(createPaymentResChan)

		// create payment
		createPaymentReq := ppb.CreatePaymentReq{
			ReferenceId: orderData.Id,
			ChannelCode: pm.ChannelCode,
			Amount:      int32(orderData.TotalAmount),
			BuyerEmail:  orderData.BuyerEmail,
		}
		if user != nil {
			createPaymentReq.BuyerMobileNumber = user.PhoneNumber
		}

		createPaymentRes, err := (*o.PaymentService).CreatePayment(c.Context(), &createPaymentReq)
		if err != nil {
			st, ok := status.FromError(err)
			if ok && st.Code() == codes.InvalidArgument {
				slog.Info("Invalid payment request", "err", err)
				paymentServiceErrChan <- fiber.NewError(fiber.StatusBadRequest, st.Message())
				createPaymentResChan <- nil
				return
			}

			slog.Error("Error occurred while calling payment service", "err", err)
			paymentServiceErrChan <- fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
			createPaymentResChan <- nil
			return
		}

		paymentServiceErrChan <- nil
		createPaymentResChan <- createPaymentRes
	}(paymentMethod)

	if err = <-paymentServiceErrChan; err != nil {
		return err
	}

	createPaymentRes := <-createPaymentResChan
	orderData.Status = createPaymentRes.Status
	orderData.FailureCode = createPaymentRes.FailureCode
	orderData.PaymentReferenceId = createPaymentRes.XenditPaymentId

	orderData.CreatedAt = time.Now()

	tx, err := o.DB.Begin()
	if err != nil {
		slog.Error("Error occurred while starting transaction", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}
	defer shared.CommitOrRollback(tx, err)

	query := `INSERT INTO orders (id, payment_reference_id, product_id, product_name, destination, server_id, buyer_id, buyer_email,
					buyer_phone, service_charge, channel_code, total_product_amount, total_amount,
					status, failure_code, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`

	result, err := tx.ExecContext(o.Ctx, query,
		orderData.Id,
		orderData.PaymentReferenceId,
		orderData.ProductId,
		orderData.ProductName,
		orderData.Destination,
		orderData.ServerId,
		orderData.BuyerId,
		orderData.BuyerEmail,
		orderData.BuyerPhone,
		orderData.ServiceCharge,
		orderData.ChannelCode,
		orderData.TotalProductAmount,
		orderData.TotalAmount,
		orderData.Status,
		orderData.FailureCode,
		orderData.CreatedAt,
		orderData.CreatedAt, // assuming updated_at is the same as created_at for new orders
	)
	if err != nil {
		slog.Error("Error occurred while inserting order", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	if affectedRows, _ := result.RowsAffected(); affectedRows == 0 {
		slog.Error("No rows affected while inserting order")
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to create order")
	}

	baseMsg := &OrderEvent{
		EventTye: NewOrder,
		Data: &OrderMsg{
			Id:                 orderData.Id,
			Status:             orderData.Status,
			ProductId:          orderData.ProductId,
			ProductName:        orderData.ProductName,
			ProductPrice:       orderData.TotalProductAmount,
			Destination:        orderData.Destination,
			ServerId:           orderData.ServerId,
			ChannelCode:        orderData.ChannelCode,
			BuyerEmail:         orderData.BuyerEmail,
			ServiceCharge:      orderData.ServiceCharge,
			TotalProductAmount: orderData.TotalProductAmount,
			TotalAmount:        orderData.TotalAmount,
			CreatedAt:          orderData.CreatedAt,
		},
	}

	newOrderMsgJson, err := json.Marshal(baseMsg)

	err = o.Producer.Write(o.Ctx, OrderTopic, [2]string{orderData.Id, string(newOrderMsgJson)})
	if err != nil {
		slog.Error("Error occurred while sending new order event", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to send new order event")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Order created successfully",
		"data":    orderData,
		"errors":  nil,
	})
}

func (o *OrderService) handleOrderSucceededWebhook(c *fiber.Ctx) error {
	var request GetResponse[XenditPaymentRequest]
	if err := c.BodyParser(&request); err != nil {
		slog.Error("Error occurred while parsing webhook request", "err", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request")
	}

	webhookRequest := request.Data

	if webhookRequest.Status != "SUCCEEDED" {
		slog.Info("Order payment not succeeded", "status", webhookRequest.Status)
		return c.SendStatus(fiber.StatusOK)
	}

	tx, err := o.DB.Begin()
	if err != nil {
		slog.Error("Error occurred while starting transaction", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}
	defer shared.CommitOrRollback(tx, nil)

	slog.Info(
		"Updating order status",
		"id",
		webhookRequest.ReferenceId,
		"status",
		webhookRequest.Status,
		"failure_code",
		webhookRequest.FailureCode,
	)

	query := `UPDATE orders SET status = ?, failure_code = ? WHERE id = ?`
	result, err := tx.ExecContext(
		o.Ctx,
		query,
		webhookRequest.Status,
		webhookRequest.FailureCode,
		webhookRequest.ReferenceId,
	)
	if err != nil {
		slog.Error(
			"Error occurred while updating order status",
			"err",
			err,
			"id",
			webhookRequest.ReferenceId,
			"status",
			webhookRequest.Status,
		)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	if affectedRows, _ := result.RowsAffected(); affectedRows == 0 {
		slog.Error("No rows affected while updating order status")
		return fiber.NewError(fiber.StatusNotFound, "Order not found")
	}

	query = `SELECT id, product_id, product_name, destination, server_id, service_charge, total_product_amount, total_amount, created_at FROM orders WHERE id = ?`
	row := tx.QueryRowContext(o.Ctx, query, webhookRequest.ReferenceId)
	var order Order
	err = row.Scan(
		&order.Id,
		&order.ProductId,
		&order.ProductName,
		&order.Destination,
		&order.ServerId,
		&order.ServiceCharge,
		&order.TotalProductAmount,
		&order.TotalAmount,
		&order.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("Order not found for payment reference ID", "id", webhookRequest.Id)
			return fiber.NewError(fiber.StatusNotFound, "Order not found")
		}
		slog.Error("Error occurred while scanning order row", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	baseMsg := &OrderEvent{
		EventTye: SuccessOrder,
		Data: &OrderMsg{
			Id:                 order.Id,
			Status:             webhookRequest.Status,
			ProductId:          order.ProductId,
			ProductName:        order.ProductName,
			ProductPrice:       order.TotalProductAmount,
			Destination:        order.Destination,
			ServerId:           order.ServerId,
			BuyerEmail:         order.BuyerEmail,
			ServiceCharge:      order.ServiceCharge,
			TotalProductAmount: order.TotalProductAmount,
			TotalAmount:        order.TotalAmount,
			CreatedAt:          order.CreatedAt,
		},
	}

	successOrderMsgByte, err := json.Marshal(baseMsg)

	err = o.Producer.Write(
		o.Ctx,
		OrderTopic,
		[2]string{webhookRequest.Id, string(successOrderMsgByte)},
	)
	if err != nil {
		slog.Error("Error occurred while sending order succeeded event", "err", err)
		return fiber.NewError(
			fiber.StatusInternalServerError,
			"Failed to send order succeeded event",
		)
	}

	return c.SendStatus(fiber.StatusOK)
}

func (o *OrderService) handleOrderFailedWebhook(c *fiber.Ctx) error {
	var request GetResponse[XenditPaymentRequest]
	if err := c.BodyParser(&request); err != nil {
		slog.Error("Error occurred while parsing webhook request", "err", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request")
	}

	webhookRequest := request.Data

	if webhookRequest.Status != "FAILED" {
		slog.Info("Order payment not failed", "status", webhookRequest.Status)
		return c.SendStatus(fiber.StatusOK)
	}

	tx, err := o.DB.Begin()
	if err != nil {
		slog.Error("Error occurred while starting transaction", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}
	defer shared.CommitOrRollback(tx, nil)

	query := `UPDATE orders SET status = ?, failure_code = ? WHERE id = ?`
	result, err := tx.ExecContext(
		o.Ctx,
		query,
		webhookRequest.Status,
		webhookRequest.FailureCode,
		webhookRequest.ReferenceId,
	)
	if err != nil {
		slog.Error("Error occurred while updating order status", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	if affectedRows, _ := result.RowsAffected(); affectedRows == 0 {
		slog.Error("No rows affected while updating order status")
		return fiber.NewError(fiber.StatusNotFound, "Order not found")
	}

	query = `SELECT id, product_id, product_name, destination, server_id, service_charge, total_product_amount, total_amount, created_at FROM orders WHERE id = ?`
	row := tx.QueryRowContext(o.Ctx, query, webhookRequest.ReferenceId)
	var order Order
	err = row.Scan(
		&order.Id,
		&order.ProductId,
		&order.ProductName,
		&order.Destination,
		&order.ServerId,
		&order.ServiceCharge,
		&order.TotalProductAmount,
		&order.TotalAmount,
		&order.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			slog.Error("Order not found for payment reference ID", "id", webhookRequest.Id)
			return fiber.NewError(fiber.StatusNotFound, "Order not found")
		}
		slog.Error("Error occurred while scanning order row", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	baseMsg := &OrderEvent{
		EventTye: FailedOrder,
		Data: &OrderMsg{
			Id:                 order.Id,
			Status:             webhookRequest.Status,
			FailureCode:        webhookRequest.FailureCode,
			ProductId:          order.ProductId,
			ProductName:        order.ProductName,
			ProductPrice:       order.TotalProductAmount,
			Destination:        order.Destination,
			ServerId:           order.ServerId,
			BuyerEmail:         order.BuyerEmail,
			ServiceCharge:      order.ServiceCharge,
			TotalProductAmount: order.TotalProductAmount,
			TotalAmount:        order.TotalAmount,
			CreatedAt:          order.CreatedAt,
		},
	}

	newOrderMsgJson, err := json.Marshal(baseMsg)

	err = o.Producer.Write(
		o.Ctx,
		OrderTopic,
		[2]string{webhookRequest.Id, string(newOrderMsgJson)},
	)
	if err != nil {
		slog.Error("Error occurred while sending order failed event", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Failed to send order failed event")
	}

	return c.SendStatus(fiber.StatusOK)
}

func (o *OrderService) handleSimulatePayment(c *fiber.Ctx) error {
	orderId := c.Params("id")
	if orderId == "" {
		return fiber.NewError(fiber.StatusBadRequest, "Order ID is required")
	}

	simulateRequest := &SimulatePaymentRequest{}
	if err := c.BodyParser(simulateRequest); err != nil {
		slog.Error("Error occurred while parsing simulate payment request", "err", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request")
	}

	tx, err := o.DB.Begin()
	if err != nil {
		slog.Error("Error occurred while starting transaction", "err", err)
		return err
	}
	defer shared.CommitOrRollback(tx, nil)

	query := `SELECT id, payment_reference_id, buyer_id, buyer_email, buyer_phone, product_id, product_name, channel_code, destination, server_id, total_product_amount, service_charge, total_amount, status, failure_code, created_at, updated_at FROM orders WHERE id = ?`

	row := tx.QueryRowContext(o.Ctx, query, orderId)

	var order Order
	err = row.Scan(
		&order.Id,
		&order.PaymentReferenceId,
		&order.BuyerId,
		&order.BuyerEmail,
		&order.BuyerPhone,
		&order.ProductId,
		&order.ProductName,
		&order.ChannelCode,
		&order.Destination,
		&order.ServerId,
		&order.TotalProductAmount,
		&order.ServiceCharge,
		&order.TotalAmount,
		&order.Status,
		&order.FailureCode,
		&order.CreatedAt,
		&order.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fiber.NewError(fiber.StatusNotFound, "Order not found")
		}
		slog.Error("Error occurred while scanning order row", "err", err)
		return err
	}

	// Get payment request data
	paymentServiceErrChan := make(chan error, 1)
	getPaymentByIdResponse := make(chan *GetPaymentByIdResponse, 1)
	defer close(getPaymentByIdResponse)
	go func() {
		defer close(paymentServiceErrChan)

		paymentServiceHost := os.Getenv("PAYMENT_SERVICE_HOST")
		paymentServicePort := os.Getenv("PAYMENT_SERVICE_PORT")
		url := fmt.Sprintf("/payments/%s", order.PaymentReferenceId)
		paymentServiceResponse, err := shared.CallService(
			paymentServiceHost,
			paymentServicePort,
			url,
			fiber.MethodGet,
			nil,
		)

		if err != nil || len(paymentServiceResponse.Errs) > 0 {
			slog.Error(
				"Error occurred while calling payment service",
				"errs",
				paymentServiceResponse.Errs,
			)
			slog.Error("Error occurred while calling payment service", "err", err)
			paymentServiceErrChan <- fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
			return
		}

		if paymentServiceResponse.StatusCode != fiber.StatusOK {
			slog.Error(
				"payment service returned non-200 status code",
				"code",
				paymentServiceResponse.StatusCode,
			)
			paymentServiceErrChan <- fiber.NewError(paymentServiceResponse.StatusCode, string(paymentServiceResponse.Body))
			return
		}

		var paymentResponse GetResponse[GetPaymentByIdResponse]
		err = json.Unmarshal(paymentServiceResponse.Body, &paymentResponse)
		if err != nil {
			slog.Error("Error occurred while unmarshalling payment response", "err", err)
			paymentServiceErrChan <- fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
			return
		}

		getPaymentByIdResponse <- &paymentResponse.Data
		paymentServiceErrChan <- nil
	}()

	if err = <-paymentServiceErrChan; err != nil {
		return err
	}

	paymentResponse := <-getPaymentByIdResponse

	if paymentResponse.Status == "SUCCEEDED" {
		return c.JSON(fiber.Map{
			"message": "Payment already succeeded",
			"data":    paymentResponse,
			"errors":  nil,
		})
	}

	if paymentResponse.Status == "FAILED" {
		return c.JSON(fiber.Map{
			"message": "Payment already failed",
			"data":    paymentResponse,
			"errors":  nil,
		})
	}

	// Determine the payment request type. There are "redirect" and "http call"
	if paymentResponse.Actions[0].Type == "REDIRECT_CUSTOMER" {
		err := handleEwalletPaymentSimulation(paymentResponse.Actions[0].Value)
		if err != nil {
			slog.Error("Error occurred while handling ewallet payment simulation", "err", err)
			return err
		}
	} else {
		err = handleOthersPaymentSimulation(order.PaymentReferenceId, paymentResponse.RequestAmount)
		if err != nil {
			slog.Error("Error occured while handling payment simulation", "err", err)
			return err
		}
	}

	return c.JSON(fiber.Map{
		"message": "Payment simulated successfully",
		"data":    paymentResponse,
		"errors":  nil,
	})
}

func handleEwalletPaymentSimulation(urlAction string) error {
	parsedPaymentUrl, err := urllib.Parse(urlAction)
	if err != nil {
		slog.Error("Error occurred while parsing ewallet payment url", "err", err)
		return err
	}

	paymentToken := parsedPaymentUrl.Query().Get("token")
	if paymentToken == "" {
		slog.Error("Ewallet payment token not found in url")
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	urlPayment := fmt.Sprintf(
		"https://ewallet-mock-connector.xendit.co/v1/ewallet_connector/payment_callbacks?token=%s",
		paymentToken,
	)
	agent := fiber.Post(urlPayment).Timeout(15 * time.Second)

	statusCode, body, errs := agent.Bytes()

	if len(errs) > 0 {
		slog.Error("Error occurred while calling payment service", "err", errs)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	if statusCode != fiber.StatusOK {
		slog.Error("failed to simulate payment", "code", statusCode)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	xenditSimulationResp := SimulateXenditResponse{}
	if err := json.Unmarshal(body, &xenditSimulationResp); err != nil {
		slog.Error("Error occurred while unmarshalling ewallet payment response", "err", err)
		return err
	}

	if xenditSimulationResp.Status == "SUCCEEDED" {
		return nil
	}

	return fiber.NewError(fiber.StatusExpectationFailed, "Payment failed. Please check callback for failure reason.")
}

func handleOthersPaymentSimulation(prId string, amount int) error {
	xenditApiKey := os.Getenv("XENDIT_API_KEY") + ":"
	xenditApiKeyBase64 := base64.StdEncoding.EncodeToString([]byte(xenditApiKey))
	xenditBaseUrl := os.Getenv("XENDIT_API_URL")
	paymentSimulationUrl := fmt.Sprintf("%s/v3/payment_requests/%s/simulate", xenditBaseUrl, prId)

	agent := fiber.Post(paymentSimulationUrl).Timeout(15*time.Second).
		Add("Authorization", fmt.Sprintf("Basic %s", xenditApiKeyBase64)).
		Add("api-version", "2024-11-11").
		JSON(fiber.Map{
			"amount": amount,
		})

	statusCode, respByte, errs := agent.Bytes()
	if len(errs) > 0 {
		slog.Error(
			"Error occurred while calling xendit payment simulation api",
			"err",
			errs,
			"resp",
			string(respByte),
		)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	if statusCode != fiber.StatusOK {
		slog.Error("failed to simulate payment", "code", statusCode, "resp", string(respByte))
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	return nil
}

func getPaymentMethodDetails(channelCode string, productPrice int) (*Order, error) {
	// Check if the payment method is valid
	for _, channel := range EwalletChannelCodes {
		if channelCode == channel {
			// implementation for creating ewallet payment
			serviceCharge := (float64(productPrice) * EwalletServiceCharge) + float64(AppServiceCharge)
			totalAmount := math.Ceil(float64(productPrice) + serviceCharge)
			return &Order{
				ChannelCode:   channelCode,
				TotalAmount:   int(totalAmount),
				ServiceCharge: serviceCharge,
			}, nil
		}
	}

	for _, channel := range VirtualAccountChannelCodes {
		if channelCode == channel {
			// implementation for creating virtual account payment
			serviceCharge := float64(VirtualAccountServiceCharge) + float64(AppServiceCharge)
			totalAmount := math.Ceil(float64(productPrice) + serviceCharge)
			return &Order{
				ChannelCode:   channelCode,
				TotalAmount:   int(totalAmount),
				ServiceCharge: serviceCharge,
			}, nil
		}
	}

	for _, channel := range QrisChannelCode {
		if channelCode == channel {
			// implementation for creating qris payment
			serviceCharge := (float64(productPrice) * QrisServiceCharge) + float64(AppServiceCharge)
			totalAmount := math.Ceil(float64(productPrice) + serviceCharge)
			return &Order{
				ChannelCode:   channelCode,
				TotalAmount:   int(totalAmount),
				ServiceCharge: serviceCharge,
			}, nil
		}
	}

	return nil, errors.New("invalid channel code")
}
