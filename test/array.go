package test

// FruitBasketArray ...
type FruitBasketArray struct {
	Fruit []Fruit `json:"fruit" minItems:"5" maxItems:"10" uniqueItems:"true"`
}

// Fruit ...
type Fruit struct {
	Color   string `json:"color"`
	HasSeed bool   `json:"has_seed"`
}
