package main

import "time"

type CreateOrderRequest struct {
	Destination   string `json:"destination" validate:"required"`
	ServerId      string `json:"server_id"`
	ProductId     string `json:"product_id" validate:"required"`
	PaymentMethod string `json:"payment_method" validate:"required"`
	BuyerEmail    string `json:"buyer_email" validate:"required"`
}

type CreatePaymentRequest struct {
	ReferenceId       string  `json:"reference_id" validate:"required"`
	PaymentMethodId   string  `json:"payment_method_id" validate:"required"`
	PaymentMethodName string  `json:"payment_method_name" validate:"required"`
	Amount            float64 `json:"amount" validate:"required,min=1"`
	BuyerEmail        string  `json:"buyer_email" validate:"required,email"`
	BuyerMobileNumber string  `json:"buyer_mobile_number" validate:"omitempty"`
}

type CreatePaymentResponse struct {
	XenditPaymentId string `json:"xendit_payment_id" validate:"required"`
	Status          string `json:"status" validate:"required"`
	FailureCode     string `json:"failure_code"`
}

type NewOrderMsg struct {
	Id                 string    `json:"id" validate:"required"`
	ProductId          int       `json:"product_id" validate:"required"`
	ProductName        string    `json:"product_name" validate:"required"`
	ProductPrice       float64   `json:"product_price" validate:"required,min=1"`
	Destination        string    `json:"destination" validate:"required"`
	ServerId           string    `json:"server_id"`
	PaymentMethodName  string    `json:"payment_method_name" validate:"required"`
	PaymentMethodId    string    `json:"payment_method_id" validate:"required"`
	BuyerEmail         string    `json:"buyer_email" validate:"required,email"`
	ServiceCharge      float64   `json:"service_charge" validate:"required,min=0"`
	TotalProductAmount float64   `json:"total_product_amount" validate:"required,min=1"`
	TotalAmount        float64   `json:"total_amount" validate:"required,min=1"`
	CreatedAt          time.Time `json:"created_at" validate:"required"`
}

type GetResponse[T interface{}] struct {
	Message string `json:"message"`
	Data    T      `json:"data"`
	Errors  any    `json:"errors"`
}
