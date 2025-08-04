package main

import "time"

type CreateOrderRequest struct {
	Destination   string `json:"destination"    validate:"required"`
	ServerId      string `json:"server_id"`
	ProductId     int    `json:"product_id"     validate:"required"`
	PaymentMethod string `json:"payment_method" validate:"required"`
	BuyerEmail    string `json:"buyer_email"    validate:"required"`
}

type CreatePaymentRequest struct {
	ReferenceId       string `json:"reference_id"        validate:"required"`
	ChannelCode       string `json:"channel_code"        validate:"required"`
	Amount            int    `json:"amount"              validate:"required,min=1"`
	BuyerEmail        string `json:"buyer_email"         validate:"required,email"`
	BuyerMobileNumber string `json:"buyer_mobile_number" validate:"omitempty"`
}

type CreatePaymentResponse struct {
	XenditPaymentId string `json:"xendit_payment_id" validate:"required"`
	Status          string `json:"status"            validate:"required"`
	FailureCode     string `json:"failure_code"`
}

type OrderMsg struct {
	Id                 string    `json:"id"                   validate:"required"`
	Status             string    `json:"status"               validate:"required"`
	FailureCode        string    `json:"failure_code"`
	ProductId          int       `json:"product_id"           validate:"required"`
	ProductName        string    `json:"product_name"         validate:"required"`
	ProductPrice       int       `json:"product_price"        validate:"required,min=1"`
	Destination        string    `json:"destination"          validate:"required"`
	ServerId           string    `json:"server_id"            validate:"required"`
	ChannelCode        string    `json:"channel_code"         validate:"required"`
	BuyerEmail         string    `json:"buyer_email"          validate:"required,email"`
	ServiceCharge      float64   `json:"service_charge"       validate:"required,min=0"`
	TotalProductAmount int       `json:"total_product_amount" validate:"required,min=1"`
	TotalAmount        int       `json:"total_amount"         validate:"required,min=1"`
	CreatedAt          time.Time `json:"created_at"           validate:"required"`
}

type GetResponse[T any] struct {
	Message string `json:"message"`
	Data    T      `json:"data"`
	Errors  any    `json:"errors"`
}

type EwalletActions struct {
	Action  string `json:"action"   validate:"required"`
	Url     string `json:"url"      validate:"required"`
	UrlType string `json:"url_type" validate:"required"`
	Method  string `json:"method"   validate:"required"`
}

type VirtualAccountActions struct {
	VirtualAccountNumber string `json:"virtual_account_number" validate:"required"`
}

type QrCodeActions struct {
	QrCodeString string `json:"qr_code_string" validate:"required"`
}

type PaymentActions struct {
	Ewallet        *EwalletActions        `json:"ewallet,omitempty"`
	VirtualAccount *VirtualAccountActions `json:"virtual_account,omitempty"`
	QrCode         *QrCodeActions         `json:"qr_code,omitempty"`
}

type ChannelProperties struct {
	DisplayName      string    `json:"DisplayName,omitempty"`
	ExpiresAt        time.Time `json:"expires_at,omitempty"`
	SuccessReturnUrl string    `json:"success_return_url,omitempty"`
	FailureReturnUrl string    `json:"failure_return_url,omitempty"`
	CancelReturnUrl  string    `json:"cancel_return_url,omitempty"`
}

type Action struct {
	Type       string `json:"type"       validate:"required"`
	Descriptor string `json:"descriptor" validate:"required"`
	Value      string `json:"value"      validate:"required"`
}

type GetPaymentByIdResponse struct {
	PaymentRequestId  string            `json:"payment_request_id" validate:"required"`
	RequestAmount     int               `json:"request_amount"     validate:"required"`
	ChannelCode       string            `json:"channel_code"       validate:"required"`
	ChannelProperties ChannelProperties `json:"channel_properties" validate:"required"`
	Actions           []Action          `json:"actions"            validate:"required"`
	Status            string            `json:"status"             validate:"required"`
	FailureCode       string            `json:"failure_code"`
	Created           time.Time         `json:"created"            validate:"required"`
	Updated           time.Time         `json:"updated"            validate:"required"`
}

type XenditPaymentRequest struct {
	Id            string              `json:"id"             validate:"required"`
	ReferenceId   string              `json:"reference_id"   validate:"required"`
	Status        string              `json:"status"         validate:"required"`
	Amount        int                 `json:"amount"         validate:"required"`
	Country       string              `json:"country"        validate:"required"`
	Currency      string              `json:"currency"       validate:"required"`
	PaymentMethod XenditPaymentMethod `json:"payment_method" validate:"required"`
	Actions       []EwalletActions    `json:"actions"        validate:"required"`
	Created       time.Time           `json:"created"        validate:"required"`
	Updated       time.Time           `json:"updated"        validate:"required"`
	FailureCode   string              `json:"failure_code"`
}

type EwalletChannelProperties struct {
	MobileNumber     string `json:"mobile_number,omitempty"      validate:"required"`
	SuccessReturnUrl string `json:"success_return_url,omitempty" validate:"required"`
}

type Ewallet struct {
	ChannelCode       string                   `json:"channel_code"       validate:"required"`
	ChannelProperties EwalletChannelProperties `json:"channel_properties" validate:"required"`
}

type VirtualAccountChannelProperties struct {
	CustomerName         string    `json:"customer_name"                    validate:"required"`
	ExpiresAt            time.Time `json:"expires_at"                       validate:"required"`
	VirtualAccountNumber string    `json:"virtual_account_number,omitempty"`
}

type VirtualAccount struct {
	ChannelCode       string                          `json:"channel_code"                 validate:"required"`
	ChannelProperties VirtualAccountChannelProperties `json:"channel_properties,omitempty" validate:"required"`
}

type QrCodeChannelProperties struct {
	ExpiresAt time.Time `json:"expires_at"          validate:"required"`
	QrString  string    `json:"qr_string,omitempty"`
}

type QrCode struct {
	QrCodeChannelProperties QrCodeChannelProperties `json:"channel_properties" validate:"required"`
}

type XenditPaymentMethod struct {
	Type           string          `json:"type"                      validate:"required"`
	Reusability    string          `json:"reusability"               validate:"required"`
	Ewallet        *Ewallet        `json:"ewallet,omitempty"`
	VirtualAccount *VirtualAccount `json:"virtual_account,omitempty"`
	QrCode         *QrCode         `json:"qr_code,omitempty"`
}

type SimulatePaymentRequest struct {
	Amount int `json:"amount" validate:"required"`
}

type SimulateXenditResponse struct {
	Status string `json:"status"`
}
