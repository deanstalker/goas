package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"os"
	"sort"
	"testing"

	"github.com/iancoleman/orderedmap"

	"gopkg.in/yaml.v2"

	"github.com/deanstalker/goas/pkg/types"

	"github.com/deanstalker/goas/internal/util"

	"github.com/stretchr/testify/assert"
)

func TestParseParamComment(t *testing.T) {
	dir, _ := os.Getwd()
	modulePath := util.ModulePath("./")
	pkgName, _ := modulePath.Get()
	tests := map[string]struct {
		pkgPath    string
		pkgName    string
		comment    string
		wantOp     *types.OperationObject
		wantSchema map[string]*types.SchemaObject
		expectErr  error
	}{
		"string param in path": {
			pkgPath: dir,
			pkgName: "main",
			comment: `locale   path   string   true   "Locale code"`,
			wantOp: &types.OperationObject{
				Parameters: []types.ParameterObject{
					{
						Name:        "locale",
						In:          "path",
						Description: "Locale code",
						Required:    true,
						Example:     nil,
						Schema: &types.SchemaObject{
							Type:   "string",
							Format: "string",
						},
					},
				},
			},
			wantSchema: make(map[string]*types.SchemaObject),
			expectErr:  nil,
		},
		"string param in path without desc": {
			pkgPath: dir,
			pkgName: "main",
			comment: `locale   path   string   true`,
			wantOp: &types.OperationObject{
				Parameters: []types.ParameterObject{
					{
						Name:        "locale",
						In:          "path",
						Description: "Locale code",
						Required:    true,
						Example:     nil,
						Schema: &types.SchemaObject{
							Type:   "string",
							Format: "string",
						},
					},
				},
			},
			wantSchema: make(map[string]*types.SchemaObject),
			expectErr:  errors.New(`parseParamComment can not parse param comment "locale   path   string   true"`),
		},
		"string in body": {
			pkgPath: dir,
			pkgName: "main",
			comment: `firstname   body   string   true   "First Name"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Type: "string",
							},
						},
					},
					Required: true,
				},
			},
			wantSchema: make(map[string]*types.SchemaObject),
			expectErr:  nil,
		},
		"[]string in body": {
			pkgPath: dir,
			pkgName: "main",
			comment: `address   body   []string   true   "Address"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Type: "array",
								Items: &types.SchemaObject{
									Type: "string",
								},
							},
						},
					},
					Required: true,
				},
			},
			wantSchema: make(map[string]*types.SchemaObject),
			expectErr:  nil,
		},
		"map[]string in body": {
			pkgPath: dir,
			pkgName: "main",
			comment: `address   body   map[]string   true   "Address"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Type: "object",
								Properties: types.NewOrderedMap().
									Set("address", &types.SchemaObject{
										Type: "string",
									}),
							},
						},
					},
					Required: true,
				},
			},
			wantSchema: make(map[string]*types.SchemaObject),
			expectErr:  nil,
		},
		"timestamp in path": {
			pkgPath: dir,
			pkgName: "main",
			comment: `time   path   time.Time   true   "Timestamp"`,
			wantOp: &types.OperationObject{
				Parameters: []types.ParameterObject{
					{
						Name:        "time",
						In:          "path",
						Description: "Timestamp",
						Required:    true,
						Example:     nil,
						Schema: &types.SchemaObject{
							Type:   "string",
							Format: "date-time",
						},
					},
				},
			},
			wantSchema: make(map[string]*types.SchemaObject),
			expectErr:  nil,
		},
		"file in body": {
			pkgPath: dir,
			pkgName: "main",
			comment: `image file ignored true "Image upload"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeForm: {
							Schema: types.SchemaObject{
								Type: "object",
								Properties: types.NewOrderedMap().
									Set("image", &types.SchemaObject{
										Type:        "string",
										Format:      "binary",
										Description: "Image upload",
									}),
							},
						},
					},
					Required: true,
				},
			},
			wantSchema: make(map[string]*types.SchemaObject),
			expectErr:  nil,
		},
		"files in body": {
			pkgPath: dir,
			pkgName: "main",
			comment: `image files string true "Image upload"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeForm: {
							Schema: types.SchemaObject{
								Type: "object",
								Properties: types.NewOrderedMap().
									Set("image", &types.SchemaObject{
										Type:        "array",
										Description: "Image upload",
										Items:       &types.SchemaObject{Type: "string", Format: "binary"},
									}),
							},
						},
					},
					Description: "",
					Required:    true,
					Ref:         "",
				},
			},
			wantSchema: make(map[string]*types.SchemaObject),
			expectErr:  nil,
		},
		"form field with string in body": {
			pkgPath: dir,
			pkgName: "main",
			comment: `content form string false "Content field"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeForm: {
							Schema: types.SchemaObject{
								Type: "object",
								Properties: types.NewOrderedMap().
									Set("content", &types.SchemaObject{
										Type:        "string",
										Format:      "string",
										Description: "Content field",
									}),
							},
						},
					},
					Description: "",
					Required:    false,
					Ref:         "",
				},
			},
			wantSchema: make(map[string]*types.SchemaObject),
			expectErr:  nil,
		},
		"struct in body": {
			pkgPath: dir,
			pkgName: "main",
			comment: `externaldocs body ExternalDocumentationObject false "External Documentation"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Ref: "#/components/schemas/ExternalDocumentationObject",
							},
						},
					},
				},
			},
			wantSchema: map[string]*types.SchemaObject{
				"ExternalDocumentationObject": {
					ID:                 "ExternalDocumentationObject",
					PkgName:            fmt.Sprintf("%s/pkg/types", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Properties: types.NewOrderedMap().
						Set("description", &types.SchemaObject{
							FieldName: "Description",
							Type:      "string",
						}).
						Set("url", &types.SchemaObject{
							FieldName: "URL",
							Type:      "string",
						}),
				},
			},
			expectErr: nil,
		},
		"array of structs in body": {
			pkgPath: dir,
			pkgName: "main",
			comment: `externaldocs body []ExternalDocumentationObject false "External Documentation"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Type: "array",
								Items: &types.SchemaObject{
									Ref: "#/components/schemas/ExternalDocumentationObject",
								},
							},
						},
					},
				},
			},
			wantSchema: map[string]*types.SchemaObject{
				"ExternalDocumentationObject": {
					ID:                 "ExternalDocumentationObject",
					PkgName:            fmt.Sprintf("%s/pkg/types", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Properties: types.NewOrderedMap().
						Set("description", &types.SchemaObject{
							FieldName: "Description",
							Type:      "string",
						}).
						Set("url", &types.SchemaObject{
							FieldName: "URL",
							Type:      "string",
						}),
				},
			},
			expectErr: nil,
		},
		"map of structs in body": {
			pkgPath: dir,
			pkgName: "main",
			comment: `externaldocs body map[]ExternalDocumentationObject false "External Documentation"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Type: "object",
								Properties: types.NewOrderedMap().
									Set("externaldocs", &types.SchemaObject{
										ID:      "ExternalDocumentationObject",
										PkgName: fmt.Sprintf("%s/pkg/types", pkgName),
										Type:    "object",
										Properties: types.NewOrderedMap().
											Set("description", &types.SchemaObject{
												FieldName: "Description",
												Type:      "string",
											}).
											Set("url", &types.SchemaObject{
												FieldName: "URL",
												Type:      "string",
											}),
										Ref:                "",
										DisabledFieldNames: map[string]struct{}{},
									}),
							},
						},
					},
				},
			},
			wantSchema: map[string]*types.SchemaObject{
				"ExternalDocumentationObject": {
					ID:                 "ExternalDocumentationObject",
					PkgName:            fmt.Sprintf("%s/pkg/types", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Properties: types.NewOrderedMap().
						Set("description", &types.SchemaObject{
							FieldName: "Description",
							Type:      "string",
						}).
						Set("url", &types.SchemaObject{
							FieldName: "URL",
							Type:      "string",
						}),
				},
			},
			expectErr: nil,
		},
		"struct in alternate package - test oneOf a kind": {
			pkgPath: dir,
			pkgName: "test",
			comment: `post body unit.FruitOneOfAKind false "Fruit - Test oneOf a Kind"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Ref: "#/components/schemas/FruitOneOfAKind",
							},
						},
					},
				},
			},
			wantSchema: map[string]*types.SchemaObject{
				"Banana": {
					ID:                 "Banana",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							Type:      "string",
						}),
				},
				"Citrus": {
					ID:                 "Citrus",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							Type:      "string",
						}),
				},
				"FruitOneOfAKind": {
					ID:                 "FruitOneOfAKind",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Title:              "One of a kind Fruit",
					Description:        "only one kind of fruit at a time",
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							OneOf: []*types.ReferenceObject{
								{
									Ref: "#/components/schemas/Citrus",
								},
								{
									Ref: "#/components/schemas/Banana",
								},
							},
						}),
				},
			},
			expectErr: nil,
		},
		// "struct in alternate package - test oneOf a kind - invalid type: {}"
		"struct in alternate package - test oneOf a kind with discriminator": {
			pkgPath: dir,
			pkgName: "test",
			comment: `post body unit.FruitOneOfAKindDisc false "Fruit - Test oneOf a Kind"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Ref: "#/components/schemas/FruitOneOfAKindDisc",
							},
						},
					},
				},
			},
			wantSchema: map[string]*types.SchemaObject{
				"Banana": {
					ID:                 "Banana",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							Type:      "string",
						}),
				},
				"Citrus": {
					ID:                 "Citrus",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							Type:      "string",
						}),
				},
				"FruitOneOfAKindDisc": {
					ID:                 "FruitOneOfAKindDisc",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Title:              "One of a kind Fruit with Discriminator",
					Description:        "only one kind of fruit at a time",
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							OneOf: []*types.ReferenceObject{
								{
									Ref: "#/components/schemas/Citrus",
								},
								{
									Ref: "#/components/schemas/Banana",
								},
							},
							Discriminator: &types.Discriminator{
								PropertyName: "kind",
							},
						}),
				},
			},
			expectErr: nil,
		},
		"struct in alternate package - test oneOf a kind with invalid discriminator": {
			pkgPath: dir,
			pkgName: "test",
			comment: `post body unit.FruitOneOfAKindInvalidDisc false "Fruit - Test oneOf a Kind"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Ref: "#/components/schemas/FruitOneOfAKindInvalidDisc",
							},
						},
					},
				},
			},
			wantSchema: map[string]*types.SchemaObject{
				"Banana": {
					ID:                 "Banana",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							Type:      "string",
						}),
				},
				"Citrus": {
					ID:                 "Citrus",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							Type:      "string",
						}),
				},
				"FruitOneOfAKindInvalidDisc": {
					ID:                 "FruitOneOfAKindInvalidDisc",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Title:              "One of a kind Fruit with Discriminator",
					Description:        "only one kind of fruit at a time",
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							OneOf: []*types.ReferenceObject{
								{
									Ref: "#/components/schemas/Citrus",
								},
								{
									Ref: "#/components/schemas/Banana",
								},
							},
							Discriminator: &types.Discriminator{
								PropertyName: "kind",
							},
						}),
				},
			},
			expectErr: fmt.Errorf("unable to find discriminator field: kindle, in schema: Citrus"),
		},
		"struct in alternate package - test allOf a kind": {
			pkgPath: dir,
			pkgName: "test",
			comment: `post body unit.FruitAllOfAKind false "Fruit - Test allOf a Kind"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Ref: "#/components/schemas/FruitAllOfAKind",
							},
						},
					},
				},
			},
			wantSchema: map[string]*types.SchemaObject{
				"Banana": {
					ID:                 "Banana",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							Type:      "string",
						}),
				},
				"Citrus": {
					ID:                 "Citrus",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							Type:      "string",
						}),
				},
				"FruitAllOfAKind": {
					ID:                 "FruitAllOfAKind",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Title:              "All of a kind",
					Description:        "only all of a kind of fruit at a time",
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							AllOf: []*types.ReferenceObject{
								{
									Ref: "#/components/schemas/Citrus",
								},
								{
									Ref: "#/components/schemas/Banana",
								},
							},
						}),
				},
			},
			expectErr: nil,
		},
		// "struct in alternate package - test allOf a kind - invalid type: {}"
		"struct in alternate package - test anyOf a kind": {
			pkgPath: dir,
			pkgName: "test",
			comment: `post body unit.FruitAnyOfAKind false "Fruit - Test anyOf a Kind"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Ref: "#/components/schemas/FruitAnyOfAKind",
							},
						},
					},
				},
			},
			wantSchema: map[string]*types.SchemaObject{
				"Banana": {
					ID:                 "Banana",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							Type:      "string",
						}),
				},
				"Citrus": {
					ID:                 "Citrus",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							Type:      "string",
						}),
				},
				"FruitAnyOfAKind": {
					ID:                 "FruitAnyOfAKind",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Title:              "Any of a kind",
					Description:        "any kind of fruit",
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							AnyOf: []*types.ReferenceObject{
								{
									Ref: "#/components/schemas/Citrus",
								},
								{
									Ref: "#/components/schemas/Banana",
								},
							},
						}),
				},
			},
			expectErr: nil,
		},
		// "struct in alternate package - test anyOf a kind - invalid type: {}"
		"test enum - string and numeric": {
			pkgPath: dir,
			pkgName: "test",
			comment: `post body unit.EnumProperties false "Enum Properties"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Ref: "#/components/schemas/EnumProperties",
							},
						},
					},
				},
			},
			wantSchema: map[string]*types.SchemaObject{
				"EnumProperties": {
					ID:                 "EnumProperties",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Title:              "Enumerator Properties",
					Description:        "test to ensure enums are handled",
					Properties: types.NewOrderedMap().
						Set("status", &types.SchemaObject{
							FieldName: "Status",
							Type:      "string",
							Enum: []string{
								"active",
								"pending",
								"disabled",
							},
						}).
						Set("error_code", &types.SchemaObject{
							FieldName: "ErrorCode",
							Type:      "integer",
							Enum: []string{
								"400",
								"404",
								"500",
							},
						}),
				},
			},
			expectErr: nil,
		},
		"test object - limited properties": {
			pkgPath: dir,
			pkgName: "test",
			comment: `post body unit.LimitedObjectProperties false "Limited Object Properties"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Ref: "#/components/schemas/LimitedObjectProperties",
							},
						},
					},
				},
			},
			wantSchema: map[string]*types.SchemaObject{
				"Citrus": {
					ID:                 "Citrus",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							Type:      "string",
						}),
				},
				"LimitedObjectProperties": {
					ID:                 "LimitedObjectProperties",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Properties: types.NewOrderedMap().
						Set("properties", &types.SchemaObject{
							FieldName:          "Properties",
							DisabledFieldNames: nil,
							Type:               "object",
							Properties: types.NewOrderedMap().
								Set("key", &types.SchemaObject{
									ID:                 "Citrus",
									PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
									Type:               "object",
									DisabledFieldNames: make(map[string]struct{}),
									Properties: types.NewOrderedMap().
										Set("kind", &types.SchemaObject{
											FieldName: "Kind",
											Type:      "string",
										}),
								}),
							MinProperties: 2,
							MaxProperties: 5,
							Example: map[string]interface{}{
								"orange": map[string]interface {
								}{"kind": "citrus"},
							},
						}),
				},
			},
			expectErr: nil,
		},
		"test array - min, max and unique": {
			pkgPath: "test",
			pkgName: fmt.Sprintf("%s/test/unit", pkgName),
			comment: `post body unit.FruitBasketArray true "Fruit Basket"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Ref: "#/components/schemas/FruitBasketArray",
							},
						},
					},
					Required: true,
				},
			},
			wantSchema: map[string]*types.SchemaObject{
				"Fruit": {
					ID:                 "Fruit",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Properties: types.NewOrderedMap().
						Set("color", &types.SchemaObject{
							FieldName: "Color",
							Type:      "string",
							Example:   "red",
						}).
						Set("has_seed", &types.SchemaObject{
							FieldName: "HasSeed",
							Type:      "boolean",
							Example:   true,
						}),
				},
				"FruitBasketArray": {
					ID:                 "FruitBasketArray",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					DisabledFieldNames: make(map[string]struct{}),
					Type:               "object",
					Properties: types.NewOrderedMap().
						Set("fruit", &types.SchemaObject{
							FieldName: "Fruit",
							Type:      "array",
							Items: &types.SchemaObject{
								Ref: "#/components/schemas/Fruit",
							},
							MinItems:    5,
							MaxItems:    10,
							UniqueItems: true,
							Example: []interface{}{
								map[string]interface{}{
									"color":    "red",
									"has_seed": "true",
								},
							},
						}),
				},
			},
			expectErr: nil,
		},
		"test scalar": {
			pkgPath: "test",
			pkgName: fmt.Sprintf("%s/test/unit", pkgName),
			comment: `post body unit.Release true "Release"`,
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Ref: "#/components/schemas/Release",
							},
						},
					},
					Required: true,
				},
			},
			wantSchema: map[string]*types.SchemaObject{
				"Release": {
					ID:      "Release",
					PkgName: fmt.Sprintf("%s/test/unit", pkgName),
					Type:    "object",
					DisabledFieldNames: map[string]struct{}{
						"deprecated": {},
						"GoasOnly":   {},
					},
					Required: []string{
						"Required",
					},
					Properties: types.NewOrderedMap().
						Set("multiple_of_10", &types.SchemaObject{
							FieldName:  "MultipleOf10",
							Type:       "integer",
							MultipleOf: 10,
						}).
						Set("multiple_of_5_pc", &types.SchemaObject{
							FieldName:  "MultipleOf5PC",
							Type:       "number",
							MultipleOf: 0.2,
						}).
						Set("range_int", &types.SchemaObject{
							FieldName:   "RangeInt",
							Type:        "integer",
							Minimum:     1,
							Maximum:     100,
							Example:     3,
							Description: "Range between 1% and 100%",
						}).
						Set("range_float", &types.SchemaObject{
							FieldName: "RangeFloat",
							Type:      "number",
							Minimum:   0.01,
							Maximum:   0.5,
							Example:   0.2,
						}).
						Set("description", &types.SchemaObject{
							FieldName:        "Description",
							Type:             "string",
							ExclusiveMinimum: true,
							ExclusiveMaximum: true,
							MaxLength:        255,
							MinLength:        30,
							Example:          "any text over 30 characters",
						}).
						Set("version", &types.SchemaObject{
							FieldName: "Version",
							Type:      "string",
							Pattern: `^(?P<major>0|[1-9][0-9]*)\.(?P<minor>0|[1-9][0-9]*)\.(?P<patch>0|[1-9][0-9]*)` +
								`(?:-(?P<prerelease>(?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9][0-9]*|[0-9]*[a-zA-Z-][0-9a-zA-Z-]*))*))` +
								`?(?:\+(?P<buildmetadata>[0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`,
							Example: "1.0.0-release+1.0.0",
						}).
						Set("Required", &types.SchemaObject{
							FieldName: "Required",
							Type:      "string",
						}),
				},
			},
			expectErr: nil,
		},
		"test custom array type - basic": {
			pkgPath: "test",
			pkgName: fmt.Sprintf("%s/test/unit", pkgName),
			comment: "post body unit.ArrayOfStrings true \"Array Of Strings\"",
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Ref: "#/components/schemas/ArrayOfStrings",
							},
						},
					},
					Required: true,
				},
			},
			wantSchema: map[string]*types.SchemaObject{
				"ArrayOfStrings": {
					ID:      "ArrayOfStrings",
					PkgName: fmt.Sprintf("%s/test/unit", pkgName),
					Type:    "array",
					Items: &types.SchemaObject{
						Type: "string",
					},
				},
			},
			expectErr: nil,
		},
		"test custom array type - object": {
			pkgPath: "test",
			pkgName: fmt.Sprintf("%s/test/unit", pkgName),
			comment: "post body unit.ArrayOfCitrus true \"Array Of Citrus\"",
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Ref: "#/components/schemas/ArrayOfCitrus",
							},
						},
					},
					Required: true,
				},
			},
			wantSchema: map[string]*types.SchemaObject{
				"ArrayOfCitrus": {
					ID:      "ArrayOfCitrus",
					PkgName: fmt.Sprintf("%s/test/unit", pkgName),
					Type:    "array",
					Items: &types.SchemaObject{
						Ref: "#/components/schemas/Citrus",
					},
				},
				"Citrus": {
					ID:                 "Citrus",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					Type:               "object",
					DisabledFieldNames: make(map[string]struct{}),
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							Type:      "string",
						}),
				},
			},
			expectErr: nil,
		},
		"test custom map type - basic": {
			pkgPath: "test",
			pkgName: fmt.Sprintf("%s/test/unit", pkgName),
			comment: "post body unit.ObjectMap true \"Object Map - String Values\"",
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Ref: "#/components/schemas/ObjectMap",
							},
						},
					},
					Required: true,
				},
			},
			wantSchema: map[string]*types.SchemaObject{
				"ObjectMap": {
					ID:      "ObjectMap",
					PkgName: fmt.Sprintf("%s/test/unit", pkgName),
					Type:    "object",
					Properties: types.NewOrderedMap().
						Set("key", &types.SchemaObject{
							Type: "string",
						}),
				},
			},
			expectErr: nil,
		},
		"test custom map type - object": {
			pkgPath: "test",
			pkgName: fmt.Sprintf("%s/test/unit", pkgName),
			comment: "post body unit.ObjectCitrus true \"Object Citrus - String Values\"",
			wantOp: &types.OperationObject{
				RequestBody: &types.RequestBodyObject{
					Content: map[string]*types.MediaTypeObject{
						types.ContentTypeJSON: {
							Schema: types.SchemaObject{
								Ref: "#/components/schemas/ObjectCitrus",
							},
						},
					},
					Required: true,
				},
			},
			wantSchema: map[string]*types.SchemaObject{
				"ObjectCitrus": {
					ID:      "ObjectCitrus",
					PkgName: fmt.Sprintf("%s/test/unit", pkgName),
					Type:    "object",
					Properties: types.NewOrderedMap().
						Set("key", &types.SchemaObject{
							Ref: "#/components/schemas/Citrus",
						}),
				},
				"Citrus": {
					ID:                 "Citrus",
					PkgName:            fmt.Sprintf("%s/test/unit", pkgName),
					Type:               "object",
					DisabledFieldNames: make(map[string]struct{}),
					Properties: types.NewOrderedMap().
						Set("kind", &types.SchemaObject{
							FieldName: "Kind",
							Type:      "string",
						}),
				},
			},
			expectErr: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p, err := partialBootstrap()
			if err != nil {
				t.Errorf("%v", err)
			}

			op := &types.OperationObject{}
			if err := p.parseParamComment(tc.pkgPath, tc.pkgName, op, tc.comment); err != nil {
				assert.Equal(t, tc.expectErr, err)
				return
			}

			assert.Equal(t, tc.wantOp, op)
			assert.Equal(t, tc.wantSchema, p.OpenAPI.Components.Schemas)
		})
	}
}

func TestParseServerVariableComments(t *testing.T) {
	tests := map[string]struct {
		comment string
		server  types.ServerObject
		want    map[string]types.ServerVariableObject
	}{
		"test without enum": {
			comment: `username "empty" "Enter a username for dev testing"`,
			server: types.ServerObject{
				URL:         "https://api.{username}.dev.lan/",
				Description: "",
				Variables:   make(map[string]types.ServerVariableObject),
			},
			want: map[string]types.ServerVariableObject{
				"username": {
					Enum:        nil,
					Default:     "empty",
					Description: "Enter a username for dev testing",
				},
			},
		},
		"test with enum": {
			comment: `port "80" "Enter a server port" "80,443,8443,8080"`,
			server: types.ServerObject{
				URL:         "https://api.{port}.dev.lan/",
				Description: "",
				Variables:   make(map[string]types.ServerVariableObject),
			},
			want: map[string]types.ServerVariableObject{
				"port": {
					Enum:        []string{"80", "443", "8443", "8080"},
					Default:     "80",
					Description: "Enter a server port",
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p, err := partialBootstrap()
			if err != nil {
				t.Errorf("%v", err)
			}

			parsed, err := p.parseServerVariableComment(tc.comment, tc.server)
			if err != nil {
				t.Errorf("%v", err)
			}
			assert.Equal(t, tc.want, parsed)
		})
	}
}

func TestParseTagComments(t *testing.T) {
	tests := map[string]struct {
		comment string
		want    types.TagObject
	}{
		"test @tag without externaldocs": {
			comment: `test-service "this is a test service"`,
			want: types.TagObject{
				Name:        "test-service",
				Description: "this is a test service",
			},
		},
		"test @tag with externaldocs": {
			comment: `test-service "this is a test service" https://docs.io  "External Docs"`,
			want: types.TagObject{
				Name:        "test-service",
				Description: "this is a test service",
				ExternalDocs: &types.ExternalDocumentationObject{
					Description: "External Docs",
					URL:         "https://docs.io",
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p, err := partialBootstrap()
			if err != nil {
				t.Errorf("%v", err)
			}

			tag, err := p.parseTagComment(tc.comment)
			if err != nil {
				t.Errorf("%v", err)
			}

			assert.Equal(t, tc.want.Description, tag.Description)
			assert.Equal(t, tc.want.Name, tag.Name)
			assert.Equal(t, tc.want.ExternalDocs, tag.ExternalDocs)
		})
	}
}

func TestParseInfo(t *testing.T) {
	tests := map[string]struct {
		comments  []string
		want      types.InfoObject
		expectErr error
	}{
		"minimum required info": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
			},
			want: types.InfoObject{
				Title:          "Test Run",
				Description:    "This is a test",
				TermsOfService: "",
				Contact:        nil,
				License:        nil,
				Version:        "1.0.0",
			},
			expectErr: nil,
		},
		"partially populated contact and license info": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				"// @TermsOfServiceURL http://docs.io",
				"// @ContactEmail joe@bloggs.com",
				"// @LicenseURL http://license.mit.org",
			},
			want: types.InfoObject{
				Title:          "Test Run",
				Description:    "This is a test",
				TermsOfService: "http://docs.io",
				Contact: &types.ContactObject{
					Name:  "",
					URL:   "",
					Email: "joe@bloggs.com",
				},
				License: &types.LicenseObject{
					Name: "",
					URL:  "http://license.mit.org",
				},
				Version: "1.0.0",
			},
			expectErr: nil,
		},
		"all populated info properties": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				"// @TermsOfServiceURL http://docs.io",
				"// @ContactName Joe Bloggs",
				"// @ContactEmail joe@bloggs.com",
				"// @ContactURL http://test.com",
				"// @LicenseName MIT",
				"// @LicenseURL http://license.mit.org",
			},
			want: types.InfoObject{
				Title:          "Test Run",
				Description:    "This is a test",
				TermsOfService: "http://docs.io",
				Contact: &types.ContactObject{
					Name:  "Joe Bloggs",
					URL:   "http://test.com",
					Email: "joe@bloggs.com",
				},
				License: &types.LicenseObject{
					Name: "MIT",
					URL:  "http://license.mit.org",
				},
				Version: "1.0.0",
			},
			expectErr: nil,
		},
		"missing info.title": {
			comments: []string{
				"// @Version 1.0.0",
				"// @Description This is a test",
			},
			want: types.InfoObject{
				Title:       "",
				Description: "This is a test",
				Version:     "1.0.0",
			},
			expectErr: errors.New("info.title cannot not be empty"),
		},
		"missing version": {
			comments: []string{
				"// @Title Test App",
				"// @Description This is a test",
			},
			want: types.InfoObject{
				Title:       "Test App",
				Description: "This is a test",
				Version:     "",
			},
			expectErr: errors.New("info.version cannot not be empty"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p, err := partialBootstrap()
			if err != nil {
				t.Errorf("%v", err)
				return
			}
			fileComments := commentSliceToCommentGroup(tc.comments)

			if err := p.parseInfo(fileComments); err != nil {
				assert.Equal(t, tc.expectErr, err)
			}

			assert.Equal(t, tc.want, p.OpenAPI.Info)
		})
	}
}

func TestParseInfoServers(t *testing.T) {
	emptyServerVariableMap := make(map[string]types.ServerVariableObject)
	serverVariableMap := make(map[string]types.ServerVariableObject, 1)
	serverVariableMap["username"] = types.ServerVariableObject{
		Enum:        nil,
		Default:     "empty",
		Description: "Dev site username",
	}

	tests := map[string]struct {
		comments  []string
		want      []types.ServerObject
		expectErr error
	}{
		"single server": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				`// @Server http://dev.site.com Development Site`,
			},
			want: []types.ServerObject{
				{
					URL:         "http://dev.site.com",
					Description: "Development Site",
					Variables:   nil,
				},
			},
			expectErr: nil,
		},
		"single server with missing url": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				`// @Server test`,
			},
			want:      nil,
			expectErr: errors.New(`server: "test" is not a valid URL`),
		},
		"multiple servers": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				`// @Server http://dev.site.com Development Site`,
				`// @Server https://staging.site.com Staging Site`,
				`// @Server https://www.site.com Production Site`,
			},
			want: []types.ServerObject{
				{
					URL:         "http://dev.site.com",
					Description: "Development Site",
					Variables:   nil,
				},
				{
					URL:         "https://staging.site.com",
					Description: "Staging Site",
					Variables:   nil,
				},
				{
					URL:         "https://www.site.com",
					Description: "Production Site",
					Variables:   nil,
				},
			},
			expectErr: nil,
		},
		"multiple servers with one server variable": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				`// @Server http://dev.{username}.site.com Development Site`,
				`// @Server https://staging.site.com Staging Site`,
				`// @Server https://www.site.com Production Site`,
				`// @ServerVariable username "empty" "Dev site username"`,
			},
			want: []types.ServerObject{
				{
					URL:         "http://dev.{username}.site.com",
					Description: "Development Site",
					Variables:   serverVariableMap,
				},
				{
					URL:         "https://staging.site.com",
					Description: "Staging Site",
					Variables:   emptyServerVariableMap,
				},
				{
					URL:         "https://www.site.com",
					Description: "Production Site",
					Variables:   emptyServerVariableMap,
				},
			},
			expectErr: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p, err := partialBootstrap()
			if err != nil {
				t.Errorf("%v", err)
			}

			fileComments := commentSliceToCommentGroup(tc.comments)

			if err := p.parseInfo(fileComments); err != nil {
				assert.Equal(t, tc.expectErr, err)
			}

			sort.Slice(p.OpenAPI.Servers, func(i, j int) bool {
				return p.OpenAPI.Servers[i].URL < p.OpenAPI.Servers[j].URL
			})

			assert.Equal(t, tc.want, p.OpenAPI.Servers)
		})
	}
}

func TestParseInfoSecurity(t *testing.T) {
	tests := map[string]struct {
		comments           []string
		wantSecurity       []map[string][]string
		wantSecurityScheme map[string]*types.SecuritySchemeObject
	}{
		"combination of apiKey and http bearer": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				"// @SecurityScheme AuthorizationToken apiKey header X-Auth-Token Input your auth token",
				"// @SecurityScheme AuthorizationHeader http bearer Input your auth token",
			},
			wantSecurity: make([]map[string][]string, 0),
			wantSecurityScheme: map[string]*types.SecuritySchemeObject{
				"AuthorizationToken": {
					Type:             "apiKey",
					Description:      "Input your auth token",
					Scheme:           "",
					In:               "header",
					Name:             "X-Auth-Token",
					OpenIDConnectURL: "",
					OAuthFlows:       nil,
				},
				"AuthorizationHeader": {
					Type:             "http",
					Description:      "Input your auth token",
					Scheme:           "bearer",
					In:               "",
					Name:             "",
					OpenIDConnectURL: "",
					OAuthFlows:       nil,
				},
			},
		},
		"http basic auth": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				"// @SecurityScheme BasicAuth http basic token Basic Auth",
			},
			wantSecurityScheme: map[string]*types.SecuritySchemeObject{
				"BasicAuth": {
					Type:             "http",
					Description:      "Basic Auth",
					Scheme:           "basic",
					In:               "",
					Name:             "token",
					OpenIDConnectURL: "",
					OAuthFlows:       nil,
				},
			},
			wantSecurity: make([]map[string][]string, 0),
		},
		"openId connect": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				"// @SecurityScheme OpenID openIdConnect /connect OpenId connect, relative to basePath",
			},
			wantSecurityScheme: map[string]*types.SecuritySchemeObject{
				"OpenID": {
					Type:             "openIdConnect",
					Description:      "OpenId connect, relative to basePath",
					Scheme:           "",
					In:               "",
					Name:             "",
					OpenIDConnectURL: "/connect",
					OAuthFlows:       nil,
				},
			},
			wantSecurity: make([]map[string][]string, 0),
		},
		"oauth2 auth code": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				"// @SecurityScheme OAuth oauth2AuthCode /oauth/auth /oauth/token",
				"// @Security OAuth read write",
				"// @SecurityScope OAuth read Read only",
			},
			wantSecurityScheme: map[string]*types.SecuritySchemeObject{
				"OAuth": {
					Type:             "oauth2",
					Description:      "",
					OpenIDConnectURL: "",
					OAuthFlows: &types.SecuritySchemeOauthObject{
						AuthorizationCode: &types.SecuritySchemeOauthFlowObject{
							AuthorizationURL: "/oauth/auth",
							TokenURL:         "/oauth/token",
							Scopes: map[string]string{
								"read": "Read only",
							},
						},
					},
				},
			},
			wantSecurity: []map[string][]string{
				{
					"OAuth": []string{
						"read",
						"write",
					},
				},
			},
		},
		"oauth2 implicit": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				"// @SecurityScheme OAuth oauth2Implicit /oauth/auth",
				"// @Security OAuth read write",
				"// @SecurityScope OAuth read Read only",
			},
			wantSecurityScheme: map[string]*types.SecuritySchemeObject{
				"OAuth": {
					Type:             "oauth2",
					Description:      "",
					OpenIDConnectURL: "",
					OAuthFlows: &types.SecuritySchemeOauthObject{
						Implicit: &types.SecuritySchemeOauthFlowObject{
							AuthorizationURL: "/oauth/auth",
							Scopes: map[string]string{
								"read": "Read only",
							},
						},
					},
				},
			},
			wantSecurity: []map[string][]string{
				{
					"OAuth": []string{
						"read",
						"write",
					},
				},
			},
		},
		"oauth2 resource owner credentials": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				"// @SecurityScheme OAuth oauth2ResourceOwnerCredentials /oauth/token",
				"// @Security OAuth read write",
				"// @SecurityScope OAuth read Read only",
			},
			wantSecurityScheme: map[string]*types.SecuritySchemeObject{
				"OAuth": {
					Type:             "oauth2",
					Description:      "",
					OpenIDConnectURL: "",
					OAuthFlows: &types.SecuritySchemeOauthObject{
						ResourceOwnerPassword: &types.SecuritySchemeOauthFlowObject{
							TokenURL: "/oauth/token",
							Scopes: map[string]string{
								"read": "Read only",
							},
						},
					},
				},
			},
			wantSecurity: []map[string][]string{
				{
					"OAuth": []string{
						"read",
						"write",
					},
				},
			},
		},
		"oauth2 client credentials": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				"// @SecurityScheme OAuth oauth2ClientCredentials /oauth/token",
				"// @Security OAuth read write",
				"// @SecurityScope OAuth read Read only",
			},
			wantSecurityScheme: map[string]*types.SecuritySchemeObject{
				"OAuth": {
					Type:             "oauth2",
					Description:      "",
					OpenIDConnectURL: "",
					OAuthFlows: &types.SecuritySchemeOauthObject{
						ClientCredentials: &types.SecuritySchemeOauthFlowObject{
							TokenURL: "/oauth/token",
							Scopes: map[string]string{
								"read": "Read only",
							},
						},
					},
				},
			},
			wantSecurity: []map[string][]string{
				{
					"OAuth": []string{
						"read",
						"write",
					},
				},
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p, err := partialBootstrap()
			if err != nil {
				t.Errorf("%v", err)
			}

			fileComments := commentSliceToCommentGroup(tc.comments)
			if err := p.parseInfo(fileComments); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, tc.wantSecurity, p.OpenAPI.Security)
			assert.Equal(t, tc.wantSecurityScheme, p.OpenAPI.Components.SecuritySchemes)
		})
	}
}

func TestParseInfoExternalDoc(t *testing.T) {
	tests := map[string]struct {
		comments  []string
		want      types.OpenAPIObject
		expectErr error
	}{
		"populate external doc": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				`// @ExternalDoc https://docs.io "Documentation"`,
			},
			want: types.OpenAPIObject{
				OpenAPI: "3.0.0",
				Info: types.InfoObject{
					Title:       "Test Run",
					Description: "This is a test",
					Version:     "1.0.0",
				},
				Servers: nil,
				Paths:   types.PathsObject{},
				Components: types.ComponentsObject{
					Schemas:         map[string]*types.SchemaObject{},
					SecuritySchemes: map[string]*types.SecuritySchemeObject{},
				},
				Security: []map[string][]string{},
				Tags:     nil,
				ExternalDocs: &types.ExternalDocumentationObject{
					Description: "Documentation",
					URL:         "https://docs.io",
				},
			},
			expectErr: nil,
		},
		"missing description": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				`// @ExternalDoc https://docs.io `,
			},
			want: types.OpenAPIObject{
				OpenAPI: "3.0.0",
				Info: types.InfoObject{
					Title:       "Test Run",
					Description: "This is a test",
					Version:     "1.0.0",
				},
				ExternalDocs: nil,
				Paths:        types.PathsObject{},
				Components: types.ComponentsObject{
					Schemas:         map[string]*types.SchemaObject{},
					SecuritySchemes: map[string]*types.SecuritySchemeObject{},
				},
				Security: []map[string][]string{},
			},
			expectErr: errors.New(`parseExternalDocComment can not parse externaldoc comment "https://docs.io"`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p, err := partialBootstrap()
			if err != nil {
				t.Errorf("%v", err)
			}

			fileComments := commentSliceToCommentGroup(tc.comments)

			if err := p.parseInfo(fileComments); err != nil {
				assert.Equal(t, tc.expectErr, err)
			}

			assert.Equal(t, tc.want, p.OpenAPI)
		})
	}
}

func TestParseInfoTags(t *testing.T) {
	tests := map[string]struct {
		comments  []string
		want      types.OpenAPIObject
		expectErr error
	}{
		"add tag": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				`// @Tag users "Users"`,
			},
			want: types.OpenAPIObject{
				OpenAPI: "3.0.0",
				Info: types.InfoObject{
					Title:       "Test Run",
					Description: "This is a test",
					Version:     "1.0.0",
				},
				Tags: []types.TagObject{
					{
						Name:         "users",
						Description:  "Users",
						ExternalDocs: nil,
					},
				},
				Paths: types.PathsObject{},
				Components: types.ComponentsObject{
					Schemas:         map[string]*types.SchemaObject{},
					SecuritySchemes: map[string]*types.SecuritySchemeObject{},
				},
				Security: []map[string][]string{},
			},
			expectErr: nil,
		},
		"add tags": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				`// @Tag users "Users"`,
				`// @Tag admins "Admins"`,
			},
			want: types.OpenAPIObject{
				OpenAPI: "3.0.0",
				Info: types.InfoObject{
					Title:       "Test Run",
					Description: "This is a test",
					Version:     "1.0.0",
				},
				Tags: []types.TagObject{
					{
						Name:         "users",
						Description:  "Users",
						ExternalDocs: nil,
					},
					{
						Name:         "admins",
						Description:  "Admins",
						ExternalDocs: nil,
					},
				},
				Paths: types.PathsObject{},
				Components: types.ComponentsObject{
					Schemas:         map[string]*types.SchemaObject{},
					SecuritySchemes: map[string]*types.SecuritySchemeObject{},
				},
				Security: []map[string][]string{},
			},
			expectErr: nil,
		},
		"add tag with external docs": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				`// @Tag users "Users" https://docs.io "User Documentation"`,
				`// @Tag admins "Admins" https://docs.io "Admin Documentation"`,
			},
			want: types.OpenAPIObject{
				OpenAPI: "3.0.0",
				Info: types.InfoObject{
					Title:       "Test Run",
					Description: "This is a test",
					Version:     "1.0.0",
				},
				Tags: []types.TagObject{
					{
						Name:        "users",
						Description: "Users",
						ExternalDocs: &types.ExternalDocumentationObject{
							Description: "User Documentation",
							URL:         "https://docs.io",
						},
					},
					{
						Name:        "admins",
						Description: "Admins",
						ExternalDocs: &types.ExternalDocumentationObject{
							Description: "Admin Documentation",
							URL:         "https://docs.io",
						},
					},
				},
				Paths: types.PathsObject{},
				Components: types.ComponentsObject{
					Schemas:         map[string]*types.SchemaObject{},
					SecuritySchemes: map[string]*types.SecuritySchemeObject{},
				},
				Security: []map[string][]string{},
			},
			expectErr: nil,
		},
		"invalid tag": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				`// @Tag users `,
			},
			want: types.OpenAPIObject{
				OpenAPI: "3.0.0",
				Info: types.InfoObject{
					Title:       "Test Run",
					Description: "This is a test",
					Version:     "1.0.0",
				},
				Tags:  nil,
				Paths: types.PathsObject{},
				Components: types.ComponentsObject{
					Schemas:         map[string]*types.SchemaObject{},
					SecuritySchemes: map[string]*types.SecuritySchemeObject{},
				},
				Security: []map[string][]string{},
			},
			expectErr: errors.New("parseTagComment can not parse tag comment users"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p, err := partialBootstrap()
			if err != nil {
				t.Errorf("%v", err)
			}

			fileComments := commentSliceToCommentGroup(tc.comments)

			if err := p.parseInfo(fileComments); err != nil {
				assert.Equal(t, tc.expectErr, err)
			}

			assert.Equal(t, tc.want, p.OpenAPI)
		})
	}
}

func TestParseOperation(t *testing.T) {
	dir, _ := os.Getwd()
	tests := map[string]struct {
		pkgPath       string
		pkgName       string
		comments      []string
		wantPaths     types.PathsObject
		wantResponses types.ResponsesObject
		expectErr     error
	}{
		"hidden operation": {
			pkgPath: dir,
			pkgName: "main",
			comments: []string{
				"// @Title Super secret endpoint",
				"// @Description Ssshhh",
				"// @Hidden",
			},
			wantPaths:     types.PathsObject{},
			wantResponses: types.ResponsesObject{},
			expectErr:     nil,
		},
		"get operation without params": {
			pkgPath: dir,
			pkgName: "main",
			comments: []string{
				"// @Title Get all the things",
				"// @Description Get all the items",
				"// @Route / [get]",
				`// @Success 200 "Success"`,
				`// @Failure 400 "Failed"`,
				`// @Resource users`,
				`// @Resource`,
				`// @ID getAll`,
				`// @ExternalDoc https://docs.io "Get documentation"`,
			},
			wantPaths: types.PathsObject{
				"/": &types.PathItemObject{
					Get: &types.OperationObject{
						Responses: map[string]*types.ResponseObject{
							"200": {
								Description: "Success",
								Content:     make(map[string]*types.MediaTypeObject),
							},
							"400": {
								Description: "Failed",
								Content:     make(map[string]*types.MediaTypeObject),
							},
						},
						Summary:     "Get all the things",
						Description: "Get all the items",
						OperationID: "getAll",
						ExternalDocs: &types.ExternalDocumentationObject{
							Description: "Get documentation",
							URL:         "https://docs.io",
						},
						Tags: []string{"users", "others"},
					},
				},
			},
			expectErr: nil,
		},
		"get operation with params": {
			pkgPath: dir,
			pkgName: "main",
			comments: []string{
				"// @Title Get all the things",
				"// @Description Get all the items",
				"// @Route /{locale} [get]",
				`// @Param locale path string true "Locale code"`,
				`// @Success 200 "Success"`,
				`// @Failure 400 "Failed"`,
				`// @Resource users`,
				`// @ID getAll`,
			},
			wantPaths: types.PathsObject{
				"/{locale}": &types.PathItemObject{
					Get: &types.OperationObject{
						Responses: map[string]*types.ResponseObject{
							"200": {
								Description: "Success",
								Content:     make(map[string]*types.MediaTypeObject),
							},
							"400": {
								Description: "Failed",
								Content:     make(map[string]*types.MediaTypeObject),
							},
						},
						Summary:      "Get all the things",
						Description:  "Get all the items",
						OperationID:  "getAll",
						ExternalDocs: nil,
						Tags:         []string{"users"},
						Parameters: []types.ParameterObject{
							{
								Name:        "locale",
								In:          "path",
								Description: "Locale code",
								Required:    true,
								Schema: &types.SchemaObject{
									Type:   "string",
									Format: "string",
								},
							},
						},
					},
				},
			},
			expectErr: nil,
		},
		"post operation with body": {
			pkgPath: dir,
			pkgName: "main",
			comments: []string{
				"// @Title Create a user",
				"// @Description Create a user",
				"// @Route /{locale} [post]",
				`// @Param locale path string true "Locale code"`,
				`// @Param username body string true "Username"`,
				`// @Success 201 "Created"`,
				`// @Failure 400 "Failed"`,
				`// @Resource users`,
				`// @ID createUser`,
			},
			wantPaths: types.PathsObject{
				"/{locale}": &types.PathItemObject{
					Post: &types.OperationObject{
						Responses: map[string]*types.ResponseObject{
							"201": {
								Description: "Created",
								Content:     make(map[string]*types.MediaTypeObject),
							},
							"400": {
								Description: "Failed",
								Content:     make(map[string]*types.MediaTypeObject),
							},
						},
						Summary:      "Create a user",
						Description:  "Create a user",
						OperationID:  "createUser",
						ExternalDocs: nil,
						Tags:         []string{"users"},
						Parameters: []types.ParameterObject{
							{
								Name:        "locale",
								In:          "path",
								Description: "Locale code",
								Required:    true,
								Schema: &types.SchemaObject{
									Type:   "string",
									Format: "string",
								},
							},
						},
						RequestBody: &types.RequestBodyObject{
							Content: map[string]*types.MediaTypeObject{
								types.ContentTypeJSON: {
									Schema: types.SchemaObject{
										Type: "string",
									},
								},
							},
							Description: "",
							Required:    true,
							Ref:         "",
						},
					},
				},
			},
			expectErr: nil,
		},
		"patch operation": {
			pkgPath: dir,
			pkgName: "main",
			comments: []string{
				"// @Title Update a user",
				"// @Description Update a user",
				"// @Route /{locale}/{id} [patch]",
				`// @Param locale path string true "Locale code"`,
				`// @Param id path int true "User ID"`,
				`// @Param username body string true "Username"`,
				`// @Success 200 "Success"`,
				`// @Failure 400 "Failed"`,
				`// @Resource users`,
				`// @ID updateUser`,
			},
			wantPaths: types.PathsObject{
				"/{locale}/{id}": &types.PathItemObject{
					Patch: &types.OperationObject{
						Responses: map[string]*types.ResponseObject{
							"200": {
								Description: "Success",
								Content:     make(map[string]*types.MediaTypeObject),
							},
							"400": {
								Description: "Failed",
								Content:     make(map[string]*types.MediaTypeObject),
							},
						},
						Summary:      "Update a user",
						Description:  "Update a user",
						OperationID:  "updateUser",
						ExternalDocs: nil,
						Tags:         []string{"users"},
						Parameters: []types.ParameterObject{
							{
								Name:        "locale",
								In:          "path",
								Description: "Locale code",
								Required:    true,
								Schema: &types.SchemaObject{
									Type:   "string",
									Format: "string",
								},
							},
							{
								Name:        "id",
								In:          "path",
								Description: "User ID",
								Required:    true,
								Schema: &types.SchemaObject{
									Type:   "integer",
									Format: "int64",
								},
							},
						},
						RequestBody: &types.RequestBodyObject{
							Content: map[string]*types.MediaTypeObject{
								types.ContentTypeJSON: {
									Schema: types.SchemaObject{
										Type: "string",
									},
								},
							},
							Description: "",
							Required:    true,
							Ref:         "",
						},
					},
				},
			},
			expectErr: nil,
		},
		"put operation": {
			pkgPath: dir,
			pkgName: "main",
			comments: []string{
				"// @Title Replace a user",
				"// @Description Replace a user",
				"// @Route /{locale}/{id} [put]",
				`// @Param locale path string true "Locale code"`,
				`// @Param id path int true "User ID"`,
				`// @Param username body string true "Username"`,
				`// @Success 200 "Success"`,
				`// @Failure 400 "Failed"`,
				`// @Resource users`,
				`// @ID replaceUser`,
			},
			wantPaths: types.PathsObject{
				"/{locale}/{id}": &types.PathItemObject{
					Put: &types.OperationObject{
						Responses: map[string]*types.ResponseObject{
							"200": {
								Description: "Success",
								Content:     make(map[string]*types.MediaTypeObject),
							},
							"400": {
								Description: "Failed",
								Content:     make(map[string]*types.MediaTypeObject),
							},
						},
						Summary:      "Replace a user",
						Description:  "Replace a user",
						OperationID:  "replaceUser",
						ExternalDocs: nil,
						Tags:         []string{"users"},
						Parameters: []types.ParameterObject{
							{
								Name:        "locale",
								In:          "path",
								Description: "Locale code",
								Required:    true,
								Schema: &types.SchemaObject{
									Type:   "string",
									Format: "string",
								},
							},
							{
								Name:        "id",
								In:          "path",
								Description: "User ID",
								Required:    true,
								Schema: &types.SchemaObject{
									Type:   "integer",
									Format: "int64",
								},
							},
						},
						RequestBody: &types.RequestBodyObject{
							Content: map[string]*types.MediaTypeObject{
								types.ContentTypeJSON: {
									Schema: types.SchemaObject{
										Type: "string",
									},
								},
							},
							Description: "",
							Required:    true,
							Ref:         "",
						},
					},
				},
			},
			expectErr: nil,
		},
		"delete operation": {
			pkgPath: dir,
			pkgName: "main",
			comments: []string{
				"// @Title Delete a user",
				"// @Description Delete a user",
				"// @Route /{locale}/{id} [delete]",
				`// @Param locale path string true "Locale code"`,
				`// @Param id path int true "User ID"`,
				`// @Success 200 "Success"`,
				`// @Failure 400 "Failed"`,
				`// @Resource users`,
				`// @ID deleteUser`,
			},
			wantPaths: types.PathsObject{
				"/{locale}/{id}": &types.PathItemObject{
					Delete: &types.OperationObject{
						Responses: map[string]*types.ResponseObject{
							"200": {
								Description: "Success",
								Content:     make(map[string]*types.MediaTypeObject),
							},
							"400": {
								Description: "Failed",
								Content:     make(map[string]*types.MediaTypeObject),
							},
						},
						Summary:      "Delete a user",
						Description:  "Delete a user",
						OperationID:  "deleteUser",
						ExternalDocs: nil,
						Tags:         []string{"users"},
						Parameters: []types.ParameterObject{
							{
								Name:        "locale",
								In:          "path",
								Description: "Locale code",
								Required:    true,
								Schema: &types.SchemaObject{
									Type:   "string",
									Format: "string",
								},
							},
							{
								Name:        "id",
								In:          "path",
								Description: "User ID",
								Required:    true,
								Schema: &types.SchemaObject{
									Type:   "integer",
									Format: "int64",
								},
							},
						},
					},
				},
			},
			expectErr: nil,
		},
		"options operation": {
			pkgPath: dir,
			pkgName: "main",
			comments: []string{
				"// @Title User pre-flight",
				"// @Description User pre-flight",
				"// @Route /{locale}/{id} [options]",
				`// @Param locale path string true "Locale code"`,
				`// @Param id path int true "User ID"`,
				`// @Success 200 "Success"`,
				`// @Failure 400 "Failed"`,
				`// @Resource users`,
			},
			wantPaths: types.PathsObject{
				"/{locale}/{id}": &types.PathItemObject{
					Options: &types.OperationObject{
						Responses: map[string]*types.ResponseObject{
							"200": {
								Description: "Success",
								Content:     make(map[string]*types.MediaTypeObject),
							},
							"400": {
								Description: "Failed",
								Content:     make(map[string]*types.MediaTypeObject),
							},
						},
						Summary:      "User pre-flight",
						Description:  "User pre-flight",
						ExternalDocs: nil,
						Tags:         []string{"users"},
						Parameters: []types.ParameterObject{
							{
								Name:        "locale",
								In:          "path",
								Description: "Locale code",
								Required:    true,
								Schema: &types.SchemaObject{
									Type:   "string",
									Format: "string",
								},
							},
							{
								Name:        "id",
								In:          "path",
								Description: "User ID",
								Required:    true,
								Schema: &types.SchemaObject{
									Type:   "integer",
									Format: "int64",
								},
							},
						},
					},
				},
			},
			expectErr: nil,
		},
		"head operation": {
			pkgPath: dir,
			pkgName: "main",
			comments: []string{
				"// @Title User Head Lookup",
				"// @Description User Head Lookup",
				"// @Route /{locale}/{id} [head]",
				`// @Param locale path string true "Locale code"`,
				`// @Param id path int true "User ID"`,
				`// @Resource users`,
			},
			wantPaths: types.PathsObject{
				"/{locale}/{id}": &types.PathItemObject{
					Head: &types.OperationObject{
						Responses:    make(map[string]*types.ResponseObject),
						Summary:      "User Head Lookup",
						Description:  "User Head Lookup",
						ExternalDocs: nil,
						Tags:         []string{"users"},
						Parameters: []types.ParameterObject{
							{
								Name:        "locale",
								In:          "path",
								Description: "Locale code",
								Required:    true,
								Schema: &types.SchemaObject{
									Type:   "string",
									Format: "string",
								},
							},
							{
								Name:        "id",
								In:          "path",
								Description: "User ID",
								Required:    true,
								Schema: &types.SchemaObject{
									Type:   "integer",
									Format: "int64",
								},
							},
						},
					},
				},
			},
			expectErr: nil,
		},
		"trace operation without params": {
			pkgPath: dir,
			pkgName: "main",
			comments: []string{
				"// @Title User Trace (should be disabled)",
				"// @Description User Trace (should be disabled)",
				"// @Route /{locale}/{id} [trace]",
				`// @Param locale path string true "Locale code"`,
				`// @Param id path int true "User ID"`,
				`// @Resource users`,
			},
			wantPaths: types.PathsObject{
				"/{locale}/{id}": &types.PathItemObject{
					Trace: &types.OperationObject{
						Responses:    make(map[string]*types.ResponseObject),
						Summary:      "User Trace (should be disabled)",
						Description:  "User Trace (should be disabled)",
						ExternalDocs: nil,
						Tags:         []string{"users"},
						Parameters: []types.ParameterObject{
							{
								Name:        "locale",
								In:          "path",
								Description: "Locale code",
								Required:    true,
								Schema: &types.SchemaObject{
									Type:   "string",
									Format: "string",
								},
							},
							{
								Name:        "id",
								In:          "path",
								Description: "User ID",
								Required:    true,
								Schema: &types.SchemaObject{
									Type:   "integer",
									Format: "int64",
								},
							},
						},
					},
				},
			},
			expectErr: nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p, err := partialBootstrap()
			if err != nil {
				t.Fatalf("%v", err)
			}

			fileComments := commentSliceToCommentGroup(tc.comments)

			if err = p.parseOperation(tc.pkgPath, tc.pkgName, fileComments[0].List); err != nil {
				assert.Equal(t, tc.expectErr, err)
				return
			}

			assert.Equal(t, tc.wantPaths, p.OpenAPI.Paths)
		})
	}
}

func TestIntegration(t *testing.T) {
	// @see https://github.com/OAI/OpenAPI-Specification/blob/master/examples/v3.0/petstore.yaml
	tests := map[string]struct {
		mode   string
		format string
	}{
		"integration test - yaml": {
			ModeTest,
			FormatYAML,
		},
		"integration test - json": {
			ModeTest,
			FormatJSON,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			modulePath := util.ModulePath("./")
			path, _ := modulePath.Get()
			p, _ := newParser(
				"./",
				"test/integration/docs.go",
				"test/integration/pkg/integration_handler",
				fmt.Sprintf("%s/test/unit", path),
				false,
			)
			test, err := p.CreateOAS("", tc.mode, tc.format)
			if err != nil {
				assert.NoError(t, err)
			}

			assert.NotEmpty(t, test)

			var oapi *types.OpenAPIObject
			switch tc.format {
			case FormatYAML:
				err = yaml.Unmarshal([]byte(*test), &oapi)
			case FormatJSON:
				err = json.Unmarshal([]byte(*test), &oapi)
			}
			assert.NoError(t, err)

			assert.Equal(t, "3.0.0", oapi.OpenAPI)
			assert.Equal(t, "Swagger Pet Store", oapi.Info.Title)
			assert.Equal(t, "MIT", oapi.Info.License.Name)
			assert.Equal(t, "http://petstore.swagger.io/v1", oapi.Servers[0].URL)
			assert.Equal(t, "List all pets", oapi.Paths["/pets"].Get.Summary)
			assert.Equal(t, "listPets", oapi.Paths["/pets"].Get.OperationID)
			assert.Equal(t, "object", oapi.Components.Schemas["Pet"].Type)
			id, ok := oapi.Components.Schemas["Pet"].Properties.Get("id")
			strictID := id.(orderedmap.OrderedMap)
			propertyType, _ := strictID.Get("type")
			assert.True(t, ok)
			assert.Equal(t, "integer", propertyType)
		})
	}
}

func commentSliceToCommentGroup(commentSlice []string) []*ast.CommentGroup {
	var comments []*ast.Comment
	for _, comment := range commentSlice {
		comments = append(comments, &ast.Comment{
			Slash: 0,
			Text:  comment,
		})
	}

	commentGroup := &ast.CommentGroup{
		List: comments,
	}

	var fileComments []*ast.CommentGroup
	fileComments = append(fileComments, commentGroup)

	return fileComments
}

func partialBootstrap() (*parser, error) {
	modulePath := util.ModulePath("./")
	path, _ := modulePath.Get()
	p, err := newParser(
		"./",
		"./main.go",
		"",
		fmt.Sprintf("%s/test/integration,%s/test/integration/pkg/integration_handler", path, path),
		false,
	)
	if err != nil {
		return nil, err
	}
	p.parseModule()
	if err := p.parseGoMod(); err != nil {
		return nil, err
	}
	if err := p.parseAPIs(); err != nil {
		return nil, err
	}

	return p, nil
}
