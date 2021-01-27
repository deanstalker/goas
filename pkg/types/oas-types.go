package types

import (
	"github.com/iancoleman/orderedmap"
)

const (
	OpenAPIVersion = "3.0.0"

	ContentTypeText = "text/plain"
	ContentTypeJSON = "application/json"
	ContentTypeForm = "multipart/form-data"

	AttributeTitle        = "@title"
	AttributeVersion      = "@version"
	AttributeDescription  = "@description"
	AttributeTOSURL       = "@termsofserviceurl"
	AttributeContactName  = "@contactname"
	AttributeContactEmail = "@contactemail"
	AttributeContactURL   = "@contacturl"

	AttributeLicenseName = "@licensename"
	AttributeLicenseURL  = "@licenseurl"

	AttributeServer         = "@server"
	AttributeServerVariable = "@servervariable"

	AttributeSecurity       = "@security"
	AttributeSecurityScheme = "@securityscheme"
	AttributeSecurityScope  = "@securityscope"

	AttributeExternalDoc = "@externaldoc"
	AttributeTag         = "@tag"

	AttributeHidden = "@hidden"

	AttributeParam   = "@param"
	AttributeSuccess = "@success"
	AttributeFailure = "@failure"

	AttributeID = "@id"

	AttributeResource = "@resource"
	AttributeRoute    = "@route"
	AttributeRouter   = "@router"

	KeywordRequired = "required"

	InFile  = "file"
	InFiles = "files"
	InForm  = "form"
	InBody  = "body"
	InPath  = "path"

	TypeBoolean = "boolean"
	TypeInteger = "integer"
	TypeNumber  = "number"
	TypeObject  = "object"
	TypeArray   = "array"

	DefaultFieldName = "key"

	GoTypeTime = "time.Time"

	MessageInvalidExample = "invalid example"
)

// GoTypesOASTypes conversion map
var GoTypesOASTypes = map[string]string{
	"bool":    "boolean",
	"uint":    "integer",
	"uint8":   "integer",
	"uint16":  "integer",
	"uint32":  "integer",
	"uint64":  "integer",
	"int":     "integer",
	"int8":    "integer",
	"int16":   "integer",
	"int32":   "integer",
	"int64":   "integer",
	"float32": "number",
	"float64": "number",
	"string":  "string",
}

// IsGoTypeOASType converts go types to openapi types
func IsGoTypeOASType(typeName string) bool {
	_, ok := GoTypesOASTypes[typeName]
	return ok
}

type OpenAPIObject struct {
	OpenAPI string         `json:"openapi"` // Required
	Info    InfoObject     `json:"info"`    // Required
	Servers []ServerObject `json:"servers,omitempty"`
	Paths   PathsObject    `json:"paths"` // Required

	Components ComponentsObject      `json:"components,omitempty"` // Required for Authorization header
	Security   []map[string][]string `json:"security,omitempty"`

	Tags         []TagObject                  `json:"tags,omitempty"`
	ExternalDocs *ExternalDocumentationObject `json:"externalDocs,omitempty"`
}

type ServerObject struct {
	URL         string `json:"url"`
	Description string `json:"description,omitempty"`

	Variables map[string]ServerVariableObject `json:"variables,omitempty"`
}

type ServerVariableObject struct {
	Enum        []string `json:"enum,omitempty"`
	Default     string   `json:"default"`
	Description string   `json:"description,omitempty"`
}

type InfoObject struct {
	Title          string         `json:"title"`
	Description    string         `json:"description,omitempty"`
	TermsOfService string         `json:"termsOfService,omitempty"`
	Contact        *ContactObject `json:"contact,omitempty"`
	License        *LicenseObject `json:"license,omitempty"`
	Version        string         `json:"version"`
}

type ContactObject struct {
	Name  string `json:"name,omitempty"`
	URL   string `json:"url,omitempty"`
	Email string `json:"email,omitempty"`
}

type LicenseObject struct {
	Name string `json:"name,omitempty"`
	URL  string `json:"url,omitempty"`
}

type PathsObject map[string]*PathItemObject

type PathItemObject struct {
	Ref         string           `json:"$ref,omitempty"`
	Summary     string           `json:"summary,omitempty"`
	Description string           `json:"description,omitempty"`
	Get         *OperationObject `json:"get,omitempty"`
	Post        *OperationObject `json:"post,omitempty"`
	Patch       *OperationObject `json:"patch,omitempty"`
	Put         *OperationObject `json:"put,omitempty"`
	Delete      *OperationObject `json:"delete,omitempty"`
	Options     *OperationObject `json:"options,omitempty"`
	Head        *OperationObject `json:"head,omitempty"`
	Trace       *OperationObject `json:"trace,omitempty"`

	// Servers
	// Parameters
}

type OperationObject struct {
	Responses ResponsesObject `json:"responses"` // Required

	Tags        []string           `json:"tags,omitempty"`
	Summary     string             `json:"summary,omitempty"`
	Description string             `json:"description,omitempty"`
	Parameters  []ParameterObject  `json:"parameters,omitempty"`
	RequestBody *RequestBodyObject `json:"requestBody,omitempty"`
	OperationID string             `json:"operationId,omitempty"`

	ExternalDocs *ExternalDocumentationObject `json:"externalDocs,omitempty"`
	Security     []map[string][]string        `json:"security,omitempty"`
	Servers      []ServerObject               `json:"servers,omitempty"` // TODO implement parser

	Deprecated bool `json:"deprecated,omitempty"`
	// Callbacks
}

type ParameterObject struct {
	Name string `json:"name"` // Required
	In   string `json:"in"`   // Required. Possible values are "query", "header", "path" or "cookie"

	Description string        `json:"description,omitempty"`
	Required    bool          `json:"required,omitempty"`
	Example     interface{}   `json:"example,omitempty"`
	Schema      *SchemaObject `json:"schema,omitempty"`

	// Ref is used when ParameterObject is a ReferenceObject
	Ref string `json:"$ref,omitempty"`

	Deprecated      bool `json:"deprecated,omitempty"`
	AllowEmptyValue bool `json:"allowEmptyValue,omitempty"`
	// Style
	// Explode
	// AllowReserved
	// Examples
	// Content
}

type ReferenceObject struct {
	Ref string `json:"$ref,omitempty"`
}

type RequestBodyObject struct {
	Content map[string]*MediaTypeObject `json:"content"` // Required

	Description string `json:"description,omitempty"`
	Required    bool   `json:"required,omitempty"`

	// Ref is used when RequestBodyObject is as a ReferenceObject
	Ref string `json:"$ref,omitempty"`
}

type MediaTypeObject struct {
	Schema SchemaObject `json:"schema,omitempty"`
	// Example string       `json:"example,omitempty"`

	// Examples
	// Encoding
}

type SchemaObject struct {
	ID                 string              `json:"-"` // For goas
	PkgName            string              `json:"-"` // For goas
	FieldName          string              `json:"-"` // For goas
	DisabledFieldNames map[string]struct{} `json:"-"` // For goas

	Type         string                       `json:"type,omitempty"`
	Format       string                       `json:"format,omitempty"`
	Required     []string                     `json:"required,omitempty"`
	Properties   *orderedmap.OrderedMap       `json:"properties,omitempty"`
	Description  string                       `json:"description,omitempty"`
	Items        *SchemaObject                `json:"items,omitempty"` // use ptr to prevent recursive error
	Example      interface{}                  `json:"example,omitempty"`
	Deprecated   bool                         `json:"deprecated,omitempty"`
	ExternalDocs *ExternalDocumentationObject `json:"externalDocs,omitempty"`

	Title string `json:"title,omitempty"`

	MultipleOf           interface{}        `json:"multipleOf,omitempty"`
	Minimum              interface{}        `json:"minimum,omitempty"`
	Maximum              interface{}        `json:"maximum,omitempty"`
	ExclusiveMinimum     bool               `json:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum     bool               `json:"exclusiveMaximum,omitempty"`
	MaxLength            interface{}        `json:"maxLength,omitempty"`
	MinLength            interface{}        `json:"minLength,omitempty"`
	Pattern              string             `json:"pattern,omitempty"`
	MaxItems             int                `json:"maxItems,omitempty"`
	MinItems             int                `json:"minItems,omitempty"`
	UniqueItems          bool               `json:"uniqueItems,omitempty"`
	MaxProperties        int                `json:"maxProperties,omitempty"`
	MinProperties        int                `json:"minProperties,omitempty"`
	Enum                 []string           `json:"enum,omitempty"`
	AllOf                []*ReferenceObject `json:"allOf,omitempty"`
	OneOf                []*ReferenceObject `json:"oneOf,omitempty"`
	AnyOf                []*ReferenceObject `json:"anyOf,omitempty"`
	Not                  *SchemaObject      `json:"not,omitempty"`
	AdditionalProperties *SchemaObject      `json:"additionalProperties,omitempty"`
	Default              interface{}        `json:"default,omitempty"`
	Nullable             bool               `json:"nullable,omitempty"`
	ReadOnly             bool               `json:"readOnly,omitempty"`
	WriteOnly            bool               `json:"writeOnly"`
	Discriminator        *Discriminator     `json:"discriminator,omitempty"`

	// Ref is used when SchemaObject is used as a ReferenceObject
	Ref string `json:"$ref,omitempty"`

	// XML
}

type Discriminator struct {
	PropertyName string `json:"propertyName"`
}

type ResponsesObject map[string]*ResponseObject // [status]ResponseObject

type ResponseObject struct {
	Description string `json:"description"` // Required

	Headers map[string]*HeaderObject    `json:"headers,omitempty"`
	Content map[string]*MediaTypeObject `json:"content,omitempty"`

	// Ref is for ReferenceObject
	Ref string `json:"$ref,omitempty"`

	// Links
}

type HeaderObject struct {
	Description string `json:"description,omitempty"`
	Type        string `json:"type,omitempty"`

	// Ref is used when HeaderObject is as a ReferenceObject
	Ref string `json:"$ref,omitempty"`
}

type ComponentsObject struct {
	Schemas         map[string]*SchemaObject         `json:"schemas,omitempty"`
	SecuritySchemes map[string]*SecuritySchemeObject `json:"securitySchemes,omitempty"`

	// The following are not populated for complexity reasons ...
	// Responses
	// Parameters
	// Examples
	// RequestBodies
	// Headers
	// Links
	// Callbacks
}

type SecuritySchemeObject struct {
	// Generic fields
	Type        string `json:"type"` // Required
	Description string `json:"description,omitempty"`

	// http
	Scheme string `json:"scheme,omitempty"`

	// apiKey
	In   string `json:"in,omitempty"`
	Name string `json:"name,omitempty"`

	// OpenID
	OpenIDConnectURL string `json:"openIdConnectUrl,omitempty"`

	// OAuth2
	OAuthFlows *SecuritySchemeOauthObject `json:"flows,omitempty"`

	// BearerFormat
}

type SecuritySchemeOauthObject struct {
	Implicit              *SecuritySchemeOauthFlowObject `json:"implicit,omitempty"`
	AuthorizationCode     *SecuritySchemeOauthFlowObject `json:"authorizationCode,omitempty"`
	ResourceOwnerPassword *SecuritySchemeOauthFlowObject `json:"password,omitempty"`
	ClientCredentials     *SecuritySchemeOauthFlowObject `json:"clientCredentials,omitempty"`
}

func (s *SecuritySchemeOauthObject) ApplyScopes(scopes map[string]string) {
	if s.Implicit != nil {
		s.Implicit.Scopes = scopes
	}

	if s.AuthorizationCode != nil {
		s.AuthorizationCode.Scopes = scopes
	}

	if s.ResourceOwnerPassword != nil {
		s.ResourceOwnerPassword.Scopes = scopes
	}

	if s.ClientCredentials != nil {
		s.ClientCredentials.Scopes = scopes
	}
}

type SecuritySchemeOauthFlowObject struct {
	AuthorizationURL string            `json:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

type ExternalDocumentationObject struct {
	Description string `json:"description,omitempty"`
	URL         string `json:"url"`
}

type TagObject struct {
	Name         string                       `json:"name"`
	Description  string                       `json:"description,omitempty"`
	ExternalDocs *ExternalDocumentationObject `json:"externalDocs,omitempty"`
}
