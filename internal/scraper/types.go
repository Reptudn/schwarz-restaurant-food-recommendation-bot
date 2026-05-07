package scraper

type Price struct {
	Label string // "Intern" or "Extern"
	Value string // "4,00 €"
}

type Allergen struct {
	Code string // "A", "A1", "C", etc.
	Name string // "Cereals containing gluten"
}

type Additive struct {
	Code string // "3"
	Name string // "with antioxidants"
}

type Nutrition struct {
	EnergyKj       string `json:"energy_kj"`        // "2343 kJ" / 100g
	EnergyKcal     string `json:"energy_kcal"`      // "559 kcal" / 100g
	Protein       string `json:"protein"`         // "26g" / 100g
	Fat           string `json:"fat"`             // "33g" / 100g
	Carbohydrates string `json:"carbohydrates"`    // "42g" / 100g
}

type MenuItem struct {
	Category string
	Name     string
	Prices   []Price
	Tags     []string // "Vegan", "Vegetarisch", "Glutenfrei", etc.
	Detail   *MenuItemDetail
}

type MenuItemDetail struct {
	ProductName  string     `json:"product_name"`
	Allergens   []Allergen  `json:"allergens"`
	SubAllergens []string   `json:"sub_allergens"` // "A1", "A2"...
	Additives   []Additive  `json:"additives"`
	Traces     []string   `json:"traces"`     // trace names
	Nutrition  *Nutrition `json:"nutrition"`
	PDFLink    string    `json:"pdf_link"`
}

type DayMenu struct {
	Date  string // "2026-05-07"
	Items []MenuItem
}