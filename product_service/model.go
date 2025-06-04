package main

import "time"

type Category struct {
	Id        int       `json:"id"`
	RefId     string    `json:"ref_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
type Operator struct {
	Id          int       `json:"id"`
	RefId       string    `json:"ref_id"`
	CategoryId  int       `json:"category_id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	ImageUrl    string    `json:"image_url"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ProductType struct {
	Id         int       `json:"id"`
	RefId      string    `json:"ref_id"`
	OperatorId int       `json:"operator_id"`
	Name       string    `json:"name"`
	FormatForm string    `json:"format_form"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Product struct {
	Id            int       `json:"id"`
	RefId         string    `json:"ref_id"`
	ProductTypeId int       `json:"product_type_id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	ImageUrl      string    `json:"image_url"`
	Price         int       `json:"price"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
