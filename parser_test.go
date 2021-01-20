package main

import (
	"errors"
	"go/ast"
	"os"
	"sort"
	"testing"

	"github.com/iancoleman/orderedmap"

	"github.com/stretchr/testify/assert"
)

func TestParseParamComment(t *testing.T) {

	fileProp := orderedmap.New()
	fileProp.Set("image", &SchemaObject{
		ID:                 "",
		PkgName:            "",
		FieldName:          "",
		DisabledFieldNames: nil,
		Type:               "string",
		Format:             "binary",
		Required:           nil,
		Properties:         nil,
		Description:        "Image upload",
		Items:              nil,
		Example:            nil,
		Deprecated:         false,
		ExternalDocs:       nil,
		Ref:                "",
	})

	filesProp := orderedmap.New()
	filesProp.Set("image", &SchemaObject{
		Type:        "array",
		Description: "Image upload",
		Items:       &SchemaObject{Type: "string", Format: "binary"},
	})

	stringProp := orderedmap.New()
	stringProp.Set("content", &SchemaObject{
		Type:        "string",
		Format:      "string",
		Description: "Content field",
	})

	mapStringProp := orderedmap.New()
	mapStringProp.Set("address", &SchemaObject{
		Type:        "string",
		Format:      "",
		Description: "",
	})

	extDocMap := orderedmap.New()
	extDocMap.Set("description", &SchemaObject{
		FieldName: "Description",
		Type:      "string",
	})
	extDocMap.Set("url", &SchemaObject{
		FieldName: "URL",
		Type:      "string",
	})

	extDocProp := orderedmap.New()
	extDocProp.Set("externaldocs", &SchemaObject{
		ID:                 "ExternalDocumentationObject",
		PkgName:            "github.com/deanstalker/goas",
		Type:               "object",
		Properties:         extDocMap,
		Ref:                "#/components/schemas/ExternalDocumentationObject",
		DisabledFieldNames: map[string]struct{}{},
	})

	tests := map[string]struct {
		pkgPath   string
		pkgName   string
		comment   string
		want      *OperationObject
		expectErr error
	}{
		"string param in path": {
			pkgPath: "/",
			pkgName: "main",
			comment: `locale   path   string   true   "Locale code"`,
			want: &OperationObject{
				Parameters: []ParameterObject{
					{
						Name:        "locale",
						In:          "path",
						Description: "Locale code",
						Required:    true,
						Example:     nil,
						Schema: &SchemaObject{
							ID:                 "",
							PkgName:            "",
							FieldName:          "",
							DisabledFieldNames: nil,
							Type:               "string",
							Format:             "string",
							Required:           nil,
							Properties:         nil,
							Description:        "Locale code",
							Items:              nil,
							Example:            nil,
							Deprecated:         false,
							Ref:                "",
						},
						Ref: "",
					},
				},
			},
			expectErr: nil,
		},
		"string param in path without desc": {
			pkgPath: "/",
			pkgName: "main",
			comment: `locale   path   string   true`,
			want: &OperationObject{
				Parameters: []ParameterObject{
					{
						Name:        "locale",
						In:          "path",
						Description: "Locale code",
						Required:    true,
						Example:     nil,
						Schema: &SchemaObject{
							ID:                 "",
							PkgName:            "",
							FieldName:          "",
							DisabledFieldNames: nil,
							Type:               "string",
							Format:             "string",
							Required:           nil,
							Properties:         nil,
							Description:        "Locale code",
							Items:              nil,
							Example:            nil,
							Deprecated:         false,
							Ref:                "",
						},
						Ref: "",
					},
				},
			},
			expectErr: errors.New(`parseParamComment can not parse param comment "locale   path   string   true"`),
		},
		"string in body": {
			pkgPath: "/",
			pkgName: "main",
			comment: `firstname   body   string   true   "First Name"`,
			want: &OperationObject{
				RequestBody: &RequestBodyObject{
					Content: map[string]*MediaTypeObject{
						ContentTypeJson: &MediaTypeObject{
							Schema: SchemaObject{
								Type: "string",
							},
						},
					},
					Required: true,
				},
			},
			expectErr: nil,
		},
		"[]string in body": {
			pkgPath: "/",
			pkgName: "main",
			comment: `address   body   []string   true   "Address"`,
			want: &OperationObject{
				RequestBody: &RequestBodyObject{
					Content: map[string]*MediaTypeObject{
						ContentTypeJson: &MediaTypeObject{
							Schema: SchemaObject{
								Type: "array",
								Items: &SchemaObject{
									Type: "string",
								},
							},
						},
					},
					Required: true,
				},
			},
			expectErr: nil,
		},
		"map[]string in body": {
			pkgPath: "/",
			pkgName: "main",
			comment: `address   body   map[]string   true   "Address"`,
			want: &OperationObject{
				RequestBody: &RequestBodyObject{
					Content: map[string]*MediaTypeObject{
						ContentTypeJson: &MediaTypeObject{
							Schema: SchemaObject{
								Type:       "object",
								Properties: mapStringProp,
							},
						},
					},
					Required: true,
				},
			},
			expectErr: nil,
		},
		"timestamp in path": {
			pkgPath: "/",
			pkgName: "main",
			comment: `time   path   time.Time   true   "Timestamp"`,
			want: &OperationObject{
				Parameters: []ParameterObject{
					{
						Name:        "time",
						In:          "path",
						Description: "Timestamp",
						Required:    true,
						Example:     nil,
						Schema: &SchemaObject{
							ID:                 "",
							PkgName:            "",
							FieldName:          "",
							DisabledFieldNames: nil,
							Type:               "string",
							Format:             "date-time",
							Required:           nil,
							Properties:         nil,
							Description:        "",
							Items:              nil,
							Example:            nil,
							Deprecated:         false,
							Ref:                "",
						},
						Ref: "",
					},
				},
			},
			expectErr: nil,
		},
		"file in body": {
			pkgPath: "/",
			pkgName: "main",
			comment: `image file ignored true "Image upload"`,
			want: &OperationObject{
				RequestBody: &RequestBodyObject{
					Content: map[string]*MediaTypeObject{
						ContentTypeForm: &MediaTypeObject{
							Schema: SchemaObject{
								ID:                 "",
								PkgName:            "",
								FieldName:          "",
								DisabledFieldNames: nil,
								Type:               "object",
								Format:             "",
								Required:           nil,
								Properties:         fileProp,
								Description:        "",
								Items:              nil,
								Example:            nil,
								Deprecated:         false,
								ExternalDocs:       nil,
								Ref:                "",
							},
						},
					},
					Description: "",
					Required:    true,
					Ref:         "",
				},
			},
			expectErr: nil,
		},
		"files in body": {
			pkgPath: "/",
			pkgName: "main",
			comment: `image files ignored true "Image upload"`,
			want: &OperationObject{
				RequestBody: &RequestBodyObject{
					Content: map[string]*MediaTypeObject{
						ContentTypeForm: &MediaTypeObject{
							Schema: SchemaObject{
								Type:       "object",
								Properties: filesProp,
							},
						},
					},
					Description: "",
					Required:    true,
					Ref:         "",
				},
			},
			expectErr: nil,
		},
		"form field with string in body": {
			pkgPath: "/",
			pkgName: "main",
			comment: `content form string false "Content field"`,
			want: &OperationObject{
				RequestBody: &RequestBodyObject{
					Content: map[string]*MediaTypeObject{
						ContentTypeForm: &MediaTypeObject{
							Schema: SchemaObject{
								Type:       "object",
								Properties: stringProp,
							},
						},
					},
					Description: "",
					Required:    false,
					Ref:         "",
				},
			},
			expectErr: nil,
		},
		"struct in body": {
			pkgPath: "github.com/deanstalker/goas",
			pkgName: "main",
			comment: `externaldocs body ExternalDocumentationObject false "External Documentation"`,
			want: &OperationObject{
				RequestBody: &RequestBodyObject{
					Content: map[string]*MediaTypeObject{
						ContentTypeJson: &MediaTypeObject{
							Schema: SchemaObject{
								Ref: "#/components/schemas/ExternalDocumentationObject",
							},
						},
					},
					Description: "",
					Required:    false,
					Ref:         "",
				},
			},
			expectErr: nil,
		},
		"struct path in body": {
			pkgName: "main.ExternalDocumentationObject",
			pkgPath: "",
			comment: `externaldocs body deanstalker.goas.ExternalDocumentationObject false "External Documentation"`,
			want: &OperationObject{
				RequestBody: &RequestBodyObject{
					Content: map[string]*MediaTypeObject{
						ContentTypeJson: &MediaTypeObject{
							Schema: SchemaObject{
								Ref: "#/components/schemas/ExternalDocumentationObject",
							},
						},
					},
					Description: "",
					Required:    false,
					Ref:         "",
				},
			},
			expectErr: nil,
		},
		"array of structs in body": {
			pkgPath: "github.com/deanstalker/goas",
			pkgName: "main",
			comment: `externaldocs body []ExternalDocumentationObject false "External Documentation"`,
			want: &OperationObject{
				RequestBody: &RequestBodyObject{
					Content: map[string]*MediaTypeObject{
						ContentTypeJson: &MediaTypeObject{
							Schema: SchemaObject{
								Type: "array",
								Items: &SchemaObject{
									ID:                 "ExternalDocumentationObject",
									PkgName:            "github.com/deanstalker/goas",
									Type:               "object",
									Properties:         extDocMap,
									Ref:                "#/components/schemas/ExternalDocumentationObject",
									DisabledFieldNames: map[string]struct{}{},
								},
							},
						},
					},
				},
			},
			expectErr: nil,
		},
		"map of structs in body": {
			pkgPath: "github.com/deanstalker/goas",
			pkgName: "main",
			comment: `externaldocs body map[]ExternalDocumentationObject false "External Documentation"`,
			want: &OperationObject{
				RequestBody: &RequestBodyObject{
					Content: map[string]*MediaTypeObject{
						ContentTypeJson: &MediaTypeObject{
							Schema: SchemaObject{
								Type:       "object",
								Properties: extDocProp,
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
			p, err := partialBootstrap(t)
			if err != nil {
				t.Errorf("%v", err)
			}

			op := &OperationObject{}
			if err := p.parseParamComment(tc.pkgPath, tc.pkgName, op, tc.comment); err != nil {
				assert.Equal(t, tc.expectErr, err)
				return
			}

			assert.Equal(t, tc.want, op)
		})
	}
}

func TestParseServerVariableComments(t *testing.T) {
	type test struct {
		comment string
		server  ServerObject
		want    map[string]ServerVariableObject
	}

	var tests = []test{
		{
			comment: `username "empty" "Enter a username for dev testing"`,
			server: ServerObject{
				URL:         "https://api.{username}.dev.lan/",
				Description: "",
				Variables:   make(map[string]ServerVariableObject),
			},
			want: map[string]ServerVariableObject{
				"username": {
					Enum:        nil,
					Default:     "empty",
					Description: "Enter a username for dev testing",
				},
			},
		},
		{
			comment: `username "80" "Enter a server port" "80,443,8443,8080"`,
			server: ServerObject{
				URL:         "https://api.{username}.dev.lan/",
				Description: "",
				Variables:   make(map[string]ServerVariableObject),
			},
			want: map[string]ServerVariableObject{
				"username": {
					Enum:        []string{"80", "443", "8443", "8080"},
					Default:     "80",
					Description: "Enter a server port",
				},
			},
		},
	}

	p := &parser{}

	for _, tc := range tests {
		parsed, err := p.parseServerVariableComment(tc.comment, tc.server)
		if err != nil {
			t.Errorf("%v", err)
		}
		assert.Equal(t, tc.want, parsed)
	}
}

func TestParseTagComments(t *testing.T) {

	type test struct {
		comment string
		want    TagObject
	}

	tests := []test{
		{comment: `test-service "this is a test service"`, want: TagObject{Name: "test-service", Description: "this is a test service"}},
		{comment: `test-service "this is a test service" https://docs.io  "External Docs"`, want: TagObject{
			Name:        "test-service",
			Description: "this is a test service",
			ExternalDocs: &ExternalDocumentationObject{
				Description: "External Docs",
				URL:         "https://docs.io",
			},
		}},
	}

	p := &parser{}

	for _, tc := range tests {
		tag, err := p.parseTagComment(tc.comment)
		if err != nil {
			t.Errorf("%v", err)
		}

		assert.Equal(t, tc.want.Description, tag.Description)
		assert.Equal(t, tc.want.Name, tag.Name)
		assert.Equal(t, tc.want.ExternalDocs, tag.ExternalDocs)
	}
}

func TestParseInfo(t *testing.T) {
	tests := map[string]struct {
		comments  []string
		want      InfoObject
		expectErr error
	}{
		"minimum required info": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
			},
			want: InfoObject{
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
			want: InfoObject{
				Title:          "Test Run",
				Description:    "This is a test",
				TermsOfService: "http://docs.io",
				Contact: &ContactObject{
					Name:  "",
					URL:   "",
					Email: "joe@bloggs.com",
				},
				License: &LicenseObject{
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
			want: InfoObject{
				Title:          "Test Run",
				Description:    "This is a test",
				TermsOfService: "http://docs.io",
				Contact: &ContactObject{
					Name:  "Joe Bloggs",
					URL:   "http://test.com",
					Email: "joe@bloggs.com",
				},
				License: &LicenseObject{
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
			want: InfoObject{
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
			want: InfoObject{
				Title:       "Test App",
				Description: "This is a test",
				Version:     "",
			},
			expectErr: errors.New("info.version cannot not be empty"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p := &parser{}

			fileComments := commentSliceToCommentGroup(tc.comments)

			if err := p.parseInfo(fileComments); err != nil {
				assert.Equal(t, tc.expectErr, err)
			}

			assert.Equal(t, tc.want, p.OpenAPI.Info)
		})
	}
}

func TestParseInfoServers(t *testing.T) {
	emptyServerVariableMap := make(map[string]ServerVariableObject)
	serverVariableMap := make(map[string]ServerVariableObject, 1)
	serverVariableMap["username"] = ServerVariableObject{
		Enum:        nil,
		Default:     "empty",
		Description: "Dev site username",
	}

	tests := map[string]struct {
		comments  []string
		want      []ServerObject
		expectErr error
	}{
		"single server": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				`// @Server http://dev.site.com Development Site`,
			},
			want: []ServerObject{
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
			want: []ServerObject{
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
			want: []ServerObject{
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
			p := &parser{}

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
		wantSecurityScheme map[string]*SecuritySchemeObject
	}{
		"combination of apiKey and http bearer": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				"// @SecurityScheme AuthorizationToken apiKey header X-Jumbo-Auth-Token Input your auth token",
				"// @SecurityScheme AuthorizationHeader http bearer Input your auth token",
			},
			wantSecurity: nil,
			wantSecurityScheme: map[string]*SecuritySchemeObject{
				"AuthorizationToken": &SecuritySchemeObject{
					Type:             "apiKey",
					Description:      "Input your auth token",
					Scheme:           "",
					In:               "header",
					Name:             "X-Jumbo-Auth-Token", // TODO remove references to the commercial name behind the change
					OpenIdConnectUrl: "",
					OAuthFlows:       nil,
				},
				"AuthorizationHeader": &SecuritySchemeObject{
					Type:             "http",
					Description:      "Input your auth token",
					Scheme:           "bearer",
					In:               "",
					Name:             "",
					OpenIdConnectUrl: "",
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
			wantSecurityScheme: map[string]*SecuritySchemeObject{
				"BasicAuth": &SecuritySchemeObject{
					Type:             "http",
					Description:      "Basic Auth",
					Scheme:           "basic",
					In:               "",
					Name:             "token",
					OpenIdConnectUrl: "",
					OAuthFlows:       nil,
				},
			},
			wantSecurity: nil,
		},
		"openId connect": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				"// @SecurityScheme OpenID openIdConnect /connect OpenId connect, relative to basePath",
			},
			wantSecurityScheme: map[string]*SecuritySchemeObject{
				"OpenID": &SecuritySchemeObject{
					Type:             "openIdConnect",
					Description:      "OpenId connect, relative to basePath",
					Scheme:           "",
					In:               "",
					Name:             "",
					OpenIdConnectUrl: "/connect",
					OAuthFlows:       nil,
				},
			},
			wantSecurity: nil,
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
			wantSecurityScheme: map[string]*SecuritySchemeObject{
				"OAuth": &SecuritySchemeObject{
					Type:             "oauth2",
					Description:      "",
					OpenIdConnectUrl: "",
					OAuthFlows: &SecuritySchemeOauthObject{
						AuthorizationCode: &SecuritySchemeOauthFlowObject{
							AuthorizationUrl: "/oauth/auth",
							TokenUrl:         "/oauth/token",
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
			wantSecurityScheme: map[string]*SecuritySchemeObject{
				"OAuth": &SecuritySchemeObject{
					Type:             "oauth2",
					Description:      "",
					OpenIdConnectUrl: "",
					OAuthFlows: &SecuritySchemeOauthObject{
						Implicit: &SecuritySchemeOauthFlowObject{
							AuthorizationUrl: "/oauth/auth",
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
			wantSecurityScheme: map[string]*SecuritySchemeObject{
				"OAuth": &SecuritySchemeObject{
					Type:             "oauth2",
					Description:      "",
					OpenIdConnectUrl: "",
					OAuthFlows: &SecuritySchemeOauthObject{
						ResourceOwnerPassword: &SecuritySchemeOauthFlowObject{
							TokenUrl: "/oauth/token",
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
			wantSecurityScheme: map[string]*SecuritySchemeObject{
				"OAuth": &SecuritySchemeObject{
					Type:             "oauth2",
					Description:      "",
					OpenIdConnectUrl: "",
					OAuthFlows: &SecuritySchemeOauthObject{
						ClientCredentials: &SecuritySchemeOauthFlowObject{
							TokenUrl: "/oauth/token",
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
			p := &parser{}

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
		want      OpenAPIObject
		expectErr error
	}{
		"populate external doc": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				`// @ExternalDoc https://docs.io "Documentation"`,
			},
			want: OpenAPIObject{
				OpenAPI: "",
				Info: InfoObject{
					Title:       "Test Run",
					Description: "This is a test",
					Version:     "1.0.0",
				},
				ExternalDocs: ExternalDocumentationObject{
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
			want: OpenAPIObject{
				OpenAPI: "",
				Info: InfoObject{
					Title:       "Test Run",
					Description: "This is a test",
					Version:     "1.0.0",
				},
				ExternalDocs: ExternalDocumentationObject{
					Description: "",
					URL:         "",
				},
			},
			expectErr: errors.New(`parseExternalDocComment can not parse externaldoc comment "https://docs.io"`),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p := &parser{}

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
		want      OpenAPIObject
		expectErr error
	}{
		"add tag": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				`// @Tag users "Users"`,
			},
			want: OpenAPIObject{
				OpenAPI: "",
				Info: InfoObject{
					Title:       "Test Run",
					Description: "This is a test",
					Version:     "1.0.0",
				},
				Tags: []TagObject{
					{
						Name:         "users",
						Description:  "Users",
						ExternalDocs: nil,
					},
				},
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
			want: OpenAPIObject{
				OpenAPI: "",
				Info: InfoObject{
					Title:       "Test Run",
					Description: "This is a test",
					Version:     "1.0.0",
				},
				Tags: []TagObject{
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
			want: OpenAPIObject{
				OpenAPI: "",
				Info: InfoObject{
					Title:       "Test Run",
					Description: "This is a test",
					Version:     "1.0.0",
				},
				Tags: []TagObject{
					{
						Name:        "users",
						Description: "Users",
						ExternalDocs: &ExternalDocumentationObject{
							Description: "User Documentation",
							URL:         "https://docs.io",
						},
					},
					{
						Name:        "admins",
						Description: "Admins",
						ExternalDocs: &ExternalDocumentationObject{
							Description: "Admin Documentation",
							URL:         "https://docs.io",
						},
					},
				},
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
			want: OpenAPIObject{
				OpenAPI: "",
				Info: InfoObject{
					Title:       "Test Run",
					Description: "This is a test",
					Version:     "1.0.0",
				},
				Tags: nil,
			},
			expectErr: errors.New("parseTagComment can not parse tag comment users"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			p := &parser{}

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
		wantPaths     PathsObject
		wantResponses ResponsesObject
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
			wantPaths:     PathsObject{},
			wantResponses: ResponsesObject{},
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
			wantPaths: PathsObject{
				"/": &PathItemObject{
					Get: &OperationObject{
						Responses: map[string]*ResponseObject{
							"200": &ResponseObject{
								Description: "Success",
								Content:     make(map[string]*MediaTypeObject),
							},
							"400": &ResponseObject{
								Description: "Failed",
								Content:     make(map[string]*MediaTypeObject),
							},
						},
						Summary:     "Get all the things",
						Description: "Get all the items",
						OperationID: "getAll",
						ExternalDocs: ExternalDocumentationObject{
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
			wantPaths: PathsObject{
				"/{locale}": &PathItemObject{
					Get: &OperationObject{
						Responses: map[string]*ResponseObject{
							"200": &ResponseObject{
								Description: "Success",
								Content:     make(map[string]*MediaTypeObject),
							},
							"400": &ResponseObject{
								Description: "Failed",
								Content:     make(map[string]*MediaTypeObject),
							},
						},
						Summary:      "Get all the things",
						Description:  "Get all the items",
						OperationID:  "getAll",
						ExternalDocs: ExternalDocumentationObject{},
						Tags:         []string{"users"},
						Parameters: []ParameterObject{
							{
								Name:        "locale",
								In:          "path",
								Description: "Locale code",
								Required:    true,
								Schema: &SchemaObject{
									Type:        "string",
									Format:      "string",
									Description: "Locale code",
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
			wantPaths: PathsObject{
				"/{locale}": &PathItemObject{
					Post: &OperationObject{
						Responses: map[string]*ResponseObject{
							"201": &ResponseObject{
								Description: "Created",
								Content:     make(map[string]*MediaTypeObject),
							},
							"400": &ResponseObject{
								Description: "Failed",
								Content:     make(map[string]*MediaTypeObject),
							},
						},
						Summary:      "Create a user",
						Description:  "Create a user",
						OperationID:  "createUser",
						ExternalDocs: ExternalDocumentationObject{},
						Tags:         []string{"users"},
						Parameters: []ParameterObject{
							{
								Name:        "locale",
								In:          "path",
								Description: "Locale code",
								Required:    true,
								Schema: &SchemaObject{
									Type:        "string",
									Format:      "string",
									Description: "Locale code",
								},
							},
						},
						RequestBody: &RequestBodyObject{
							Content: map[string]*MediaTypeObject{
								ContentTypeJson: &MediaTypeObject{
									Schema: SchemaObject{
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
			wantPaths: PathsObject{
				"/{locale}/{id}": &PathItemObject{
					Patch: &OperationObject{
						Responses: map[string]*ResponseObject{
							"200": &ResponseObject{
								Description: "Success",
								Content:     make(map[string]*MediaTypeObject),
							},
							"400": &ResponseObject{
								Description: "Failed",
								Content:     make(map[string]*MediaTypeObject),
							},
						},
						Summary:      "Update a user",
						Description:  "Update a user",
						OperationID:  "updateUser",
						ExternalDocs: ExternalDocumentationObject{},
						Tags:         []string{"users"},
						Parameters: []ParameterObject{
							{
								Name:        "locale",
								In:          "path",
								Description: "Locale code",
								Required:    true,
								Schema: &SchemaObject{
									Type:        "string",
									Format:      "string",
									Description: "Locale code",
								},
							},
							{
								Name:        "id",
								In:          "path",
								Description: "User ID",
								Required:    true,
								Schema: &SchemaObject{
									Type:        "integer",
									Format:      "int64",
									Description: "User ID",
								},
							},
						},
						RequestBody: &RequestBodyObject{
							Content: map[string]*MediaTypeObject{
								ContentTypeJson: &MediaTypeObject{
									Schema: SchemaObject{
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
			wantPaths: PathsObject{
				"/{locale}/{id}": &PathItemObject{
					Put: &OperationObject{
						Responses: map[string]*ResponseObject{
							"200": &ResponseObject{
								Description: "Success",
								Content:     make(map[string]*MediaTypeObject),
							},
							"400": &ResponseObject{
								Description: "Failed",
								Content:     make(map[string]*MediaTypeObject),
							},
						},
						Summary:      "Replace a user",
						Description:  "Replace a user",
						OperationID:  "replaceUser",
						ExternalDocs: ExternalDocumentationObject{},
						Tags:         []string{"users"},
						Parameters: []ParameterObject{
							{
								Name:        "locale",
								In:          "path",
								Description: "Locale code",
								Required:    true,
								Schema: &SchemaObject{
									Type:        "string",
									Format:      "string",
									Description: "Locale code",
								},
							},
							{
								Name:        "id",
								In:          "path",
								Description: "User ID",
								Required:    true,
								Schema: &SchemaObject{
									Type:        "integer",
									Format:      "int64",
									Description: "User ID",
								},
							},
						},
						RequestBody: &RequestBodyObject{
							Content: map[string]*MediaTypeObject{
								ContentTypeJson: &MediaTypeObject{
									Schema: SchemaObject{
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
			wantPaths: PathsObject{
				"/{locale}/{id}": &PathItemObject{
					Delete: &OperationObject{
						Responses: map[string]*ResponseObject{
							"200": &ResponseObject{
								Description: "Success",
								Content:     make(map[string]*MediaTypeObject),
							},
							"400": &ResponseObject{
								Description: "Failed",
								Content:     make(map[string]*MediaTypeObject),
							},
						},
						Summary:      "Delete a user",
						Description:  "Delete a user",
						OperationID:  "deleteUser",
						ExternalDocs: ExternalDocumentationObject{},
						Tags:         []string{"users"},
						Parameters: []ParameterObject{
							{
								Name:        "locale",
								In:          "path",
								Description: "Locale code",
								Required:    true,
								Schema: &SchemaObject{
									Type:        "string",
									Format:      "string",
									Description: "Locale code",
								},
							},
							{
								Name:        "id",
								In:          "path",
								Description: "User ID",
								Required:    true,
								Schema: &SchemaObject{
									Type:        "integer",
									Format:      "int64",
									Description: "User ID",
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
			wantPaths: PathsObject{
				"/{locale}/{id}": &PathItemObject{
					Options: &OperationObject{
						Responses: map[string]*ResponseObject{
							"200": &ResponseObject{
								Description: "Success",
								Content:     make(map[string]*MediaTypeObject),
							},
							"400": &ResponseObject{
								Description: "Failed",
								Content:     make(map[string]*MediaTypeObject),
							},
						},
						Summary:      "User pre-flight",
						Description:  "User pre-flight",
						ExternalDocs: ExternalDocumentationObject{},
						Tags:         []string{"users"},
						Parameters: []ParameterObject{
							{
								Name:        "locale",
								In:          "path",
								Description: "Locale code",
								Required:    true,
								Schema: &SchemaObject{
									Type:        "string",
									Format:      "string",
									Description: "Locale code",
								},
							},
							{
								Name:        "id",
								In:          "path",
								Description: "User ID",
								Required:    true,
								Schema: &SchemaObject{
									Type:        "integer",
									Format:      "int64",
									Description: "User ID",
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
			wantPaths: PathsObject{
				"/{locale}/{id}": &PathItemObject{
					Head: &OperationObject{
						Responses:    make(map[string]*ResponseObject),
						Summary:      "User Head Lookup",
						Description:  "User Head Lookup",
						ExternalDocs: ExternalDocumentationObject{},
						Tags:         []string{"users"},
						Parameters: []ParameterObject{
							{
								Name:        "locale",
								In:          "path",
								Description: "Locale code",
								Required:    true,
								Schema: &SchemaObject{
									Type:        "string",
									Format:      "string",
									Description: "Locale code",
								},
							},
							{
								Name:        "id",
								In:          "path",
								Description: "User ID",
								Required:    true,
								Schema: &SchemaObject{
									Type:        "integer",
									Format:      "int64",
									Description: "User ID",
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
			wantPaths: PathsObject{
				"/{locale}/{id}": &PathItemObject{
					Trace: &OperationObject{
						Responses:    make(map[string]*ResponseObject),
						Summary:      "User Trace (should be disabled)",
						Description:  "User Trace (should be disabled)",
						ExternalDocs: ExternalDocumentationObject{},
						Tags:         []string{"users"},
						Parameters: []ParameterObject{
							{
								Name:        "locale",
								In:          "path",
								Description: "Locale code",
								Required:    true,
								Schema: &SchemaObject{
									Type:        "string",
									Format:      "string",
									Description: "Locale code",
								},
							},
							{
								Name:        "id",
								In:          "path",
								Description: "User ID",
								Required:    true,
								Schema: &SchemaObject{
									Type:        "integer",
									Format:      "int64",
									Description: "User ID",
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
			p, err := partialBootstrap(t)
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

func partialBootstrap(t *testing.T) (*parser, error) {
	p, err := NewParser(
		"./",
		"./main.go",
		"",
		false,
	)
	if err != nil {
		return nil, err
	}
	if err = p.parseModule(); err != nil {
		return nil, err
	}
	if err = p.parseGoMod(); err != nil {
		return nil, err
	}
	if err = p.parseAPIs(); err != nil {
		return nil, err
	}

	return p, nil
}
