package main

import (
	"encoding/json"
	"fmt"
	"go/ast"
	goparser "go/parser"
	"go/token"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode"

	"gopkg.in/yaml.v2"

	"github.com/deanstalker/goas/pkg/types"

	"github.com/deanstalker/goas/internal/util"

	module "golang.org/x/mod/modfile"
)

type parser struct {
	ModulePath string
	ModuleName string

	MainFilePath string

	HandlerPath string

	GoModFilePath string

	GoModCachePath string

	OpenAPI types.OpenAPIObject

	KnownPkgs         []pkg
	KnownNamePkg      map[string]*pkg
	KnownPathPkg      map[string]*pkg
	KnownIDSchema     map[string]*types.SchemaObject
	KnownOperationIDs []string

	ExcludePkgs []string

	TypeSpecs               map[string]map[string]*ast.TypeSpec
	PkgPathAstPkgCache      map[string]map[string]*ast.Package
	PkgNameImportedPkgAlias map[string]map[string][]string

	Debug bool
}

const (
	ModeStdOut     = "stdout"
	ModeFileWriter = "file"
	ModeTest       = "test"

	FormatJSON = "json"
	FormatYAML = "yaml"
)

type pkg struct {
	Name string
	Path string
}

func newParser(modulePath, mainFilePath, handlerPath, excludePackages string, debug bool) (*parser, error) {
	p := &parser{
		ExcludePkgs:             []string{},
		KnownPkgs:               []pkg{},
		KnownNamePkg:            map[string]*pkg{},
		KnownPathPkg:            map[string]*pkg{},
		KnownIDSchema:           map[string]*types.SchemaObject{},
		TypeSpecs:               map[string]map[string]*ast.TypeSpec{},
		PkgPathAstPkgCache:      map[string]map[string]*ast.Package{},
		PkgNameImportedPkgAlias: map[string]map[string][]string{},
		Debug:                   debug,
	}
	p.OpenAPI.OpenAPI = types.OpenAPIVersion
	p.OpenAPI.Paths = make(types.PathsObject)
	p.OpenAPI.Security = []map[string][]string{}
	p.OpenAPI.Components.Schemas = make(map[string]*types.SchemaObject)
	p.OpenAPI.Components.SecuritySchemes = map[string]*types.SecuritySchemeObject{}

	// check modulePath is exist
	modulePath, err := util.CheckModulePathExists(modulePath)
	if err != nil {
		return nil, fmt.Errorf("check module path failed: %v", err)
	}
	p.ModulePath = modulePath

	// check go.mod file is exist
	goModFilePath, goModFileInfo, err := util.CheckGoModExists(modulePath)
	if err != nil {
		return nil, fmt.Errorf("check go.mod file exists, failed: %v", err)
	}
	p.GoModFilePath = goModFilePath

	// check mainFilePath is exist
	mainFilePath, err = util.CheckMainFilePathExists(mainFilePath, modulePath)
	if err != nil {
		return nil, fmt.Errorf("check main file path exists, failed: %v", err)
	}
	p.MainFilePath = mainFilePath

	// get module name from go.mod file
	moduleName, err := util.GetModulePath(goModFilePath)
	if err != nil {
		return nil, fmt.Errorf("unable to get module name from go.mod file: %v", err)
	}
	if moduleName == "" {
		return nil, fmt.Errorf("cannot get module name from %s", goModFileInfo)
	}
	p.ModuleName = moduleName

	// check go module cache path is exist ($GOPATH/pkg/mod)
	goPath := os.Getenv("GOPATH")
	if goPath == "" {
		u, uerr := user.Current()
		if uerr != nil {
			return nil, fmt.Errorf("cannot get current user: %s", uerr)
		}
		goPath = filepath.Join(u.HomeDir, "go")
	}
	goModCachePath := filepath.Join(goPath, "pkg", "mod")
	goModCacheInfo, err := os.Stat(goModCachePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, err
		}
		return nil, fmt.Errorf("cannot get information of %s: %s", goModCachePath, err)
	}
	if !goModCacheInfo.IsDir() {
		return nil, fmt.Errorf("%s should be a directory", goModCachePath)
	}
	p.GoModCachePath = goModCachePath

	if handlerPath != "" {
		handlerPath, _ = filepath.Abs(handlerPath)
		_, err := os.Stat(handlerPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil, err
			}
			return nil, fmt.Errorf("cannot get information of %s: %s", handlerPath, err)
		}
	}
	p.HandlerPath = handlerPath

	p.ExcludePkgs = strings.Split(excludePackages, ",")

	return p, nil
}

func (p *parser) CreateOAS(path, mode, format string) (*string, error) {
	comments, err := p.parseFileComments()
	if err != nil {
		return nil, err
	}

	// parse basic info
	err = p.parseInfo(comments)
	if err != nil {
		return nil, err
	}

	// parse sub-package
	p.parseModule()

	// parse go.mod info
	err = p.parseGoMod()
	if err != nil {
		return nil, err
	}

	// parse APIs info
	err = p.parseAPIs()
	if err != nil {
		return nil, err
	}

	var output []byte
	switch format {
	case FormatJSON:
		output, err = json.MarshalIndent(p.OpenAPI, "", "  ")
		if err != nil {
			return nil, err
		}
	case FormatYAML:
		output, err = yaml.Marshal(p.OpenAPI)
		if err != nil {
			return nil, err
		}
	}

	var fd *os.File
	switch mode {
	case ModeFileWriter:
		fd, err = os.Create(path)
		if err != nil {
			return nil, fmt.Errorf("can not create the file %s: %v", path, err)
		}
		defer fd.Close()
		_, _ = fd.WriteString(string(output))
	case ModeStdOut:
		_, err = os.Stdout.WriteString(string(output))
	case ModeTest:
		test := string(output)
		return &test, nil
	}

	return nil, err
}

func (p *parser) parseFileComments() ([]*ast.CommentGroup, error) {
	fileTree, err := goparser.ParseFile(token.NewFileSet(), p.MainFilePath, nil, goparser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("can not parse general API information: %v", err)
	}

	return fileTree.Comments, nil
}

func (p *parser) parseSchemaComments(comments []*ast.Comment, schemaObject *types.SchemaObject) {
	for i := range comments {
		for _, comment := range strings.Split(comments[i].Text, "\n") {
			comment = strings.TrimSpace(strings.Trim(comment, "/"))
			attribute := strings.ToLower(strings.Split(comment, " ")[0])
			if attribute == "" || attribute[0] != '@' {
				continue
			}
			value := strings.TrimSpace(comment[len(attribute):])
			if value == "" {
				continue
			}
			switch attribute {
			case types.AttributeTitle:
				schemaObject.Title = value
			case types.AttributeDescription:
				schemaObject.Description = value
			}
		}
	}
}

func (p *parser) parseInfo(comments []*ast.CommentGroup) error {
	// Security Scopes are defined at a different level in the hierarchy as where they need to end up in the OpenAPI structure,
	// so a temporary list is needed.
	oauthScopes := make(map[string]map[string]string)

	for i := range comments {
		for _, comment := range strings.Split(comments[i].Text(), "\n") {
			attribute := strings.ToLower(strings.Split(comment, " ")[0])
			if attribute == "" || attribute[0] != '@' {
				continue
			}
			value := strings.TrimSpace(comment[len(attribute):])
			if value == "" {
				continue
			}
			switch attribute {
			case types.AttributeVersion:
				p.OpenAPI.Info.Version = value
			case types.AttributeTitle:
				p.OpenAPI.Info.Title = value
			case types.AttributeDescription:
				p.OpenAPI.Info.Description = value
			case types.AttributeTOSURL:
				p.OpenAPI.Info.TermsOfService = value
			case types.AttributeContactName:
				if p.OpenAPI.Info.Contact == nil {
					p.OpenAPI.Info.Contact = &types.ContactObject{}
				}
				p.OpenAPI.Info.Contact.Name = value
			case types.AttributeContactEmail:
				if p.OpenAPI.Info.Contact == nil {
					p.OpenAPI.Info.Contact = &types.ContactObject{}
				}
				p.OpenAPI.Info.Contact.Email = value
			case types.AttributeContactURL:
				if p.OpenAPI.Info.Contact == nil {
					p.OpenAPI.Info.Contact = &types.ContactObject{}
				}
				p.OpenAPI.Info.Contact.URL = value
			case types.AttributeLicenseName:
				if p.OpenAPI.Info.License == nil {
					p.OpenAPI.Info.License = &types.LicenseObject{}
				}
				p.OpenAPI.Info.License.Name = value
			case types.AttributeLicenseURL:
				if p.OpenAPI.Info.License == nil {
					p.OpenAPI.Info.License = &types.LicenseObject{}
				}
				p.OpenAPI.Info.License.URL = value
			case types.AttributeServer:
				fields := strings.Split(value, " ")
				_, err := url.ParseRequestURI(fields[0])
				// allow server variable tokens through
				if err != nil && !strings.Contains(fields[0], "{") {
					return fmt.Errorf(`server: "%s" is not a valid URL`, fields[0])
				}
				s := types.ServerObject{
					URL:         fields[0],
					Description: strings.TrimSpace(value[len(fields[0]):]),
				}
				p.OpenAPI.Servers = append(p.OpenAPI.Servers, s)
			case types.AttributeSecurity:
				fields := strings.Split(value, " ")
				security := map[string][]string{
					fields[0]: fields[1:],
				}
				p.OpenAPI.Security = append(p.OpenAPI.Security, security)
			case types.AttributeSecurityScheme:
				p.parseSecurityScheme(value)
			case types.AttributeSecurityScope:
				fields := strings.Split(value, " ")

				if _, ok := oauthScopes[fields[0]]; !ok {
					oauthScopes[fields[0]] = make(map[string]string)
				}

				oauthScopes[fields[0]][fields[1]] = strings.Join(fields[2:], " ")
			case types.AttributeExternalDoc:
				externalDocs, err := p.parseExternalDocComment(strings.TrimSpace(comment[len(attribute):]))
				if err != nil {
					return err
				}
				if externalDocs == nil {
					return fmt.Errorf("couldn't populate externalDocs")
				}

				p.OpenAPI.ExternalDocs = externalDocs
			case types.AttributeTag:
				tag, err := p.parseTagComment(strings.TrimSpace(comment[len(attribute):]))
				if err != nil {
					return fmt.Errorf("%v", err)
				}

				p.OpenAPI.Tags = append(p.OpenAPI.Tags, *tag)
			case types.AttributeServerVariable:
				for i, server := range p.OpenAPI.Servers {
					if server.Variables == nil {
						server.Variables = make(map[string]types.ServerVariableObject)
					}
					server.Variables, _ = p.parseServerVariableComment(comment, server)

					p.OpenAPI.Servers[i] = server
				}
			}
		}
	}

	// Apply security scopes to their security schemes
	p.applySecurityScopes(oauthScopes)

	if err := p.validateInfo(); err != nil {
		return err
	}

	return nil
}

func (p *parser) validateInfo() error {
	if p.OpenAPI.Info.Title == "" {
		return fmt.Errorf("info.title cannot not be empty")
	}
	if p.OpenAPI.Info.Version == "" {
		return fmt.Errorf("info.version cannot not be empty")
	}
	for i := range p.OpenAPI.Servers {
		if p.OpenAPI.Servers[i].URL == "" {
			return fmt.Errorf("servers[%d].url cannot not be empty", i)
		}
	}
	return nil
}

func (p *parser) applySecurityScopes(oauthScopes map[string]map[string]string) {
	for scheme := range p.OpenAPI.Components.SecuritySchemes {
		if p.OpenAPI.Components.SecuritySchemes[scheme].Type == "oauth2" {
			if scopes, ok := oauthScopes[scheme]; ok {
				p.OpenAPI.Components.SecuritySchemes[scheme].OAuthFlows.ApplyScopes(scopes)
			}
		}
	}
}

func (p *parser) parseModule() {
	walker := func(path string, info os.FileInfo, err error) error {
		if info != nil && info.IsDir() {
			if strings.HasPrefix(strings.Trim(strings.TrimPrefix(path, p.ModulePath), "/"), ".git") {
				return nil
			}
			fns, err := filepath.Glob(filepath.Join(path, "*.go"))
			if len(fns) == 0 || err != nil {
				return nil
			}

			name := filepath.Join(p.ModuleName, strings.TrimPrefix(path, p.ModulePath))
			name = filepath.ToSlash(name)

			for _, excludeName := range p.ExcludePkgs {
				if strings.EqualFold(name, excludeName) {
					return nil
				}
			}
			p.KnownPkgs = append(p.KnownPkgs, pkg{
				Name: name,
				Path: path,
			})
			p.KnownNamePkg[name] = &p.KnownPkgs[len(p.KnownPkgs)-1]
			p.KnownPathPkg[path] = &p.KnownPkgs[len(p.KnownPkgs)-1]
		}
		return nil
	}
	_ = filepath.Walk(p.ModulePath, walker)
}
func fixer(path, version string) (string, error) {
	_ = path
	return version, nil
}

func (p *parser) parseGoMod() error {
	b, err := ioutil.ReadFile(p.GoModFilePath)
	if err != nil {
		return err
	}
	goMod, err := module.ParseLax(p.GoModFilePath, b, fixer)
	if err != nil {
		return err
	}
	for i := range goMod.Require {
		var pathRunes []rune
		for _, v := range goMod.Require[i].Mod.Path {
			if !unicode.IsUpper(v) {
				pathRunes = append(pathRunes, v)
				continue
			}
			pathRunes = append(pathRunes, '!', unicode.ToLower(v))
		}
		pkgName := goMod.Require[i].Mod.Path
		pkgPath := filepath.Join(p.GoModCachePath, string(pathRunes)+"@"+goMod.Require[i].Mod.Version)
		pkgName = filepath.ToSlash(pkgName)
		p.KnownPkgs = append(p.KnownPkgs, pkg{
			Name: pkgName,
			Path: pkgPath,
		})
		p.KnownNamePkg[pkgName] = &p.KnownPkgs[len(p.KnownPkgs)-1]
		p.KnownPathPkg[pkgPath] = &p.KnownPkgs[len(p.KnownPkgs)-1]

		walker := func(path string, info os.FileInfo, err error) error {
			if info != nil && info.IsDir() {
				if strings.HasPrefix(strings.Trim(strings.TrimPrefix(path, p.ModulePath), "/"), ".git") {
					return nil
				}
				fns, err := filepath.Glob(filepath.Join(path, "*.go"))
				if len(fns) == 0 || err != nil {
					return nil
				}
				name := filepath.Join(pkgName, strings.TrimPrefix(path, pkgPath))
				name = filepath.ToSlash(name)
				p.KnownPkgs = append(p.KnownPkgs, pkg{
					Name: name,
					Path: path,
				})
				p.KnownNamePkg[name] = &p.KnownPkgs[len(p.KnownPkgs)-1]
				p.KnownPathPkg[path] = &p.KnownPkgs[len(p.KnownPkgs)-1]
			}
			return nil
		}
		_ = filepath.Walk(pkgPath, walker)
	}
	return nil
}

func (p *parser) getPkgAst(pkgPath string) (map[string]*ast.Package, error) {
	if cache, ok := p.PkgPathAstPkgCache[pkgPath]; ok {
		return cache, nil
	}
	ignoreFileFilter := func(info os.FileInfo) bool {
		name := info.Name()
		return !info.IsDir() && !strings.HasPrefix(name, ".") && strings.HasSuffix(name, ".go") && !strings.HasSuffix(name, "_test.go")
	}
	astPackages, err := goparser.ParseDir(token.NewFileSet(), pkgPath, ignoreFileFilter, goparser.ParseComments)
	if err != nil {
		return nil, err
	}
	p.PkgPathAstPkgCache[pkgPath] = astPackages
	return astPackages, nil
}

func (p *parser) parseAPIs() error {
	err := p.parseImportStatements()
	if err != nil {
		return err
	}

	err = p.parseTypeSpecs()
	if err != nil {
		return err
	}

	return p.parsePaths()
}

func (p *parser) parseImportStatements() error {
	for i := range p.KnownPkgs {
		pkgPath := p.KnownPkgs[i].Path
		pkgName := p.KnownPkgs[i].Name

		astPkgs, err := p.getPkgAst(pkgPath)
		if err != nil {
			return fmt.Errorf("parseImportStatements: parse of %s package cause error: %s", pkgPath, err)
		}

		p.PkgNameImportedPkgAlias[pkgName] = map[string][]string{}
		for _, astPackage := range astPkgs {
			for _, astFile := range astPackage.Files {
				for _, astImport := range astFile.Imports {
					importedPkgName := strings.Trim(astImport.Path.Value, "\"")
					importedPkgAlias := ""

					if astImport.Name != nil && astImport.Name.Name != "." && astImport.Name.Name != "_" {
						importedPkgAlias = astImport.Name.String()
					} else {
						s := strings.Split(importedPkgName, "/")
						importedPkgAlias = s[len(s)-1]
					}

					exist := false
					for _, v := range p.PkgNameImportedPkgAlias[pkgName][importedPkgAlias] {
						if v == importedPkgName {
							exist = true
							break
						}
					}
					if !exist {
						p.PkgNameImportedPkgAlias[pkgName][importedPkgAlias] = append(p.PkgNameImportedPkgAlias[pkgName][importedPkgAlias], importedPkgName)
					}
				}
			}
		}
	}
	return nil
}

func (p *parser) parseTypeSpecs() error {
	for i := range p.KnownPkgs {
		pkgPath := p.KnownPkgs[i].Path
		pkgName := p.KnownPkgs[i].Name

		_, ok := p.TypeSpecs[pkgName]
		if !ok {
			p.TypeSpecs[pkgName] = map[string]*ast.TypeSpec{}
		}
		astPkgs, err := p.getPkgAst(pkgPath)
		if err != nil {
			return fmt.Errorf("parseTypeSpecs: parse of %s package cause error: %s", pkgPath, err)
		}
		for _, astPackage := range astPkgs {
			for _, astFile := range astPackage.Files {
				for _, astDeclaration := range astFile.Decls {
					if astGenDeclaration, ok := astDeclaration.(*ast.GenDecl); ok && astGenDeclaration.Tok == token.TYPE {
						// find type declaration
						p.findTypeDeclaration(pkgName, astGenDeclaration)
					} else if astFuncDeclaration, ok := astDeclaration.(*ast.FuncDecl); ok {
						// find type declaration in func, method
						p.findTypeDeclarationFunc(pkgName, astFuncDeclaration)
					}
				}
			}
		}
	}

	return nil
}

func (p *parser) findTypeDeclaration(pkgName string, astGenDeclaration *ast.GenDecl) {
	for _, astSpec := range astGenDeclaration.Specs {
		if typeSpec, ok := astSpec.(*ast.TypeSpec); ok {
			typeSpec.Doc = astGenDeclaration.Doc // assign the gendec Doc block to the typeSpec docblock
			p.TypeSpecs[pkgName][typeSpec.Name.String()] = typeSpec
		}
	}
}

func (p *parser) findTypeDeclarationFunc(pkgName string, astFuncDeclaration *ast.FuncDecl) {
	if astFuncDeclaration.Doc != nil && astFuncDeclaration.Doc.List != nil && astFuncDeclaration.Body != nil {
		funcName := astFuncDeclaration.Name.String()
		for _, astStmt := range astFuncDeclaration.Body.List {
			if astDeclStmt, ok := astStmt.(*ast.DeclStmt); ok {
				if astGenDeclaration, ok := astDeclStmt.Decl.(*ast.GenDecl); ok {
					for _, astSpec := range astGenDeclaration.Specs {
						if typeSpec, ok := astSpec.(*ast.TypeSpec); ok {
							// type in func
							if astFuncDeclaration.Recv == nil {
								p.TypeSpecs[pkgName][strings.Join([]string{funcName, typeSpec.Name.String()}, "@")] = typeSpec
								continue
							}
							// type in method
							var recvTypeName string
							switch astFuncDec := astFuncDeclaration.Recv.List[0].Type.(type) {
							case *ast.StarExpr:
								recvTypeName = fmt.Sprintf("%s", astFuncDec.X)
							case *ast.Ident:
								recvTypeName = astFuncDec.String()
							}
							p.TypeSpecs[pkgName][strings.Join([]string{recvTypeName, funcName, typeSpec.Name.String()}, "@")] = typeSpec
						}
					}
				}
			}
		}
	}
}

func (p *parser) parsePaths() error {
	for i := range p.KnownPkgs {
		pkgPath := p.KnownPkgs[i].Path
		pkgName := p.KnownPkgs[i].Name

		astPkgs, err := p.getPkgAst(pkgPath)
		if err != nil {
			return fmt.Errorf("parsePaths: parse of %s package cause error: %s", pkgPath, err)
		}
		for _, astPackage := range astPkgs {
			for _, astFile := range astPackage.Files {
				for _, astDeclaration := range astFile.Decls {
					if astFuncDeclaration, ok := astDeclaration.(*ast.FuncDecl); ok {
						if astFuncDeclaration.Doc != nil && astFuncDeclaration.Doc.List != nil {
							err = p.parseOperation(pkgPath, pkgName, astFuncDeclaration.Doc.List)
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func isHidden(astComments []*ast.Comment) bool {
	for _, astComment := range astComments {
		comment := strings.TrimSpace(strings.TrimLeft(astComment.Text, "/"))
		if comment == "" {
			// ignore empty lines
			continue
		}
		attribute := strings.Fields(comment)[0]
		if strings.EqualFold(attribute, types.AttributeHidden) {
			return true
		}
	}
	return false
}

func (p *parser) parseOperation(pkgPath, pkgName string, astComments []*ast.Comment) error {
	operation := &types.OperationObject{
		Responses: map[string]*types.ResponseObject{},
	}
	if !strings.HasPrefix(pkgPath, p.ModulePath) {
		return nil
	} else if p.HandlerPath != "" && !strings.HasPrefix(pkgPath, p.HandlerPath) {
		return nil
	}
	if isHidden(astComments) {
		return nil
	}
	for _, astComment := range astComments {
		comment := strings.TrimSpace(strings.TrimLeft(astComment.Text, "/"))
		if comment == "" {
			// ignore empty lines
			continue
		}
		attribute := strings.Fields(comment)[0]
		switch strings.ToLower(attribute) {
		case types.AttributeTitle:
			operation.Summary = strings.TrimSpace(comment[len(attribute):])
		case types.AttributeDescription:
			operation.Description = strings.TrimSpace(
				strings.Join([]string{operation.Description, strings.TrimSpace(comment[len(attribute):])}, " "),
			)
		case types.AttributeParam:
			if err := p.parseParamComment(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):])); err != nil {
				return err
			}
		case types.AttributeHeader:
			if err := p.parseResponseHeader(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):])); err != nil {
				return err
			}
		case types.AttributeSuccess, types.AttributeFailure:
			if err := p.parseResponseComment(pkgPath, pkgName, operation, strings.TrimSpace(comment[len(attribute):])); err != nil {
				return err
			}
		case types.AttributeID:
			id := strings.TrimSpace(comment[len(attribute):])
			if err := p.validateOperationID(id); err != nil {
				return err
			}
			operation.OperationID = id
		case types.AttributeExternalDoc:
			externalDocs, err := p.parseExternalDocComment(strings.TrimSpace(comment[len(attribute):]))
			if err != nil {
				return err
			}
			if externalDocs == nil {
				return fmt.Errorf("couldn't populate externalDocs")
			}

			operation.ExternalDocs = externalDocs
		case types.AttributeResource, types.AttributeTag:
			resource := strings.TrimSpace(comment[len(attribute):])
			if resource == "" {
				resource = "others"
			}
			if !util.IsInStringList(operation.Tags, resource) {
				operation.Tags = append(operation.Tags, resource)
			}
		case types.AttributeRoute, types.AttributeRouter:
			if err := p.parseRouteComment(operation, comment); err != nil {
				return err
			}
		case types.AttributeSecurity:
			security := strings.TrimSpace(comment[len(attribute):])
			matches := strings.Split(security, " ")

			operation.Security = append(operation.Security, map[string][]string{
				matches[0]: {},
			})
		}
	}
	return nil
}

func (p *parser) parseSecurityScheme(value string) {
	// {key} http {in} {name} {description}
	// {key} apiKey {in} {name} {description}
	// {key} openIdConnect {connect_url} {description}
	// {key} oauth2AuthCode {auth_url} {token_url}
	// {key} oauth2Implicit {auth_url}
	// {key} oauth2ResourceOwnerCredentials {token_url}
	// {key} oauth2ClientCredentials {token_url}
	fields := strings.Split(value, " ")

	var scheme *types.SecuritySchemeObject
	if strings.Contains(fields[1], "oauth2") {
		if oauthScheme, ok := p.OpenAPI.Components.SecuritySchemes[fields[0]]; ok {
			scheme = oauthScheme
		} else {
			scheme = &types.SecuritySchemeObject{
				Type:       "oauth2",
				OAuthFlows: &types.SecuritySchemeOauthObject{},
			}
		}
	}

	if scheme == nil {
		scheme = &types.SecuritySchemeObject{
			Type: fields[1],
		}
	}
	switch fields[1] {
	case "http":
		scheme.Scheme = fields[2]
		if scheme.Scheme == "bearer" {
			scheme.Description = strings.Join(fields[3:], " ")
		} else {
			scheme.Name = fields[3]
			scheme.Description = strings.Join(fields[4:], " ")
		}
	case "apiKey":
		scheme.In = fields[2]
		scheme.Name = fields[3]
		scheme.Description = strings.Join(fields[4:], " ")
	case "openIdConnect":
		scheme.OpenIDConnectURL = fields[2]
		scheme.Description = strings.Join(fields[3:], " ")
	case "oauth2AuthCode":
		scheme.OAuthFlows.AuthorizationCode = &types.SecuritySchemeOauthFlowObject{
			AuthorizationURL: fields[2],
			TokenURL:         fields[3],
			Scopes:           make(map[string]string),
		}
	case "oauth2Implicit":
		scheme.OAuthFlows.Implicit = &types.SecuritySchemeOauthFlowObject{
			AuthorizationURL: fields[2],
			Scopes:           make(map[string]string),
		}
	case "oauth2ResourceOwnerCredentials":
		scheme.OAuthFlows.ResourceOwnerPassword = &types.SecuritySchemeOauthFlowObject{
			TokenURL: fields[2],
			Scopes:   make(map[string]string),
		}
	case "oauth2ClientCredentials":
		scheme.OAuthFlows.ClientCredentials = &types.SecuritySchemeOauthFlowObject{
			TokenURL: fields[2],
			Scopes:   make(map[string]string),
		}
	}
	if p.OpenAPI.Components.SecuritySchemes == nil {
		p.OpenAPI.Components.SecuritySchemes = make(map[string]*types.SecuritySchemeObject)
	}
	p.OpenAPI.Components.SecuritySchemes[fields[0]] = scheme
}

func (p *parser) parseServerVariableComment(comment string, server types.ServerObject) (map[string]types.ServerVariableObject, error) {
	// {name} {default} {description} {enum1,enum2,...}
	re := regexp.MustCompile(`([-\w]+)[\s]+"([^"]+)"[\s]*(?:"([^"]+)"(?:[\s]+"([\w,^"]+)"|$))`)
	matches := re.FindStringSubmatch(comment)
	validSegments := 5
	if len(matches) != validSegments {
		return nil, fmt.Errorf(`parseServerVariableComment can not parse servervariable comment %s`, comment)
	}

	if !strings.Contains(server.URL, fmt.Sprintf(`{%s}`, matches[1])) {
		return server.Variables, nil
	}

	serverVar := types.ServerVariableObject{
		Enum:        nil,
		Default:     matches[2],
		Description: matches[3],
	}

	if matches[4] != "" {
		enums := strings.Split(matches[4], ",")
		serverVar.Enum = enums
	}

	server.Variables[matches[1]] = serverVar

	return server.Variables, nil
}

func (p *parser) parseExternalDocComment(comment string) (*types.ExternalDocumentationObject, error) {
	// {url}  {description}

	re := regexp.MustCompile(`([\w?&#/:.]+)\s+"([^"]+)"`)
	matches := re.FindStringSubmatch(comment)
	validSegments := 3
	if len(matches) != validSegments {
		return nil, fmt.Errorf("parseExternalDocComment can not parse externaldoc comment \"%s\"", comment)
	}
	extURL := matches[1]
	description := matches[2]

	return &types.ExternalDocumentationObject{
		Description: description,
		URL:         extURL,
	}, nil
}

func (p *parser) parseTagComment(comment string) (*types.TagObject, error) {
	// {name} {description} {externalDocURL} {externalDocDesc}

	re := regexp.MustCompile(`([-\w]+)\s+"([^"]+)"\s*(?:([\w?&#/:.]+)\s+"([^"]+)"|$)`)
	matches := re.FindStringSubmatch(comment)

	if len(matches) != 5 || matches[1] == "" || matches[2] == "" {
		return nil, fmt.Errorf(`parseTagComment can not parse tag comment %s`, comment)
	}

	tag := &types.TagObject{
		Name:         matches[1],
		Description:  matches[2],
		ExternalDocs: nil,
	}

	if matches[3] != "" && matches[4] != "" {
		tag.ExternalDocs = &types.ExternalDocumentationObject{
			Description: matches[4],
			URL:         matches[3],
		}
	}

	return tag, nil
}

func (p *parser) parseParamComment(pkgPath, pkgName string, operation *types.OperationObject, comment string) error {
	// {name}  {in}  {goType}  {required}  {description}
	// user    body  User      true        "Info of a user."
	// f       file  ignored   true        "Upload a file."
	re := regexp.MustCompile(`([-\w]+)[\s]+([\w]+)[\s]+([\w./\[\]]+)[\s]+([\w]+)[\s]+"([^"]+)"`)
	matches := re.FindStringSubmatch(comment)
	validSegments := 6
	if len(matches) != validSegments {
		return fmt.Errorf("parseParamComment can not parse param comment \"%s\"", comment)
	}
	name := matches[1]
	in := matches[2]

	re = regexp.MustCompile(`\[\w*]`)
	goType := re.ReplaceAllString(matches[3], "[]")

	required := false
	switch strings.ToLower(matches[4]) {
	case "true", types.KeywordRequired:
		required = true
	}
	description := matches[5]

	// `file`, `form`
	if ok := p.handleFileOrForm(name, in, operation, goType, description, required); ok {
		return nil
	}

	// `path`, `query`, `header`, `cookie`
	if in != types.InBody {
		if err := p.handleParam(name, in, operation, description, goType, required, pkgPath, pkgName); err != nil {
			return fmt.Errorf("unable to handle params: %v", err)
		}
		return nil
	}

	if operation.RequestBody == nil {
		operation.RequestBody = &types.RequestBodyObject{
			Content:  map[string]*types.MediaTypeObject{},
			Required: required,
		}
	}

	if strings.HasPrefix(goType, "[]") || strings.HasPrefix(goType, "map[]") || goType == types.GoTypeTime {
		schema, err := p.parseSchemaObject(pkgPath, pkgName, name, goType)
		if err != nil {
			return fmt.Errorf("parseResponseComment cannot parse goType: %s", goType)
		}
		operation.RequestBody.Content[types.ContentTypeJSON] = &types.MediaTypeObject{
			Schema: *schema,
		}
	} else {
		typeName, err := p.registerType(pkgPath, pkgName, matches[3])
		if err != nil {
			return err
		}
		if types.IsBasicGoType(typeName) {
			operation.RequestBody.Content[types.ContentTypeJSON] = &types.MediaTypeObject{
				Schema: types.SchemaObject{
					Type: "string",
				},
			}
		} else {
			operation.RequestBody.Content[types.ContentTypeJSON] = &types.MediaTypeObject{
				Schema: types.SchemaObject{
					Ref: util.AddSchemaRefLinkPrefix(typeName),
				},
			}
		}
	}

	return nil
}

func (p *parser) handleParam(
	name string,
	in string,
	operation *types.OperationObject,
	description string,
	goType string,
	required bool,
	pkgPath string,
	pkgName string) error {
	parameterObject := types.ParameterObject{
		Name:        name,
		In:          in,
		Description: description,
		Required:    required,
	}
	if in == types.InPath {
		parameterObject.Required = true
	}
	if goType == types.GoTypeTime {
		var err error
		parameterObject.Schema, err = p.parseSchemaObject(pkgPath, pkgName, name, goType)
		if err != nil {
			return fmt.Errorf("parseResponseComment cannot parse goType: %s", goType)
		}
		operation.Parameters = append(operation.Parameters, parameterObject)
	} else if types.IsGoTypeOASType(goType) {
		parameterObject.Schema = &types.SchemaObject{
			Type:   types.GoTypesOASTypes[goType],
			Format: types.GoTypesOASFormats[goType],
			//Description: description,
		}
		operation.Parameters = append(operation.Parameters, parameterObject)
	}
	return nil
}

func (p *parser) handleFileOrForm(name, in string, operation *types.OperationObject, goType, description string, required bool) bool {
	if in == types.InFile || in == types.InFiles || in == types.InForm {
		if operation.RequestBody == nil {
			operation.RequestBody = &types.RequestBodyObject{
				Content: map[string]*types.MediaTypeObject{
					types.ContentTypeForm: {
						Schema: types.SchemaObject{
							Type:       types.TypeObject,
							Properties: types.NewOrderedMap(),
						},
					},
				},
				Required: required,
			}
		}
		if in == types.InFile {
			operation.RequestBody.Content[types.ContentTypeForm].Schema.Properties.Set(name, &types.SchemaObject{
				Type:        "string",
				Format:      "binary",
				Description: description,
			})
		} else if in == types.InFiles {
			operation.RequestBody.Content[types.ContentTypeForm].Schema.Properties.Set(name, &types.SchemaObject{
				Type: types.TypeArray,
				Items: &types.SchemaObject{
					Type:   "string",
					Format: "binary",
				},
				Description: description,
			})
		} else if types.IsGoTypeOASType(goType) {
			operation.RequestBody.Content[types.ContentTypeForm].Schema.Properties.Set(name, &types.SchemaObject{
				Type:        types.GoTypesOASTypes[goType],
				Format:      types.GoTypesOASFormats[goType],
				Description: description,
			})
		}
		return true
	}
	return false
}

func (p *parser) parseResponseHeader(pkgPath, pkgName string, operation *types.OperationObject, comment string) error {
	// {status} {name} {jsonType} {goType} {description}
	// 201  x-next  object  string  "A link"
	minValidSegments := 4
	re := regexp.MustCompile(`(?P<status>[\w-]+)[\s]*(?P<name>[\w-]+)[\s]*(?P<jsonType>[\w{}]+)?[\s]+(?P<goType>[\w\-./\[\]]+)?[^"]*(?P<description>.*)?`)
	matches := re.FindStringSubmatch(comment)

	paramsMap := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i > 0 && i <= len(matches) {
			paramsMap[name] = matches[i]
		}
	}

	if len(matches) <= minValidSegments {
		return fmt.Errorf("parseResponseHeader can not parse response header \"%s\", matches: %v", comment, matches)
	}

	status := paramsMap["status"]
	if strings.EqualFold(status, "default") {
		_, err := strconv.Atoi(status)
		if err != nil {
			return fmt.Errorf("parseResponseHeader: http status must be int, but got %s", status)
		}
	}

	var responseObject *types.ResponseObject
	if _, ok := operation.Responses[status]; !ok {
		responseObject = &types.ResponseObject{
			Content: map[string]*types.MediaTypeObject{},
			Headers: make(map[string]*types.HeaderObject),
		}
	} else {
		responseObject = operation.Responses[status]
	}

	if responseObject.Headers == nil {
		responseObject.Headers = make(map[string]*types.HeaderObject)
	}

	if goTypeRaw := paramsMap["goType"]; goTypeRaw != "" {
		re = regexp.MustCompile(`\[\w*]`)
		goType := re.ReplaceAllString(goTypeRaw, "[]")
		if strings.HasPrefix(goType, "[]") || strings.HasPrefix(goType, "map[]") {
			schema, err := p.parseSchemaObject(pkgPath, pkgName, "", goType)
			if err != nil {
				return fmt.Errorf("parseResponseHeader: cannot parse goType: %s", goType)
			}
			responseObject.Headers[paramsMap["name"]] = &types.HeaderObject{
				Description: strings.Trim(paramsMap["description"], "\""),
				Schema:      schema,
			}
		} else {
			typeName, err := p.registerType(pkgPath, pkgName, matches[3])
			if err != nil {
				return err
			}
			if types.IsBasicGoType(typeName) {
				responseObject.Headers[paramsMap["name"]] = &types.HeaderObject{
					Description: strings.Trim(paramsMap["description"], "\""),
					Schema: &types.SchemaObject{
						Type: "string",
					},
				}
			} else {
				responseObject.Headers[paramsMap["name"]] = &types.HeaderObject{
					Description: strings.Trim(paramsMap["description"], "\""),
					Schema: &types.SchemaObject{
						Ref: util.AddSchemaRefLinkPrefix(typeName),
					},
				}
			}
		}
	}

	return nil
}

func (p *parser) parseResponseComment(pkgPath, pkgName string, operation *types.OperationObject, comment string) error {
	// {status}  {jsonType}  {goType}     {description}
	// 201       object      models.User  "User Model"
	// if 204 or something else without empty return payload
	// 204 "User Model"
	minValidSegments := 2
	re := regexp.MustCompile(`(?P<status>[\w]+)[\s]*(?P<jsonType>[\w{}]+)?[\s]+(?P<goType>[\w\-./\[\]]+)?[^"]*(?P<description>.*)?`)
	matches := re.FindStringSubmatch(comment)

	paramsMap := make(map[string]string)
	for i, name := range re.SubexpNames() {
		if i > 0 && i <= len(matches) {
			paramsMap[name] = matches[i]
		}
	}

	if len(matches) <= minValidSegments {
		return fmt.Errorf("parseResponseComment can not parse response comment \"%s\", matches: %v", comment, matches)
	}

	status := paramsMap["status"]
	if !strings.EqualFold(status, "default") {
		_, err := strconv.Atoi(status)
		if err != nil {
			return fmt.Errorf("parseResponseComment: http status must be int, but got %s", status)
		}
	}

	// ignore type if not set
	if jsonType := paramsMap["jsonType"]; jsonType != "" {
		switch jsonType {
		case types.TypeObject, types.TypeArray, "{object}", "{array}":
		default:
			return fmt.Errorf("parseResponseComment: invalid jsonType \"%s\"", paramsMap["jsonType"])
		}
	}

	responseObject := &types.ResponseObject{
		Content: map[string]*types.MediaTypeObject{},
	}
	responseObject.Description = strings.Trim(paramsMap["description"], "\"")

	if goTypeRaw := paramsMap["goType"]; goTypeRaw != "" {
		re = regexp.MustCompile(`\[\w*]`)
		goType := re.ReplaceAllString(goTypeRaw, "[]")
		if strings.HasPrefix(goType, "[]") || strings.HasPrefix(goType, "map[]") {
			schema, err := p.parseSchemaObject(pkgPath, pkgName, "", goType)
			if err != nil {
				return fmt.Errorf("parseResponseComment: cannot parse goType: %s", goType)
			}
			responseObject.Content[types.ContentTypeJSON] = &types.MediaTypeObject{
				Schema: *schema,
			}
		} else {
			typeName, err := p.registerType(pkgPath, pkgName, matches[3])
			if err != nil {
				return err
			}
			if types.IsBasicGoType(typeName) {
				responseObject.Content[types.ContentTypeText] = &types.MediaTypeObject{
					Schema: types.SchemaObject{
						Type: "string",
					},
				}
			} else {
				responseObject.Content[types.ContentTypeJSON] = &types.MediaTypeObject{
					Schema: types.SchemaObject{
						Ref: util.AddSchemaRefLinkPrefix(typeName),
					},
				}
			}
		}
	}
	operation.Responses[status] = responseObject

	return nil
}

func (p *parser) parseRouteComment(operation *types.OperationObject, comment string) error {
	sourceString := strings.TrimSpace(comment[len("@Router"):])
	validSegments := 3

	// /path [method]
	//goland:noinspection ALL
	re := regexp.MustCompile(`([\w./\-{}]+)[^\[]+\[([^\]]+)`)
	matches := re.FindStringSubmatch(sourceString)
	if len(matches) != validSegments {
		return fmt.Errorf(`can not parse router comment "%s", skipped`, comment)
	}

	_, ok := p.OpenAPI.Paths[matches[1]]
	if !ok {
		p.OpenAPI.Paths[matches[1]] = &types.PathItemObject{}
	}

	switch strings.ToUpper(matches[2]) {
	case http.MethodGet:
		p.OpenAPI.Paths[matches[1]].Get = operation
	case http.MethodPost:
		p.OpenAPI.Paths[matches[1]].Post = operation
	case http.MethodPatch:
		p.OpenAPI.Paths[matches[1]].Patch = operation
	case http.MethodPut:
		p.OpenAPI.Paths[matches[1]].Put = operation
	case http.MethodDelete:
		p.OpenAPI.Paths[matches[1]].Delete = operation
	case http.MethodOptions:
		p.OpenAPI.Paths[matches[1]].Options = operation
	case http.MethodHead:
		p.OpenAPI.Paths[matches[1]].Head = operation
	case http.MethodTrace:
		p.OpenAPI.Paths[matches[1]].Trace = operation
	}

	return nil
}

func (p *parser) registerType(pkgPath, pkgName, typeName string) (string, error) {
	var registerTypeName string

	if types.IsBasicGoType(typeName) {
		registerTypeName = typeName
	} else {
		var schemaObject *types.SchemaObject

		// see if we've already parsed this type
		if knownObj, ok := p.KnownIDSchema[util.GenSchemaObjectID(typeName)]; ok {
			schemaObject = knownObj
		} else {
			// if not, parse it now
			parsedObject, err := p.parseSchemaObject(pkgPath, pkgName, "", typeName)
			if err != nil {
				return "", err
			}
			schemaObject = parsedObject
		}
		registerTypeName = schemaObject.ID
	}
	return registerTypeName, nil
}

func (p *parser) parseSchemaObject(pkgPath, pkgName, fieldName, typeName string) (*types.SchemaObject, error) {
	var typeSpec *ast.TypeSpec
	var exist bool
	schemaObject := &types.SchemaObject{}
	var err error

	// handler basic and some specific typeName
	if strings.HasPrefix(typeName, "[]") {
		schemaObject.Type = types.TypeArray
		itemTypeName := typeName[2:]
		schemaObject.Items, err = p.parseSchemaObject(pkgPath, pkgName, fieldName, itemTypeName)
		if err != nil {
			return nil, err
		}
		schema, ok := p.KnownIDSchema[util.GenSchemaObjectID(itemTypeName)]
		if ok {
			schemaObject.Items = &types.SchemaObject{Ref: util.AddSchemaRefLinkPrefix(schema.ID)}
			return schemaObject, nil
		}
		return schemaObject, nil
	} else if strings.HasPrefix(typeName, "map[]") {
		schemaObject.Type = types.TypeObject
		itemTypeName := typeName[5:]
		schema, ok := p.KnownIDSchema[util.GenSchemaObjectID(itemTypeName)]
		if ok {
			schemaObject.Items = &types.SchemaObject{Ref: util.AddSchemaRefLinkPrefix(schema.ID)}
			return schemaObject, nil
		}
		schemaProperty, err := p.parseSchemaObject(pkgPath, pkgName, fieldName, itemTypeName)
		if err != nil {
			return nil, err
		}
		schemaObject.Properties = types.NewOrderedMap()
		if fieldName == "" {
			fieldName = types.DefaultFieldName
		}
		schemaObject.Properties.Set(fieldName, schemaProperty)
		return schemaObject, nil
	} else if typeName == types.GoTypeTime {
		schemaObject.Type = "string"
		schemaObject.Format = "date-time"
		return schemaObject, nil
	} else if strings.HasPrefix(typeName, "interface{}") {
		return schemaObject, nil
	} else if types.IsGoTypeOASType(typeName) {
		schemaObject.Type = types.GoTypesOASTypes[typeName]
		return schemaObject, nil
	}

	// handler other type
	typeNameParts := strings.Split(typeName, ".")
	if len(typeNameParts) == 1 && typeNameParts[0] != types.GoTypeIgnored {
		typeSpec, exist = p.getTypeSpec(pkgName, typeName)
		if !exist {
			for _, value := range p.KnownNamePkg {
				typeSpec, exist = p.getTypeSpec(value.Name, typeName)
				if exist {
					pkgPath = value.Path
					pkgName = value.Name
					break
				}
			}
			if !exist {
				log.Fatalf("Can not find definition of %s ast.TypeSpec. Current package %s", typeName, pkgName)
			}
		}
		schemaObject.PkgName = pkgName
		schemaObject.ID = util.GenSchemaObjectID(typeName)
		p.KnownIDSchema[schemaObject.ID] = schemaObject
		if typeSpec.Doc != nil {
			p.parseSchemaComments(typeSpec.Doc.List, p.KnownIDSchema[schemaObject.ID])
		}
	} else {
		guessPkgName := strings.Join(typeNameParts[:len(typeNameParts)-1], "/")
		guessPkgPath := ""
		for i := range p.KnownPkgs {
			if strings.Contains(p.KnownPkgs[i].Name, guessPkgName) {
				guessPkgName = p.KnownPkgs[i].Name
				guessPkgPath = p.KnownPkgs[i].Path
				break
			}
		}
		guessTypeName := typeNameParts[len(typeNameParts)-1]
		typeSpec, exist = p.getTypeSpec(guessPkgName, guessTypeName)
		if !exist {
			found := false
			for k := range p.PkgNameImportedPkgAlias[pkgName] {
				if k == guessPkgName && len(p.PkgNameImportedPkgAlias[pkgName][guessPkgName]) != 0 {
					found = true
					break
				}
			}
			if !found {
				return schemaObject, nil
			}
			guessPkgName = p.PkgNameImportedPkgAlias[pkgName][guessPkgName][0]
			guessPkgPath = ""
			for i := range p.KnownPkgs {
				if guessPkgName == p.KnownPkgs[i].Name {
					guessPkgPath = p.KnownPkgs[i].Path
					break
				}
			}

			typeSpec, exist = p.getTypeSpec(guessPkgName, guessTypeName)
			if !exist {
				return schemaObject, fmt.Errorf("can not find definition of guess %s ast.TypeSpec in package %s", guessTypeName, guessPkgName)
			}
			schemaObject.PkgName = guessPkgName
			schemaObject.ID = util.GenSchemaObjectID(guessTypeName)
			p.KnownIDSchema[schemaObject.ID] = schemaObject
			p.parseSchemaComments(typeSpec.Doc.List, p.KnownIDSchema[schemaObject.ID])
		} else {
			schemaObject.PkgName = guessPkgName
			schemaObject.ID = util.GenSchemaObjectID(guessTypeName)
			p.KnownIDSchema[schemaObject.ID] = schemaObject
			if typeSpec.Doc != nil {
				p.parseSchemaComments(typeSpec.Doc.List, p.KnownIDSchema[schemaObject.ID])
			}
		}
		pkgPath, pkgName = guessPkgPath, guessPkgName
	}

	switch t := typeSpec.Type.(type) {
	case *ast.Ident:
		_ = t
	case *ast.StructType:
		if err := p.handleStructType(schemaObject, t, pkgPath, pkgName); err != nil {
			return nil, err
		}
	case *ast.ArrayType:
		if err := p.handleArrayType(schemaObject, t, pkgPath, pkgName); err != nil {
			return nil, err
		}
	case *ast.MapType:
		if err := p.handleMapType(fieldName, schemaObject, t, pkgPath, pkgName); err != nil {
			return nil, err
		}
	}

	// register schema object in spec tree if it doesn't exist
	registerTypeName := schemaObject.ID
	_, ok := p.OpenAPI.Components.Schemas[util.ReplaceBackslash(registerTypeName)]
	if !ok {
		p.OpenAPI.Components.Schemas[util.ReplaceBackslash(registerTypeName)] = schemaObject
	}

	return schemaObject, nil
}

func (p *parser) handleStructType(schemaObject *types.SchemaObject, t *ast.StructType, pkgPath, pkgName string) error {
	schemaObject.Type = types.TypeObject
	if t.Fields != nil {
		if err := p.parseSchemaPropertiesFromStructFields(pkgPath, pkgName, schemaObject, t.Fields.List); err != nil {
			return err
		}
	}
	return nil
}

func (p *parser) handleArrayType(schemaObject *types.SchemaObject, t *ast.ArrayType, pkgPath, pkgName string) error {
	schemaObject.Type = types.TypeArray
	schemaObject.Items = &types.SchemaObject{}
	typeAsString := p.getTypeAsString(t.Elt)
	typeAsString = strings.TrimLeft(typeAsString, "*")
	if !types.IsBasicGoType(typeAsString) {
		schemaItemsSchemaObjectID, err := p.registerType(pkgPath, pkgName, typeAsString)
		if err != nil {
			return fmt.Errorf("parseSchemaObject parse array items err: %v", err)
		}
		schemaObject.Items.Ref = util.AddSchemaRefLinkPrefix(schemaItemsSchemaObjectID)
	} else if types.IsGoTypeOASType(typeAsString) {
		schemaObject.Items.Type = types.GoTypesOASTypes[typeAsString]
	}
	return nil
}

func (p *parser) handleMapType(fieldName string, schemaObject *types.SchemaObject, t *ast.MapType, pkgPath, pkgName string) error {
	schemaObject.Type = types.TypeObject
	schemaObject.Properties = types.NewOrderedMap()
	propertySchema := &types.SchemaObject{}
	if fieldName == "" {
		fieldName = types.DefaultFieldName
	}
	schemaObject.Properties.Set(fieldName, propertySchema)
	typeAsString := p.getTypeAsString(t.Value)
	typeAsString = strings.TrimLeft(typeAsString, "*")
	if !types.IsBasicGoType(typeAsString) {
		schemaItemsSchemaObjectID, err := p.registerType(pkgPath, pkgName, typeAsString)
		if err != nil {
			return fmt.Errorf("parseSchemaObject parse array items err: %v", err)
		}
		propertySchema.Ref = util.AddSchemaRefLinkPrefix(schemaItemsSchemaObjectID)
	} else if types.IsGoTypeOASType(typeAsString) {
		propertySchema.Type = types.GoTypesOASTypes[typeAsString]
	}
	return nil
}

func (p *parser) getTypeSpec(pkgName, typeName string) (*ast.TypeSpec, bool) {
	pkgTypeSpecs, exist := p.TypeSpecs[pkgName]
	if !exist {
		return nil, false
	}
	astTypeSpec, exist := pkgTypeSpecs[typeName]
	if !exist {
		return nil, false
	}
	return astTypeSpec, true
}

func (p *parser) parseSchemaPropertiesFromStructFields(
	pkgPath,
	pkgName string,
	structSchema *types.SchemaObject,
	astFields []*ast.Field) error {
	// TODO this method is too complex
	// TODO errors are not bubbled up
	if astFields == nil {
		return nil
	}
	var err error
	structSchema.Properties = types.NewOrderedMap()
	if structSchema.DisabledFieldNames == nil {
		structSchema.DisabledFieldNames = map[string]struct{}{}
	}
astFieldsLoop:
	for _, astField := range astFields {
		if len(astField.Names) == 0 {
			continue
		}
		fieldSchema := &types.SchemaObject{}
		typeAsString := p.getTypeAsString(astField.Type)
		typeAsString = strings.TrimLeft(typeAsString, "*")
		if strings.HasPrefix(typeAsString, "[]") {
			fieldSchema, err = p.parseSchemaObject(pkgPath, pkgName, "", typeAsString)
			if err != nil {
				return fmt.Errorf("could not parse type %s as array: %v", typeAsString, err)
			}
		} else if strings.HasPrefix(typeAsString, "map[]") {
			fieldSchema, err = p.parseSchemaObject(pkgPath, pkgName, "", typeAsString)
			if err != nil {
				return fmt.Errorf("could not parse type %s as map: %v", typeAsString, err)
			}
		} else if typeAsString == types.GoTypeTime {
			fieldSchema, err = p.parseSchemaObject(pkgPath, pkgName, "", typeAsString)
			if err != nil {
				return fmt.Errorf("could not parse type %s as time.Time: %v", typeAsString, err)
			}
		} else if strings.HasPrefix(typeAsString, "interface{}") {
			fieldSchema, err = p.parseSchemaObject(pkgPath, pkgName, "", typeAsString)
			if err != nil {
				return fmt.Errorf("could not parse type %s as interface{}: %v", typeAsString, err)
			}
		} else if !types.IsBasicGoType(typeAsString) {
			fieldSchemaSchemeObjectID, err := p.registerType(pkgPath, pkgName, typeAsString)
			if err != nil {
				return fmt.Errorf("could not register type %s: %v", typeAsString, err)
			}
			fieldSchema.ID = fieldSchemaSchemeObjectID
			schema, ok := p.KnownIDSchema[fieldSchemaSchemeObjectID]
			if ok {
				fieldSchema.Type = schema.Type
				if schema.Items != nil {
					fieldSchema.Items = schema.Items
				}
			}
			fieldSchema.Ref = util.AddSchemaRefLinkPrefix(fieldSchemaSchemeObjectID)
			if fieldSchema.Type != "" {
				fieldSchema.Type = ""
			}
		} else if types.IsGoTypeOASType(typeAsString) {
			fieldSchema.Type = types.GoTypesOASTypes[typeAsString]
		}

		name := astField.Names[0].Name
		fieldSchema.FieldName = name
		_, disabled := structSchema.DisabledFieldNames[name]
		if disabled {
			continue
		}

		if astField.Tag != nil {
			astFieldTag := reflect.StructTag(strings.Trim(astField.Tag.Value, "`"))
			tagText := ""

			if tag := astFieldTag.Get("goas"); tag != "" {
				tagText = tag
			}
			tagValues := strings.Split(tagText, ",")
			for _, v := range tagValues {
				if v == "-" {
					structSchema.DisabledFieldNames[name] = struct{}{}
					fieldSchema.Deprecated = true
					continue astFieldsLoop
				}
			}

			if tag := astFieldTag.Get("json"); tag != "" {
				tagText = tag
			}
			tagValues = strings.Split(tagText, ",")
			isRequired := false
			for _, v := range tagValues {
				if v == "-" {
					structSchema.DisabledFieldNames[name] = struct{}{}
					fieldSchema.Deprecated = true
					continue astFieldsLoop
				} else if v == types.KeywordRequired {
					isRequired = true
				} else if v != "" && v != types.KeywordRequired && v != "omitempty" {
					name = v
				}
			}

			if err := p.parseFieldTags(name, astFieldTag, structSchema, fieldSchema, isRequired); err != nil {
				return err
			}
		}
		structSchema.Properties.Set(name, fieldSchema)
	}

	return nil
}

func (p *parser) parseFieldTags(
	name string,
	astFieldTag reflect.StructTag,
	structSchema,
	fieldSchema *types.SchemaObject,
	isRequired bool) error {
	if err := p.handleExample(astFieldTag, fieldSchema); err != nil {
		return err
	}

	if _, ok := astFieldTag.Lookup("required"); ok || isRequired {
		structSchema.Required = append(structSchema.Required, name)
	}

	if desc := astFieldTag.Get("description"); desc != "" {
		fieldSchema.Description = desc
	}

	if err := p.handleMultipleOf(astFieldTag, fieldSchema); err != nil {
		return err
	}

	if err := p.handleRange(astFieldTag, fieldSchema); err != nil {
		return err
	}

	if pattern := astFieldTag.Get("pattern"); pattern != "" {
		fieldSchema.Pattern = pattern
	}

	p.handleLengthMinMax(astFieldTag, fieldSchema)

	if fieldSchema.Type == types.TypeArray {
		p.handleItemMinMax(astFieldTag, fieldSchema)

		if uniqueItems := astFieldTag.Get("uniqueItems"); uniqueItems != "" {
			fieldSchema.UniqueItems, _ = strconv.ParseBool(uniqueItems)
		}
	}

	if fieldSchema.Type == types.TypeObject {
		p.handlePropertyMinMax(astFieldTag, fieldSchema)
	}

	p.handleEnumTag(astFieldTag, fieldSchema)

	if err := p.handleAllOfTag(astFieldTag, fieldSchema); err != nil {
		return err
	}

	if err := p.handleOneOfTag(astFieldTag, fieldSchema); err != nil {
		return err
	}

	if err := p.handleAnyOfTag(astFieldTag, fieldSchema); err != nil {
		return err
	}
	return nil
}

func (p *parser) handleExample(astFieldTag reflect.StructTag, fieldSchema *types.SchemaObject) error {
	if tag := astFieldTag.Get("example"); tag != "" {
		switch fieldSchema.Type {
		case types.TypeBoolean:
			fieldSchema.Example, _ = strconv.ParseBool(tag)
		case types.TypeInteger:
			fieldSchema.Example, _ = strconv.Atoi(tag)
		case types.TypeNumber:
			fieldSchema.Example, _ = strconv.ParseFloat(tag, 64)
		case types.TypeArray:
			b, err := json.RawMessage(tag).MarshalJSON()
			if err != nil {
				fieldSchema.Example = types.MessageInvalidExample
			} else {
				var sliceOfInterface []interface{}
				err := json.Unmarshal(b, &sliceOfInterface)
				if err != nil {
					fieldSchema.Example = types.MessageInvalidExample
				} else {
					fieldSchema.Example = sliceOfInterface
				}
			}
		case types.TypeObject:
			b, err := json.RawMessage(tag).MarshalJSON()
			if err != nil {
				fieldSchema.Example = types.MessageInvalidExample
			} else {
				mapOfInterface := map[string]interface{}{}
				err := json.Unmarshal(b, &mapOfInterface)
				if err != nil {
					fieldSchema.Example = types.MessageInvalidExample
				} else {
					fieldSchema.Example = mapOfInterface
				}
			}
		default:
			fieldSchema.Example = tag
		}

		if fieldSchema.Example != nil && fieldSchema.Ref != "" {
			fieldSchema.Ref = ""
		}
	}
	return nil
}

func (p *parser) handleMultipleOf(astFieldTag reflect.StructTag, fieldSchema *types.SchemaObject) error {
	if multipleOf := astFieldTag.Get("multipleOf"); multipleOf != "" {
		switch fieldSchema.Type {
		case types.TypeInteger:
			fieldSchema.MultipleOf, _ = strconv.Atoi(multipleOf)
		case types.TypeNumber:
			fieldSchema.MultipleOf, _ = strconv.ParseFloat(multipleOf, 64)
		default:
			return fmt.Errorf(`unable to parse %s value: %v`, "multipleOf", multipleOf)
		}
	}
	return nil
}

func (p *parser) handleRange(astFieldTag reflect.StructTag, fieldSchema *types.SchemaObject) error {
	if min := astFieldTag.Get("minimum"); min != "" {
		switch fieldSchema.Type {
		case types.TypeInteger:
			fieldSchema.Minimum, _ = strconv.Atoi(min)
		case types.TypeNumber:
			fieldSchema.Minimum, _ = strconv.ParseFloat(min, 64)
		default:
			return fmt.Errorf("unable to parse %s value: %v", "minimum", min)
		}
	}

	if max := astFieldTag.Get("maximum"); max != "" {
		switch fieldSchema.Type {
		case types.TypeInteger:
			fieldSchema.Maximum, _ = strconv.Atoi(max)
		case types.TypeNumber:
			fieldSchema.Maximum, _ = strconv.ParseFloat(max, 64)
		default:
			return fmt.Errorf("unable to parse %s value: %v", "maximum", max)
		}
	}

	if exclusiveMin := astFieldTag.Get("exclusiveMinimum"); exclusiveMin != "" {
		fieldSchema.ExclusiveMinimum, _ = strconv.ParseBool(exclusiveMin)
	}

	if exclusiveMax := astFieldTag.Get("exclusiveMaximum"); exclusiveMax != "" {
		fieldSchema.ExclusiveMaximum, _ = strconv.ParseBool(exclusiveMax)
	}

	return nil
}

func (p *parser) handleLengthMinMax(astFieldTag reflect.StructTag, fieldSchema *types.SchemaObject) {
	if minLength := astFieldTag.Get("minLength"); minLength != "" {
		fieldSchema.MinLength, _ = strconv.Atoi(minLength)
	}

	if maxLength := astFieldTag.Get("maxLength"); maxLength != "" {
		fieldSchema.MaxLength, _ = strconv.Atoi(maxLength)
	}
}

func (p *parser) handleItemMinMax(astFieldTag reflect.StructTag, fieldSchema *types.SchemaObject) {
	if minItems := astFieldTag.Get("minItems"); minItems != "" {
		fieldSchema.MinItems, _ = strconv.Atoi(minItems)
	}

	if maxItems := astFieldTag.Get("maxItems"); maxItems != "" {
		fieldSchema.MaxItems, _ = strconv.Atoi(maxItems)
	}
}

func (p *parser) handlePropertyMinMax(astFieldTag reflect.StructTag, fieldSchema *types.SchemaObject) {
	if minProperties := astFieldTag.Get("minProperties"); minProperties != "" {
		fieldSchema.MinProperties, _ = strconv.Atoi(minProperties)
	}

	if maxProperties := astFieldTag.Get("maxProperties"); maxProperties != "" {
		fieldSchema.MaxProperties, _ = strconv.Atoi(maxProperties)
	}
}

func (p *parser) handleEnumTag(astFieldTag reflect.StructTag, fieldSchema *types.SchemaObject) {
	if enum := astFieldTag.Get("enum"); enum != "" {
		enums := strings.Split(strings.TrimSpace(enum), ",")
		fieldSchema.Enum = enums
	}
}

func (p *parser) handleAllOfTag(astFieldTag reflect.StructTag, fieldSchema *types.SchemaObject) error {
	if allOf := astFieldTag.Get("allOf"); allOf != "" {
		typeNames := strings.Split(strings.TrimSpace(allOf), ",")
		for _, typeName := range typeNames {
			schemaObject, err := p.parseSchemaObject("", "", "", typeName)
			if err != nil {
				return fmt.Errorf("unable to find object with name %s: %v", typeName, err)
			}
			fieldSchema.AllOf = append(fieldSchema.AllOf, &types.ReferenceObject{
				Ref: util.AddSchemaRefLinkPrefix(schemaObject.ID),
			})
		}
	}
	return nil
}

func (p *parser) handleOneOfTag(astFieldTag reflect.StructTag, fieldSchema *types.SchemaObject) error {
	if oneOf := astFieldTag.Get("oneOf"); oneOf != "" {
		// get discriminator if available
		if discriminator := astFieldTag.Get("discriminator"); discriminator != "" {
			fieldSchema.Discriminator = &types.Discriminator{PropertyName: discriminator}
		}

		typeNames := strings.Split(strings.TrimSpace(oneOf), ",")
		for _, typeName := range typeNames {
			schemaObject, err := p.parseSchemaObject("", "", "", typeName)
			if err != nil {
				return fmt.Errorf("unable to find object with name %s: %v", typeName, err)
			}

			if fieldSchema.Discriminator != nil && schemaObject.Properties != nil {
				if _, ok := schemaObject.Properties.Get(fieldSchema.Discriminator.PropertyName); !ok {
					return fmt.Errorf("unable to find discriminator field: %s, in schema: %s", fieldSchema.Discriminator.PropertyName, schemaObject.ID)
				}
			}

			fieldSchema.OneOf = append(fieldSchema.OneOf, &types.ReferenceObject{
				Ref: util.AddSchemaRefLinkPrefix(schemaObject.ID),
			})
		}
	}
	return nil
}

func (p *parser) handleAnyOfTag(astFieldTag reflect.StructTag, fieldSchema *types.SchemaObject) error {
	if anyOf := astFieldTag.Get("anyOf"); anyOf != "" {
		typeNames := strings.Split(strings.TrimSpace(anyOf), ",")
		for _, typeName := range typeNames {
			schemaObject, err := p.parseSchemaObject("", "", "", typeName)
			if err != nil {
				return fmt.Errorf("unable to find object with name %s: %v", typeName, err)
			}
			fieldSchema.AnyOf = append(fieldSchema.AnyOf, &types.ReferenceObject{
				Ref: util.AddSchemaRefLinkPrefix(schemaObject.ID),
			})
		}
	}
	return nil
}

func (p *parser) getTypeAsString(fieldType interface{}) string {
	astArrayType, ok := fieldType.(*ast.ArrayType)
	if ok {
		return fmt.Sprintf("[]%v", p.getTypeAsString(astArrayType.Elt))
	}

	astMapType, ok := fieldType.(*ast.MapType)
	if ok {
		return fmt.Sprintf("map[]%v", p.getTypeAsString(astMapType.Value))
	}

	_, ok = fieldType.(*ast.InterfaceType)
	if ok {
		return "interface{}"
	}

	astStarExpr, ok := fieldType.(*ast.StarExpr)
	if ok {
		return fmt.Sprintf("%v", p.getTypeAsString(astStarExpr.X))
	}

	astSelectorExpr, ok := fieldType.(*ast.SelectorExpr)
	if ok {
		packageNameIdent, _ := astSelectorExpr.X.(*ast.Ident)
		return packageNameIdent.Name + "." + astSelectorExpr.Sel.Name
	}

	return fmt.Sprint(fieldType)
}

func (p *parser) validateOperationID(id string) error {
	for _, oid := range p.KnownOperationIDs {
		if oid == id {
			return fmt.Errorf("operationID %s is already in use", id)
		}
	}

	p.KnownOperationIDs = append(p.KnownOperationIDs, id)
	return nil
}
