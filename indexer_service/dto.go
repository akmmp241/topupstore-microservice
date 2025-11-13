package main

type BaseEvent[T Product] struct {
	EventType string `json:"event_type"`
	Data      T      `json:"data"`
}

type Category struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Operator struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	ImageUrl string `json:"image_url"`
}

type ProductType struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type Product struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Price       int         `json:"price"`
	ImageUrl    string      `json:"image_url"`
	Category    Category    `json:"category"`
	Operator    Operator    `json:"operator"`
	ProductType ProductType `json:"product_type"`
}
