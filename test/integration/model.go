package integration

type Pet struct {
	ID   int64  `json:"id" yaml:"id"`
	Name string `json:"name" yaml:"name"`
	Tag  string `json:"tag" yaml:"tag"`
}

type Pets []Pet

type Error struct {
	Code    int32  `json:"code" yaml:"code"`
	Message string `json:"message" yaml:"message"`
}
