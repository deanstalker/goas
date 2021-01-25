package test

type Release struct {
	RangeInt    int64   `json:"range_int" minimum:"1" maximum:"100"`
	RangeFloat  float64 `json:"range_float" minimum:"0.01" maximum:"0.5"`
	Description string  `json:"description" minLength:"30" maxLength:"255" exclusiveMinimum:"true" exclusiveMaximum:"true"`
	Version     string  `json:"version" pattern:"^(?P<major>0|[1-9][0-9]*)\\.(?P<minor>0|[1-9][0-9]*)\\.(?P<patch>0|[1-9][0-9]*)(?:-(?P<prerelease>(?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(?:\\.(?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))?(?:\\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\\.[0-9a-zA-Z-]+)*))?$"`
}
