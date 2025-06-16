package main

import "time"

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

type XenditPaymentRequestResponse struct {
	Id            string              `json:"id" validate:"required"`
	ReferenceId   string              `json:"reference_id" validate:"required"`
	Status        string              `json:"status" validate:"required"`
	Amount        int                 `json:"amount" validate:"required"`
	Country       string              `json:"country" validate:"required"`
	Currency      string              `json:"currency" validate:"required"`
	PaymentMethod XenditPaymentMethod `json:"payment_method" validate:"required"`
	Actions       []EwalletActions    `json:"actions" validate:"required"`
	Created       time.Time           `json:"created" validate:"required"`
	Updated       time.Time           `json:"updated" validate:"required"`
	FailureCode   string              `json:"failure_code"`
}

type XenditCustomerIndividualDetail struct {
	GivenNames string `json:"given_names" validate:"required"`
}

type XenditCustomer struct {
	ReferenceId      string                         `json:"reference_id" validate:"required"`
	Type             string                         `json:"type" validate:"required"`
	Email            string                         `json:"email" validate:"required,email"`
	IndividualDetail XenditCustomerIndividualDetail `json:"individual_detail" validate:"required"`
}

type EwalletChannelProperties struct {
	MobileNumber     string `json:"mobile_number,omitempty" validate:"required"`
	SuccessReturnUrl string `json:"success_return_url,omitempty" validate:"required"`
}

type Ewallet struct {
	ChannelCode       string                   `json:"channel_code" validate:"required"`
	ChannelProperties EwalletChannelProperties `json:"channel_properties" validate:"required"`
}

type VirtualAccountChannelProperties struct {
	CustomerName         string    `json:"customer_name" validate:"required"`
	ExpiresAt            time.Time `json:"expires_at" validate:"required"`
	VirtualAccountNumber string    `json:"virtual_account_number,omitempty"`
}

type VirtualAccount struct {
	ChannelCode       string                          `json:"channel_code" validate:"required"`
	ChannelProperties VirtualAccountChannelProperties `json:"channel_properties,omitempty" validate:"required"`
}

type QrCodeChannelProperties struct {
	ExpiresAt time.Time `json:"expires_at" validate:"required"`
	QrString  string    `json:"qr_string,omitempty"`
}

type QrCode struct {
	QrCodeChannelProperties QrCodeChannelProperties `json:"channel_properties" validate:"required"`
}

type XenditPaymentMethod struct {
	Type           string          `json:"type" validate:"required"`
	Reusability    string          `json:"reusability" validate:"required"`
	Ewallet        *Ewallet        `json:"ewallet,omitempty"`
	VirtualAccount *VirtualAccount `json:"virtual_account,omitempty"`
	QrCode         *QrCode         `json:"qr_code,omitempty"`
}

type XenditRequestBody struct {
	Currency      string              `json:"currency" validate:"required"`
	Amount        int                 `json:"amount" validate:"required,min=1"`
	ReferenceId   string              `json:"reference_id" validate:"required"`
	Customer      XenditCustomer      `json:"customer" validate:"required"`
	PaymentMethod XenditPaymentMethod `json:"payment_method" validate:"required"`
}

type EwalletActions struct {
	Action  string `json:"action" validate:"required"`
	Url     string `json:"url" validate:"required"`
	UrlType string `json:"url_type" validate:"required"`
	Method  string `json:"method" validate:"required"`
}

type VirtualAccountActions struct {
	VirtualAccountNumber string `json:"virtual_account_number" validate:"required"`
}

type QrCodeActions struct {
	QrCodeString string `json:"qr_code_string" validate:"required"`
}

type PaymentActions struct {
	Ewallet        *EwalletActions        `json:"ewallet"`
	VirtualAccount *VirtualAccountActions `json:"virtual_account"`
	QrCode         *QrCodeActions         `json:"qr_code"`
}

type GetPaymentByIdResponse struct {
	XenditPaymentId string         `json:"xendit_payment_id" validate:"required"`
	Amount          int            `json:"amount" validate:"required"`
	Status          string         `json:"status" validate:"required"`
	Actions         PaymentActions `json:"actions" validate:"required"`
	Created         time.Time      `json:"created" validate:"required"`
	Updated         time.Time      `json:"updated" validate:"required"`
}

type XenditErrMsg struct {
	ErrorCode string `json:"error_code"`
	Message   string `json:"message"`
}
