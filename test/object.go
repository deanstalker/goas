package test

// EnumProperties ...
// @Title Enumerator Properties
// @Description test to ensure enums are handled
type EnumProperties struct {
	Status    string `json:"status" enum:"active,pending,disabled"`
	ErrorCode int64  `json:"error_code" enum:"400,404,500"`
}

// LimitedObjectProperties ...
type LimitedObjectProperties struct {
	Properties map[string]Citrus `json:"properties" minProperties:"2" maxProperties:"5"`
}
