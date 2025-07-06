package main

import "time"

type CreatePaymentRequest struct {
	ReferenceId       string  `json:"reference_id" validate:"required"`
	ChannelCode       string  `json:"channel_code" validate:"required"`
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

type XenditPaymentRequestResponse struct {
	ReferenceId       string            `json:"reference_id" validate:"required"`
	PaymentRequestId  string            `json:"payment_request_id" validate:"required"`
	PaymentTokenId    string            `json:"payment_token_id" validate:"required"`
	LatestPaymentId   string            `json:"latest_payment_id" validate:"required"`
	Type              string            `json:"type" validate:"required"`
	RequestAmount     int               `json:"request_amount" validate:"required"`
	CaptureMethod     string            `json:"capture_method" validate:"required"`
	ChannelCode       string            `json:"channel_code" validate:"required"`
	ChannelProperties ChannelProperties `json:"channel_properties" validate:"required"`
	Country           string            `json:"country" validate:"required"`
	Currency          string            `json:"currency" validate:"required"`
	Actions           []Action          `json:"actions" validate:"required"`
	Status            string            `json:"status" validate:"required"`
	FailureCode       string            `json:"failure_code"`
	Created           time.Time         `json:"created" validate:"required"`
	Updated           time.Time         `json:"updated" validate:"required"`
}

type ChannelProperties struct {
	DisplayName      string    `json:"DisplayName,omitempty"`
	ExpiresAt        time.Time `json:"expires_at,omitempty"`
	SuccessReturnUrl string    `json:"success_return_url,omitempty"`
	FailureReturnUrl string    `json:"failure_return_url,omitempty"`
	CancelReturnUrl  string    `json:"cancel_return_url,omitempty"`
}

type XenditRequestBody struct {
	Currency          string            `json:"currency" validate:"required"`
	RequestAmount     int               `json:"request_amount" validate:"required,min=1"`
	Country           string            `json:"country"`
	ReferenceId       string            `json:"reference_id" validate:"required"`
	ChannelCode       string            `json:"channel_code"`
	ChannelProperties ChannelProperties `json:"channel_properties"`
}

type Action struct {
	Type       string `json:"type" validate:"required"`
	Descriptor string `json:"descriptor" validate:"required"`
	Value      string `json:"value" validate:"required"`
}

type GetPaymentByIdResponse struct {
	PaymentRequestId  string            `json:"payment_request_id" validate:"required"`
	RequestAmount     int               `json:"request_amount" validate:"required"`
	ChannelCode       string            `json:"channel_code" validate:"required"`
	ChannelProperties ChannelProperties `json:"channel_properties" validate:"required"`
	Actions           []Action          `json:"actions" validate:"required"`
	Status            string            `json:"status" validate:"required"`
	FailureCode       string            `json:"failure_code"`
	Created           time.Time         `json:"created" validate:"required"`
	Updated           time.Time         `json:"updated" validate:"required"`
}

type XenditErrMsg struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}
