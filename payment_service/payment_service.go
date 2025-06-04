package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/akmmp241/topupstore-microservice/shared"
	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v2"
	"github.com/xendit/xendit-go/v4/payment_method"
	"log/slog"
	"os"
	"time"
)

type PaymentService struct {
	DB        *sql.DB
	Validator *validator.Validate
	Ctx       context.Context
}

func NewPaymentService(DB *sql.DB, validator *validator.Validate) *PaymentService {
	return &PaymentService{DB: DB, Validator: validator, Ctx: context.Background()}
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

	if err := p.Validator.Struct(paymentRequest); err != nil && errors.As(err, &validator.ValidationErrors{}) {
		return shared.NewFailedValidationError(*paymentRequest, err.(validator.ValidationErrors))
	}

	xenditRequestBody := XenditRequestBody{
		Currency:    "IDR",
		Amount:      paymentRequest.Amount,
		ReferenceId: paymentRequest.ReferenceId,
		Customer: XenditCustomer{
			ReferenceId: fmt.Sprintf("%s@%s", paymentRequest.ReferenceId, paymentRequest.BuyerEmail),
			Type:        "INDIVIDUAL",
			Email:       paymentRequest.BuyerEmail,
			IndividualDetail: XenditCustomerIndividualDetail{
				GivenNames: fmt.Sprintf("Topupstore-customer %s", paymentRequest.BuyerEmail),
			},
		},
		PaymentMethod: XenditPaymentMethod{
			Type:           paymentRequest.PaymentMethodId,
			Reusability:    "ONE_TIME_USE",
			Ewallet:        nil,
			VirtualAccount: nil,
			QrCode:         nil,
		},
	}

	// Check if the payment method is valid
	for _, method := range payment_method.AllowedPaymentMethodTypeEnumValues {
		if paymentRequest.PaymentMethodId != string(method) {
			continue
		}

		// implementation for creating ewallet payment
		if channel, err := payment_method.NewEWalletChannelCodeFromValue(paymentRequest.PaymentMethodName); channel != nil && err == nil {
			xenditRequestBody.PaymentMethod.Ewallet = &Ewallet{
				ChannelCode: channel.String(),
				ChannelProperties: EwalletChannelProperties{
					MobileNumber:     paymentRequest.BuyerMobileNumber,
					SuccessReturnUrl: "https://example.com/success",
				},
			}
			break
		}

		// implementation for creating virtual account payment
		if channel, err := payment_method.NewVirtualAccountChannelCodeFromValue(paymentRequest.PaymentMethodName); channel != nil && err == nil {
			xenditRequestBody.PaymentMethod.VirtualAccount = &VirtualAccount{
				ChannelCode: channel.String(),
				ChannelProperties: VirtualAccountChannelProperties{
					CustomerName: fmt.Sprintf("Topupstore-customer %s", paymentRequest.BuyerEmail),
					ExpiresAt:    time.Now().Add(24 * time.Hour),
				},
			}
			break
		}

		// implementation for creating QR code payment
		if channel, err := payment_method.NewQRCodeChannelCodeFromValue(paymentRequest.PaymentMethodName); channel != nil && err == nil {
			xenditRequestBody.PaymentMethod.QrCode = &QrCode{
				QrCodeChannelProperties: QrCodeChannelProperties{
					ExpiresAt: time.Now().Add(24 * time.Hour),
				},
			}
			break
		}
	}

	bytes, _ := json.Marshal(xenditRequestBody)
	slog.Info("Request body", "requestBody", string(bytes))

	slog.Info("Creating payment request with Xendit", "requestBody", xenditRequestBody.PaymentMethod.VirtualAccount)

	xenditApiKey := os.Getenv("XENDIT_API_KEY") + ":"
	xenditApiKeyBase64 := base64.StdEncoding.EncodeToString([]byte(xenditApiKey))
	xenditHost := os.Getenv("XENDIT_API_URL")
	paymentReqUrl := fmt.Sprintf("%s/payment_requests", xenditHost)

	agent := fiber.Post(paymentReqUrl).Timeout(15*time.Second).
		Add("Authorization", fmt.Sprintf("Basic %s", xenditApiKeyBase64)).
		ContentType(fiber.MIMEApplicationJSON).JSON(xenditRequestBody)

	statusCode, respByte, errs := agent.Bytes()

	if len(errs) > 0 {
		slog.Error("Error occurred while calling xendit payment request api", "errs", errs, "resp", string(respByte))
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	if statusCode >= 300 {
		slog.Error("xendit payment request api returned non-200 status code", "code", statusCode, "resp", string(respByte))
		return fiber.NewError(statusCode, string(respByte))
	}

	var paymentRequestResponse XenditPaymentRequestResponse
	err = json.Unmarshal(respByte, &paymentRequestResponse)
	if err != nil {
		slog.Error("Error occurred while unmarshalling xendit payment request api response", "err", err)
		return fiber.NewError(fiber.StatusInternalServerError, "Internal Server Error")
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"message": "Payment created successfully",
		"data": &CreatePaymentResponse{
			XenditPaymentId: paymentRequestResponse.Id,
			Status:          paymentRequestResponse.Status,
			FailureCode:     paymentRequestResponse.FailureCode,
		},
		"errors": nil,
	})
}

func (p *PaymentService) GetPayment(c *fiber.Ctx) error {
	// Implementation for getting a payment
	return c.SendString("Get Payment")
}
