package test

// FruitOneOfAKind ...
// @Title One of a kind Fruit
// @Description only one kind of fruit at a time
type FruitOneOfAKind struct {
	Kind interface{} `json:"kind" oneOf:"test.Citrus,test.Banana"`
}

// FruitOneOfAKindDisc ...
// @Title One of a kind Fruit with Discriminator
// @Description only one kind of fruit at a time
type FruitOneOfAKindDisc struct {
	Kind interface{} `json:"kind" oneOf:"test.Citrus,test.Banana" discriminator:"kind"`
}

// FruitOneOfAKindInvalidDisc ...
// @Title One of a kind Fruit with Invalid Discriminator
// @Description only one kind of fruit at a time
type FruitOneOfAKindInvalidDisc struct {
	Kind interface{} `json:"kind" oneOf:"test.Citrus,test.Banana" discriminator:"kindle"`
}

// FruitAllOfAKind ...
// @Title All of a kind
// @Description only all of a kind of fruit at a time
type FruitAllOfAKind struct {
	Kind interface{} `json:"kind" allOf:"test.Citrus,test.Banana"`
}

// FruitAnyOfAKind ...
// @Title Any of a kind
// @Description any kind of fruit
type FruitAnyOfAKind struct {
	Kind interface{} `json:"kind" anyOf:"test.Citrus,test.Banana"`
}

// Citrus ...
type Citrus struct {
	Kind string `json:"kind"`
}

// Banana ...
type Banana struct {
	Kind string `json:"kind"`
}
