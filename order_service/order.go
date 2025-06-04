package main

import "time"

type Order struct {
	Id                 string    `json:"id"`
	BuyerId            int       `json:"buyer_id"`
	BuyerEmail         string    `json:"buyer_email"`
	BuyerPhone         string    `json:"buyer_phone"`
	ProductId          int       `json:"product_id"`
	ProductName        string    `json:"product_name"`
	Destination        string    `json:"destination"`
	ServerId           string    `json:"server_id"`
	PaymentMethodId    string    `json:"payment_method_id"`
	PaymentMethodName  string    `json:"payment_method_name"`
	TotalProductAmount float64   `json:"total_product_amount"`
	ServiceCharge      float64   `json:"service_charge"`
	TotalAmount        float64   `json:"total_amount"`
	Status             string    `json:"status"`
	FailureCode        string    `json:"failure_code"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
