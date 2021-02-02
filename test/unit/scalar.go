package unit

type ArrayOfStrings []string
type ArrayOfCitrus []Citrus

type ObjectMap map[string]string
type ObjectCitrus map[string]Citrus

type Release struct {
	MultipleOf10  int64   `json:"multiple_of_10" multipleOf:"10"`
	MultipleOf5PC float64 `json:"multiple_of_5_pc" multipleOf:"0.2"`
	RangeInt      int64   `json:"range_int" description:"Range between 1% and 100%" minimum:"1" maximum:"100" example:"3"`
	RangeFloat    float64 `json:"range_float" minimum:"0.01" maximum:"0.5" example:"0.2"`
	Description   string  `json:"description" minLength:"30" maxLength:"255" exclusiveMinimum:"true" exclusiveMaximum:"true" example:"any text over 30 characters"`
	Version       string  `json:"version" pattern:"^(?P<major>0|[1-9][0-9]*)\\.(?P<minor>0|[1-9][0-9]*)\\.(?P<patch>0|[1-9][0-9]*)(?:-(?P<prerelease>(?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$" example:"1.0.0-release+1.0.0"`
	Deprecated    string  `json:"deprecated,-"`
	Required      string  `json:"required,required"`
	GoasOnly      string  `json:"goas_only" goas:"-"`
}
