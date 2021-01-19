package main

import (
	"errors"
	"github.com/stretchr/testify/assert"
	"go/ast"
	"sort"
	"testing"
)

func TestParseParamComment(t *testing.T) {
	type test struct {
		pkgPath   string
		pkgName   string
		comment   string
		want      ParameterObject
		expectErr error
	}

	var tests = []test{
		{
			pkgPath: "/",
			pkgName: "main",
			comment: `locale   path   string   true   "Locale code"`,
			want: ParameterObject{
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
			expectErr: nil,
		},
		{
			pkgPath: "/",
			pkgName: "main",
			comment: `locale   path   string   true`,
			want: ParameterObject{
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
			expectErr: errors.New(`parseParamComment can not parse param comment "locale   path   string   true"`),
		},
	}

	p := &parser{}
	for _, tc := range tests {
		op := &OperationObject{
			Responses:    nil,
			Tags:         nil,
			Summary:      "",
			Description:  "",
			Parameters:   nil,
			RequestBody:  nil,
			OperationID:  "",
			ExternalDocs: ExternalDocumentationObject{},
			Security:     nil,
		}
		if err := p.parseParamComment(tc.pkgPath, tc.pkgName, op, tc.comment); err != nil {
			assert.Equal(t, tc.expectErr, err)
			return
		}

		assert.Equal(t, tc.want, op.Parameters[0])
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
				Variables:   make(map[string]ServerVariableObject, 0),
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
				Variables:   make(map[string]ServerVariableObject, 0),
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
	type test struct {
		comments []string
		want     InfoObject
	}

	tests := []test{
		{ // test minimum required info only
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
		},
		{ // test partially populated contact and license info
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
		},
		{ // test all populated info properties
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
		},
	}

	for _, tc := range tests {
		p := &parser{}

		fileComments := commentSliceToCommentGroup(tc.comments)

		if err := p.parseInfo(fileComments); err != nil {
			t.Error(err)
		}

		assert.Equal(t, tc.want, p.OpenAPI.Info)
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

func TestParseInfoServers(t *testing.T) {
	type test struct {
		comments []string
		want     []ServerObject
	}

	emptyServerVariableMap := make(map[string]ServerVariableObject, 0)
	serverVariableMap := make(map[string]ServerVariableObject, 1)
	serverVariableMap["username"] = ServerVariableObject{
		Enum:        nil,
		Default:     "empty",
		Description: "Dev site username",
	}

	tests := []test{
		{ // Single Server
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
		},
		{ // Multiple Servers
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
		},
		{ // Multiple Servers with one server variable
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
		},
	}

	for _, tc := range tests {
		p := &parser{}

		fileComments := commentSliceToCommentGroup(tc.comments)

		if err := p.parseInfo(fileComments); err != nil {
			t.Error(err)
		}

		sort.Slice(p.OpenAPI.Servers, func(i, j int) bool {
			return p.OpenAPI.Servers[i].URL < p.OpenAPI.Servers[j].URL
		})

		assert.Equal(t, tc.want, p.OpenAPI.Servers)
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
				"// @Security AuthorizationHeader read write",
				"// @Security AuthorizationToken read write",
			},
			wantSecurity: []map[string][]string{
				{
					"AuthorizationHeader": []string{
						"read",
						"write",
					},
				},
				{
					"AuthorizationToken": []string{
						"read",
						"write",
					},
				},
			},
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
				"// @SecurityScheme OAuth oauth2AuthCode /auth /token",
			},
			wantSecurityScheme: map[string]*SecuritySchemeObject{
				"OAuth": &SecuritySchemeObject{
					Type:             "oauth2",
					Description:      "",
					OpenIdConnectUrl: "",
					OAuthFlows: &SecuritySchemeOauthObject{
						AuthorizationCode: &SecuritySchemeOauthFlowObject{
							AuthorizationUrl: "/auth",
							TokenUrl:         "/token",
							Scopes:           map[string]string{},
						},
					},
				},
			},
			wantSecurity: nil,
		},
		"oauth2 implicit": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				"// @SecurityScheme OAuth oauth2Implicit /auth",
			},
			wantSecurityScheme: map[string]*SecuritySchemeObject{
				"OAuth": &SecuritySchemeObject{
					Type:             "oauth2",
					Description:      "",
					OpenIdConnectUrl: "",
					OAuthFlows: &SecuritySchemeOauthObject{
						Implicit: &SecuritySchemeOauthFlowObject{
							AuthorizationUrl: "/auth",
							Scopes:           map[string]string{},
						},
					},
				},
			},
			wantSecurity: nil,
		},
		"oauth2 resource owner credentials": {
			comments: []string{
				"// @Title Test Run",
				"// @Version 1.0.0",
				"// @Description This is a test",
				"// @SecurityScheme OAuth oauth2ResourceOwnerCredentials /token",
			},
			wantSecurityScheme: map[string]*SecuritySchemeObject{
				"OAuth": &SecuritySchemeObject{
					Type:             "oauth2",
					Description:      "",
					OpenIdConnectUrl: "",
					OAuthFlows: &SecuritySchemeOauthObject{
						ResourceOwnerPassword: &SecuritySchemeOauthFlowObject{
							TokenUrl: "/token",
							Scopes: map[string]string{},
						},
					},
				},
			},
			wantSecurity: nil,
		},
		//"oauth2 client credentials": {},
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

func TestParseInfoTags(t *testing.T) {

}
