package main

import "time"

type Product struct {
	Id            int       `json:"id"`
	RefId         string    `json:"ref_id"`
	ProductTypeId int       `json:"product_type_id"`
	Name          string    `json:"name"`
	Code          string    `json:"code"`
	Description   string    `json:"description"`
	ImageUrl      string    `json:"image_url"`
	Price         float64   `json:"price"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
