package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"os"
	"strings"
	"time"

	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
)

type PaymentService struct {
	Validator *validator.Validate
	Ctx       context.Context
}

func NewPaymentService(validator *validator.Validate) *PaymentService {
	return &PaymentService{Validator: validator, Ctx: context.Background()}
}

func (p *PaymentService) RegisterRoutes(app fiber.Router) {
	app.Use(shared.JWTServiceMiddleware)
	app.Post("/payments", p.CreatePayment)
	app.Get("/payments/:id", p.GetPayment)
}

func (p *PaymentService) CreatePayment(c *fiber.Ctx) error {
	paymentRequest := &CreatePaymentRequest{}

	err := c.BodyParser(paymentRequest)
	if err != nil {
		slog.Error("Failed to parse request body", "error", err)
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := p.Validator.Struct(paymentRequest); err != nil &&
		errors.As(err, &validator.ValidationErrors{}) {
		return shared.NewFailedValidationError(*paymentRequest, err.(validator.ValidationErrors))
	}

	xenditRequestBody := XenditRequestBody{
		Currency:      "IDR",
		Country:       "ID",
		Type:          "PAY",
		CaptureMethod: "AUTOMATIC",
		RequestAmount: int(math.Ceil(paymentRequest.Amount)),
		ReferenceId:   paymentRequest.ReferenceId,
	}

	displayName := strings.Replace(paymentRequest.BuyerEmail, "@", "..", 1)
	xenditRequestBody.ChannelProperties = ChannelProperties{
		DisplayName:      displayName,
		ExpiresAt:        time.Now().Add(time.Hour),
		SuccessReturnUrl: "https://www.xendit.co/success",
		FailureReturnUrl: "https://www.xendit.co/failure",
		CancelReturnUrl:  "https://www.xendit.co/cancel",
	}

	// Check if the payment method is valid
	for _, channel := range EwalletChannelCodes {
		if paymentRequest.ChannelCode != channel {
			continue
		}

		// implementation for creating ewallet payment
		xenditRequestBody.ChannelCode = channel
	}

	for _, channel := range VirtualAccountChannelCodes {
		if paymentRequest.ChannelCode != channel {
			continue
		}

		// implementation for creating virtual account payment
		xenditRequestBody.ChannelCode = channel
	}

	for _, channel := range QrisChannelCode {
		if paymentRequest.ChannelCode != channel {
			continue
		}

		// implementation for creating qris payment
		xenditRequestBody.ChannelCode = channel
	}

	// check if channel code is valid
	if xenditRequestBody.ChannelCode == "" {
		slog.Error("Channel code is not valid", "channel_code", paymentRequest.ChannelCode)
		return fiber.NewError(fiber.StatusBadRequest, "Channel code is not valid")
	}

	xenditApiKey := os.Getenv("XENDIT_API_KEY") + ":"
	xenditApiKeyBase64 := base64.StdEncoding.EncodeToString([]byte(xenditApiKey))
	xenditHost := os.Getenv("XENDIT_API_URL")
	paymentReqUrl := fmt.Sprintf("%s/v3/payment_requests", xenditHost)

	agent := fiber.Post(paymentReqUrl).Timeout(15*time.Second).
		Add("Authorization", fmt.Sprintf("Basic %s", xenditApiKeyBase64)).
		Add("api-version", "2024-11-11").
		ContentType(fiber.MIMEApplicationJSON).JSON(xenditRequestBody)

	statusCode, respByte, errs := agent.Bytes()

	if len(errs) > 0 {
		slog.Error(
			"Error occurred while calling xendit payment request api",
			"errs",
			errs,
			"resp",
			string(respByte),
		)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	if statusCode >= 300 {

		var errMsg XenditErrMsg
		err := json.Unmarshal(respByte, &errMsg)
		if err != nil {
			return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
		}
		slog.Error(
			"xendit payment request api returned non-200 status code",
			"code",
			statusCode,
			"resp",
			string(respByte),
		)
		return fiber.NewError(statusCode, errMsg.Message)
	}

	var paymentRequestResponse XenditPaymentRequestResponse
	err = json.Unmarshal(respByte, &paymentRequestResponse)
	if err != nil {
		slog.Error(
			"Error occurred while unmarshalling xendit payment request api response",
			"err",
			err,
		)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Payment created successfully",
		"data": &CreatePaymentResponse{
			XenditPaymentId: paymentRequestResponse.PaymentRequestId,
			Status:          paymentRequestResponse.Status,
			FailureCode:     paymentRequestResponse.FailureCode,
		},
		"errors": nil,
	})
}

func (p *PaymentService) GetPayment(c *fiber.Ctx) error {
	paymentId := c.Params("id")
	if paymentId == "" {
		slog.Error("Payment ID is required", "error", "Payment ID cannot be empty")
		return fiber.NewError(fiber.StatusBadRequest, "Payment ID cannot be empty")
	}

	getPaymentErrChan := make(chan error, 1)
	defer close(getPaymentErrChan)

	paymentResponseChan := make(chan *XenditPaymentRequestResponse, 1)
	go func() {
		defer close(paymentResponseChan)

		xenditApiKey := os.Getenv("XENDIT_API_KEY") + ":"
		xenditApiKeyBase64 := base64.StdEncoding.EncodeToString([]byte(xenditApiKey))
		xenditHost := os.Getenv("XENDIT_API_URL")
		paymentReqUrl := fmt.Sprintf("%s/v3/payment_requests/%s", xenditHost, paymentId)

		agent := fiber.Get(paymentReqUrl).Timeout(15*time.Second).
			Add("Authorization", fmt.Sprintf("Basic %s", xenditApiKeyBase64)).
			Add("api-version", "2024-11-11")

		statusCode, respByte, errs := agent.Bytes()

		if len(errs) > 0 {
			slog.Error(
				"Error occurred while calling xendit payment request api",
				"errs",
				errs,
				"resp",
				string(respByte),
			)
			getPaymentErrChan <- fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
			return
		}

		if statusCode >= 300 {

			var errMsg XenditErrMsg
			err := json.Unmarshal(respByte, &errMsg)
			if err != nil {
				getPaymentErrChan <- fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
				return
			}
			slog.Error(
				"xendit payment request api returned non-200 status code",
				"code",
				statusCode,
				"resp",
				string(respByte),
			)
			getPaymentErrChan <- fiber.NewError(statusCode, errMsg.Message)
		}

		var paymentRequestResponse XenditPaymentRequestResponse
		err := json.Unmarshal(respByte, &paymentRequestResponse)
		if err != nil {
			slog.Error(
				"Error occurred while unmarshalling xendit payment request api response",
				"err",
				err,
			)
			getPaymentErrChan <- fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
		}

		paymentResponseChan <- &paymentRequestResponse
		getPaymentErrChan <- nil
	}()

	if err := <-getPaymentErrChan; err != nil {
		slog.Error("Error occurred while getting payment", "error", err)
		return err
	}

	paymentResponse := <-paymentResponseChan

	getPaymentByIdResponse := &GetPaymentByIdResponse{
		PaymentRequestId:  paymentResponse.PaymentRequestId,
		Status:            paymentResponse.Status,
		RequestAmount:     paymentResponse.RequestAmount,
		ChannelCode:       paymentResponse.ChannelCode,
		ChannelProperties: paymentResponse.ChannelProperties,
		FailureCode:       paymentResponse.FailureCode,
		Created:           paymentResponse.Created,
		Updated:           paymentResponse.Updated,
		Actions:           paymentResponse.Actions,
	}

	return c.JSON(fiber.Map{
		"message": "Payment retrieved successfully",
		"data":    &getPaymentByIdResponse,
		"errors":  nil,
	})
}
