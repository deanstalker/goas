package test

// @Title Post Request Body
// @Description A wrapper for incoming content
// PostRequest ...
type PostRequestBody struct {
	Content string      `json:"content" maxLength:"255" minLength:"1"`
	Percent int64       `json:"percent" multipleOf:"10"`
	Range   int64       `json:"range" minimum:"1" maximum:"255" exclusiveMinimum:"true" exclusiveMaximum:"true"`
	Generic interface{} `json:"generic" oneOf:"test.ColourType,test.PatternType"`
}

// @Title Colours of the content
type ColourType struct {
}

// @Title Patterns of the content
type PatternType struct {
}
