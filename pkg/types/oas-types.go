package types

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
	OpenAPI string         `json:"openapi" yaml:"openapi"` // Required
	Info    InfoObject     `json:"info" yaml:"info"`       // Required
	Servers []ServerObject `json:"servers,omitempty" yaml:",omitempty"`
	Paths   PathsObject    `json:"paths" yaml:"paths"` // Required

	Components ComponentsObject      `json:"components,omitempty" yaml:",omitempty"` // Required for Authorization header
	Security   []map[string][]string `json:"security,omitempty" yaml:",omitempty"`

	Tags         []TagObject                  `json:"tags,omitempty" yaml:",omitempty"`
	ExternalDocs *ExternalDocumentationObject `json:"externalDocs,omitempty" yaml:",omitempty"`
}

type ServerObject struct {
	URL         string `json:"url" yaml:"url"`
	Description string `json:"description,omitempty" yaml:",omitempty"`

	Variables map[string]ServerVariableObject `json:"variables,omitempty" yaml:",omitempty"`
}

type ServerVariableObject struct {
	Enum        []string `json:"enum,omitempty" yaml:",omitempty"`
	Default     string   `json:"default" yaml:"default"`
	Description string   `json:"description,omitempty" yaml:",omitempty"`
}

type InfoObject struct {
	Title          string         `json:"title" yaml:"title"`
	Description    string         `json:"description,omitempty" yaml:",omitempty"`
	TermsOfService string         `json:"termsOfService,omitempty" yaml:",omitempty"`
	Contact        *ContactObject `json:"contact,omitempty" yaml:",omitempty"`
	License        *LicenseObject `json:"license,omitempty" yaml:",omitempty"`
	Version        string         `json:"version" yaml:"version"`
}

type ContactObject struct {
	Name  string `json:"name,omitempty" yaml:",omitempty"`
	URL   string `json:"url,omitempty" yaml:",omitempty"`
	Email string `json:"email,omitempty" yaml:",omitempty"`
}

type LicenseObject struct {
	Name string `json:"name,omitempty" yaml:",omitempty"`
	URL  string `json:"url,omitempty" yaml:",omitempty"`
}

type PathsObject map[string]*PathItemObject

type PathItemObject struct {
	Ref         string           `json:"$ref,omitempty" yaml:"$ref,omitempty"`
	Summary     string           `json:"summary,omitempty" yaml:",omitempty"`
	Description string           `json:"description,omitempty" yaml:",omitempty"`
	Get         *OperationObject `json:"get,omitempty" yaml:",omitempty"`
	Post        *OperationObject `json:"post,omitempty" yaml:",omitempty"`
	Patch       *OperationObject `json:"patch,omitempty" yaml:",omitempty"`
	Put         *OperationObject `json:"put,omitempty" yaml:",omitempty"`
	Delete      *OperationObject `json:"delete,omitempty" yaml:",omitempty"`
	Options     *OperationObject `json:"options,omitempty" yaml:",omitempty"`
	Head        *OperationObject `json:"head,omitempty" yaml:",omitempty"`
	Trace       *OperationObject `json:"trace,omitempty" yaml:",omitempty"`

	// Servers
	// Parameters
}

type OperationObject struct {
	Responses ResponsesObject `json:"responses"` // Required

	Tags        []string           `json:"tags,omitempty" yaml:",omitempty"`
	Summary     string             `json:"summary,omitempty" yaml:",omitempty"`
	Description string             `json:"description,omitempty" yaml:",omitempty"`
	Parameters  []ParameterObject  `json:"parameters,omitempty" yaml:",omitempty"`
	RequestBody *RequestBodyObject `json:"requestBody,omitempty" yaml:"requestBody,omitempty"`
	OperationID string             `json:"operationId,omitempty" yaml:"operationId,omitempty"`

	ExternalDocs *ExternalDocumentationObject `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
	Security     []map[string][]string        `json:"security,omitempty" yaml:",omitempty"`
	Servers      []ServerObject               `json:"servers,omitempty" yaml:",omitempty"` // TODO implement parser

	Deprecated bool `json:"deprecated,omitempty" yaml:",omitempty"`
	// Callbacks
}

type ParameterObject struct {
	Name string `json:"name"` // Required
	In   string `json:"in"`   // Required. Possible values are "query", "header", "path" or "cookie"

	Description string        `json:"description,omitempty" yaml:",omitempty"`
	Required    bool          `json:"required,omitempty" yaml:",omitempty"`
	Example     interface{}   `json:"example,omitempty" yaml:",omitempty"`
	Schema      *SchemaObject `json:"schema,omitempty" yaml:",omitempty"`

	// Ref is used when ParameterObject is a ReferenceObject
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`

	Deprecated      bool `json:"deprecated,omitempty" yaml:",omitempty"`
	AllowEmptyValue bool `json:"allowEmptyValue,omitempty" yaml:"allowEmptyValue,omitempty"`
	// Style
	// Explode
	// AllowReserved
	// Examples
	// Content
}

type ReferenceObject struct {
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
}

type RequestBodyObject struct {
	Content map[string]*MediaTypeObject `json:"content"` // Required

	Description string `json:"description,omitempty" yaml:",omitempty"`
	Required    bool   `json:"required,omitempty" yaml:",omitempty"`

	// Ref is used when RequestBodyObject is as a ReferenceObject
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
}

type MediaTypeObject struct {
	Schema SchemaObject `json:"schema,omitempty" yaml:",omitempty"`
	// Example string       `json:"example,omitempty"`

	// Examples
	// Encoding
}

type SchemaObject struct {
	ID                 string              `json:"-"`          // For goas
	PkgName            string              `json:"-"`          // For goas
	FieldName          string              `json:"-" yaml:"-"` // For goas
	DisabledFieldNames map[string]struct{} `json:"-" yaml:"-"` // For goas

	Type         string                       `json:"type,omitempty" yaml:",omitempty"`
	Format       string                       `json:"format,omitempty" yaml:",omitempty"`
	Required     []string                     `json:"required,omitempty" yaml:",omitempty"`
	Properties   *ChainedOrderedMap           `json:"properties,omitempty" yaml:",omitempty"`
	Description  string                       `json:"description,omitempty" yaml:",omitempty"`
	Items        *SchemaObject                `json:"items,omitempty" yaml:",omitempty"` // use ptr to prevent recursive error
	Example      interface{}                  `json:"example,omitempty" yaml:",omitempty"`
	Deprecated   bool                         `json:"deprecated,omitempty" yaml:",omitempty"`
	ExternalDocs *ExternalDocumentationObject `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`

	Title string `json:"title,omitempty" yaml:",omitempty"`

	MultipleOf           interface{}        `json:"multipleOf,omitempty" yaml:"multipleOf,omitempty"`
	Minimum              interface{}        `json:"minimum,omitempty" yaml:",omitempty"`
	Maximum              interface{}        `json:"maximum,omitempty" yaml:",omitempty"`
	ExclusiveMinimum     bool               `json:"exclusiveMinimum,omitempty" yaml:"exclusiveMinimum,omitempty"`
	ExclusiveMaximum     bool               `json:"exclusiveMaximum,omitempty" yaml:"exclusiveMaximum,omitempty"`
	MaxLength            interface{}        `json:"maxLength,omitempty" yaml:"maxLength,omitempty"`
	MinLength            interface{}        `json:"minLength,omitempty" yaml:"minLength,omitempty"`
	Pattern              string             `json:"pattern,omitempty" yaml:",omitempty"`
	MaxItems             int                `json:"maxItems,omitempty" yaml:"maxItems,omitempty"`
	MinItems             int                `json:"minItems,omitempty" yaml:"minItems,omitempty"`
	UniqueItems          bool               `json:"uniqueItems,omitempty" yaml:"uniqueItems,omitempty"`
	MaxProperties        int                `json:"maxProperties,omitempty" yaml:"maxProperties,omitempty"`
	MinProperties        int                `json:"minProperties,omitempty" yaml:"minProperties,omitempty"`
	Enum                 []string           `json:"enum,omitempty" yaml:",omitempty"`
	AllOf                []*ReferenceObject `json:"allOf,omitempty" yaml:"allOf,omitempty"`
	OneOf                []*ReferenceObject `json:"oneOf,omitempty" yaml:"oneOf,omitempty"`
	AnyOf                []*ReferenceObject `json:"anyOf,omitempty" yaml:"anyOf,omitempty"`
	Not                  *SchemaObject      `json:"not,omitempty" yaml:",omitempty"`
	AdditionalProperties *SchemaObject      `json:"additionalProperties,omitempty" yaml:"additionalProperties,omitempty"`
	Default              interface{}        `json:"default,omitempty" yaml:",omitempty"`
	Nullable             bool               `json:"nullable,omitempty" yaml:",omitempty"`
	ReadOnly             bool               `json:"readOnly,omitempty" yaml:"readOnly,omitempty"`
	WriteOnly            bool               `json:"writeOnly,omitempty" yaml:"writeOnly,omitempty"`
	Discriminator        *Discriminator     `json:"discriminator,omitempty" yaml:",omitempty"`

	// Ref is used when SchemaObject is used as a ReferenceObject
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`

	// XML
}

type Discriminator struct {
	PropertyName string `json:"propertyName" yaml:"propertyName"`
}

type ResponsesObject map[string]*ResponseObject // [status]ResponseObject

type ResponseObject struct {
	Description string `json:"description"` // Required

	Headers map[string]*HeaderObject    `json:"headers,omitempty" yaml:",omitempty"`
	Content map[string]*MediaTypeObject `json:"content,omitempty" yaml:",omitempty"`

	// Ref is for ReferenceObject
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`

	// Links
}

type HeaderObject struct {
	Description string `json:"description,omitempty" yaml:",omitempty"`
	Type        string `json:"type,omitempty" yaml:",omitempty"`

	// Ref is used when HeaderObject is as a ReferenceObject
	Ref string `json:"$ref,omitempty" yaml:"$ref,omitempty"`
}

type ComponentsObject struct {
	Schemas         map[string]*SchemaObject         `json:"schemas,omitempty" yaml:",omitempty"`
	SecuritySchemes map[string]*SecuritySchemeObject `json:"securitySchemes,omitempty" yaml:"securitySchemes,omitempty"`

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
	Description string `json:"description,omitempty" yaml:",omitempty"`

	// http
	Scheme string `json:"scheme,omitempty" yaml:",omitempty"`

	// apiKey
	In   string `json:"in,omitempty" yaml:",omitempty"`
	Name string `json:"name,omitempty" yaml:",omitempty"`

	// OpenID
	OpenIDConnectURL string `json:"openIdConnectUrl,omitempty" yaml:"openIdConnectUrl,omitempty"`

	// OAuth2
	OAuthFlows *SecuritySchemeOauthObject `json:"flows,omitempty" yaml:",omitempty"`

	// BearerFormat
}

type SecuritySchemeOauthObject struct {
	Implicit              *SecuritySchemeOauthFlowObject `json:"implicit,omitempty" yaml:",omitempty"`
	AuthorizationCode     *SecuritySchemeOauthFlowObject `json:"authorizationCode,omitempty" yaml:"authorizationCode,omitempty"`
	ResourceOwnerPassword *SecuritySchemeOauthFlowObject `json:"password,omitempty" yaml:"password,omitempty"`
	ClientCredentials     *SecuritySchemeOauthFlowObject `json:"clientCredentials,omitempty" yaml:"clientCredentials,omitempty"`
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
	AuthorizationURL string            `json:"authorizationUrl,omitempty" yaml:"authorizationUrl,omitempty"`
	TokenURL         string            `json:"tokenUrl,omitempty" yaml:"tokenUrl,omitempty"`
	Scopes           map[string]string `json:"scopes"`
}

type ExternalDocumentationObject struct {
	Description string `json:"description,omitempty" yaml:",omitempty"`
	URL         string `json:"url"`
}

type TagObject struct {
	Name         string                       `json:"name"`
	Description  string                       `json:"description,omitempty" yaml:",omitempty"`
	ExternalDocs *ExternalDocumentationObject `json:"externalDocs,omitempty" yaml:"externalDocs,omitempty"`
}
