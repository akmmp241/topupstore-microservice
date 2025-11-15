package main

type CategorySearch struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type OperatorSearch struct {
	ID       int    `json:"id"`
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	ImageUrl string `json:"image_url"`
}

type ProductTypeSearch struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProductSearch struct {
	ID          int               `json:"id"`
	Name        string            `json:"name"`
	Price       int               `json:"price"`
	ImageUrl    string            `json:"image_url"`
	Category    CategorySearch    `json:"category"`
	Operator    OperatorSearch    `json:"operator"`
	ProductType ProductTypeSearch `json:"product_type"`
}

type EsResponse struct {
	Hits struct {
		Total struct {
			Value int `json:"value"`
		} `json:"total"`
		Hits []struct {
			Source ProductSearch `json:"_source"`
		} `json:"hits"`
	} `json:"hits"`
}
