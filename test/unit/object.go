package unit

// EnumProperties enumerated properties
// @Title Enumerator Properties
// @Description test to ensure enums are handled
type EnumProperties struct {
	Status    string `json:"status" enum:"active,pending,disabled"`
	ErrorCode int64  `json:"error_code" enum:"400,404,500"`
}

// LimitedObjectProperties properties with applied limits
type LimitedObjectProperties struct {
	Properties map[string]Citrus `json:"properties" minProperties:"2" maxProperties:"5" example:"{\"orange\":{\"kind\":\"citrus\"}}"`
}
