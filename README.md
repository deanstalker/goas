# goas

The project is based on  
- [yvasiyarov/swagger](https://github.com/yvasiyarov/swagger) repository.  
- [uudashr/go-module](https://github.com/uudashr/go-module) repository. (currently deprecated)
- [mikunalpha/goas](https://github.com/mikunalpha/goas) repository
- [nicocesar/goas](https://github.com/nicocesar/goas) repository
- [launchdarkly/goas](https://github.com/launchdarkly/goas) repository

This fork takes the goas library further with a number of missing spec objects, validations and tests filled in.

Generate [OpenAPI Specification](https://swagger.io/specification) json file with comments in Go.

## Limit
- Only support go module.
- Anonymous struct field is not supported.

## CI

[![Build Status](https://travis-ci.com/deanstalker/goas.svg?branch=master)](https://travis-ci.com/deanstalker/goas)
[![codecov](https://codecov.io/gh/deanstalker/goas/branch/master/graph/badge.svg?token=SAV4O3BSEU)](https://codecov.io/gh/deanstalker/goas)

## Install

```
go get -u github.com/deanstalker/goas
```

## Usage

You can document your service by placing annotations inside your godoc at various places in your code.

### Generating Documentation

```
NAME:
   goas

USAGE:
   goas [global options] [arguments...]

VERSION:
   v1.0.0

GLOBAL OPTIONS:
   --module-path value     goas will search @comment under the module
   --main-file-path value  goas will start to search @comment from this main file
   --handler-path value    goas only search handleFunc comments under the path
   --output value          output file
   --format value          json (default) or yaml format - for stdout only (default: "json")
   --debug                 show debug message
   --version, -v           print the version

COPYRIGHT:
   (c) 2018 mikun800527@gmail.com
```

Go to the root folder of your project
```
// go.mod and main file are in the same directory
goas --module-path . --output oas.json

// go.mod and main file are in a different directory
goas --module-path . --main-file-path ./cmd/xxx/main.go --output oas.json

// output spec to file in yaml format
goas --module-path . --output oas.yaml

// output spec to stdout in yaml format
goas --module-path . --format yaml 2>&1
```

#### Using go generate

* Create a new folder called `docs` under your project's root directory
* Create a new file ie. `docs.go`, `api.go`, etc. in the `docs` folder.
* In the file, add the following line underneath the package name (customise the command itself to suit your needs)
```go
//go:generate goas --module-path ../ --main-file-path ../docs/api.go --output ./api.yaml --handler-path ../pkg/api
```
* Go back to your root folder, and run
```sh
go generate ./docs
```


### Service Description

The service description comments can be located in any of your .go files. They provide general information about the service you are documenting.

```go
// @Version 1.0.0
// @Title Backend API
// @Description API usually works as expected. But sometimes its not true.
// @ContactName Abcd
// @ContactEmail abce@email.com
// @ContactURL http://someurl.oxox
// @TermsOfServiceUrl http://someurl.oxox
// @LicenseName MIT
// @LicenseURL https://en.wikipedia.org/wiki/MIT_License
// @Server http://www{env}fake.com Server-1
// @Server http://www.fake2.com Server-2
// @ServerVariable env "." "Default environment, . for prod" ".,staging.,dev."
// @Security AuthorizationHeader read write
// @SecurityScheme AuthorizationHeader http bearer Input your token
// @SecurityScope AuthorizationHeader read "Read-only"
// @SecurityScope AuthorizationHeader write "Writable"
```

#### Server Variables

If a server variable is required, add the placeholder to the `@Server` comment's url. Then reference the placeholder in the
`@ServerVariable` tag as follows:

```go
// @Server http://{subdomain}.fake.com Server 1
// @ServerVariable subdomain "www" "Subdomains" "www,dev"
```

The Server Variable will only be applied to Server objects that contain a url with the nominated placeholder/s. 

#### Security

If authorization is required, you must define security schemes and then apply those to the API.
A scheme is defined using `@SecurityScheme [name] [type] [parameters]` and applied by adding `@Security [scheme-name] [scope1] [scope2] [...]`. 

All examples in this section use `MyApiAuth` as the name. This name can be anything you chose; multiple named schemes are supported.
Each scheme must have its own name, except for OAuth2 schemes - OAuth2 supports multiple schemes by the same name.

A number of different types is supported, they all have different parameters:

|Type|Description|Parameters|Example|
|---|---|---|---|
|HTTP|A HTTP Authentication scheme using the `Authorization` header|scheme: any [HTTP Authentication scheme](https://www.iana.org/assignments/http-authschemes/http-authschemes.xhtml)|`@SecurityScheme MyApiAuth basic`|
|APIKey|Authorization by passing an API Key along with the request|in: Location of the API Key, options are `header`, `query` and `cookie`. name: The name of the field where the API Key must be set|`@SecurityScheme MyApiAuth apiKey header X-MyCustomHeader`|
|OpenIdConnect|Delegating security to a known OpenId server|url: The URL of the OpenId server|`@SecurityScheme MyApiAuth openIdConnect https://example.com/.well-known/openid-configuration`|
|OAuth2AuthCode|Using the "Authentication Code" flow of OAuth2|authorizationUrl, tokenUrl|`@SecurityScheme MyApiAuth oauth2AuthCode /oauth/authorize /oauth/token`| 
|OAuth2Implicit|Using the "Implicit" flow of OAuth2|authorizationUrl|`@SecurityScheme MyApiAuth oauth2Implicit /oauth/authorize| 
|OAuth2ResourceOwnerCredentials|Using the "Resource Owner Credentials" flow of OAuth2|authorizationUrl|`@SecurityScheme MyApiAuth oauth2ResourceOwnerCredentials /oauth/token| 
|OAuth2ClientCredentials|Using the "Client Credentials" flow of OAuth2|authorizationUrl|`@SecurityScheme MyApiAuth oauth2ClientCredentials /oauth/token| 

Any text that is present after the last parameter wil be used as the description. For instance `@SecurityScheme MyApiAuth basic Login with your admin credentials`.

Once all security schemes have been defined, they must be configured. This is done with the `@Security` comment.
Depending on the `type` of the scheme, scopes (see below) may be supported. *At the moment, it is only possible to configure security for the entire service*.

```go
// @Security MyApiAuth read_user write_user
```

##### Scopes
For OAuth2 security schemes, it is possible to define scopes using the `@SecurityScope [schema-name] [scope-code] [scope-description]` comment.

```go
// @SecurityScope MyApiAuth read_user Read a user from the system
// @SecurityScope MyApiAuth write_user Write a user to the system
```

### Handler funcs

By adding comments to your handler func godoc, you can document individual actions as well as their input and output.

```go
package handler

type User struct {
  ID   uint64 `json:"id" example:"100" description:"User identity"`
  Name string `json:"name" example:"Mikun"` 
}

type UsersResponse struct {
  Data []User `json:"users" example:"[{\"id\":100, \"name\":\"Mikun\"}]"`
}

type Error struct {
  Code string `json:"code"`
  Msg  string `json:"msg"`
}

type ErrorResponse struct {
  ErrorInfo Error `json:"error"`
}

// @Title Get user list of a group.
// @Description Get users related to a specific group.
// @Param  groupID  path  int  true  "Id of a specific group."
// @Success  200  object  UsersResponse  "UsersResponse JSON"
// @Failure  400  object  ErrorResponse  "ErrorResponse JSON"
// @Resource users
// @Route /api/group/{groupID}/users [get]
func GetGroupUsers() {
  // ...
}

// @Title Get user list of a group.
// @Description Create a new user.
// @Param  user  body  User  true  "Info of a user."
// @Success  200  object  User           "UsersResponse JSON"
// @Failure  400  object  ErrorResponse  "ErrorResponse JSON"
// @Resource users
// @Route /api/user [post]
func PostUser() {
  // ...
}
```

#### Title & Description
```
@Title {title}
@Title Get user list of a group.

@Description {description}.
@Description Get users related to a specific group.
```
- {title}: The title of the route.
- {description}: The description of the route.

#### Parameter
```
@Param  {name}  {in}  {goType}  {required}  {description}
@Param  user    body  User      true        "Info of a user."
```
- {name}: The parameter name.
- {in}: The parameter is in `path`, `query`, `form`, `header`, `cookie`, `body` or `file`.
- {goType}: The type in go code. This will be ignored when {in} is `file`.
- {required}: `true`, `false`, `required` or `optional`. 
- {description}: The description of the parameter. Must be quoted.

#### Response
```
@Success  {status}  {jsonType}  {goType}       {description}
@Success  200       object      UsersResponse  "UsersResponse JSON"

@Failure  {status}  {jsonType}  {goType}       {description}
@Failure  400       object      ErrorResponse  "ErrorResponse JSON"
```
- {status}: The HTTP status code.
- {jsonType}: The value can be `object` or `array`. 
- {goType}: The type in go code.
- {description}: The description of the response. Must be quoted.

#### Resource & Tag
```
@Resource {resource}
@Resource users

@Tag {tag}
@tag xxx
```
- {resource}, {tag}: Tag of the route.

#### Route
```
@Route {path}    {method}
@Route /api/user [post]
```
- {path}: The URL path.
- {method}: The HTTP Method. Must be put in brackets.