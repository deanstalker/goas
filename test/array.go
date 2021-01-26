package test

// FruitBasketArray ...
type FruitBasketArray struct {
	Fruit []Fruit `json:"fruit" minItems:"5" maxItems:"10" uniqueItems:"true" example:"[{\"color\":\"red\",\"has_seed\":\"true\"}]"`
}

// Fruit ...
type Fruit struct {
	Color   string `json:"color" example:"red"`
	HasSeed bool   `json:"has_seed" example:"true"`
}
