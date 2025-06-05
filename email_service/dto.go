package main

import "time"

type SendMail struct {
	To      string
	Subject string
	Body    string
}

type NewLoginMessage struct {
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	LoginTime time.Time `json:"login_time"`
	IpAddress string    `json:"ip_address"`
	Device    string    `json:"device"`
}

type NewRegistrationMessage struct {
	Name            string `json:"name"`
	VerificationUrl string `json:"verification_url"`
	Email           string `json:"email"`
}

type ForgotPasswordMessage struct {
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	ResetUrl  string    `json:"reset_url"`
	ExpiresAt time.Time `json:"expired_at"`
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
