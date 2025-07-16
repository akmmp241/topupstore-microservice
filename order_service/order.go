package main

import "time"

type Order struct {
	Id                 string    `json:"id"`
	PaymentReferenceId string    `json:"payment_reference_id"`
	BuyerId            int       `json:"buyer_id"`
	BuyerEmail         string    `json:"buyer_email"`
	BuyerPhone         string    `json:"buyer_phone"`
	ProductId          int       `json:"product_id"`
	ProductName        string    `json:"product_name"`
	Destination        string    `json:"destination"`
	ServerId           string    `json:"server_id"`
	ChannelCode        string    `json:"channel_code"`
	TotalProductAmount int       `json:"total_product_amount"`
	ServiceCharge      float64   `json:"service_charge"`
	TotalAmount        int       `json:"total_amount"`
	Status             string    `json:"status"`
	FailureCode        string    `json:"failure_code"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}
