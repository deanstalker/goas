package util

import "strings"

// AddSchemaRefLinkPrefix for prefixing an object id with the oas path
func AddSchemaRefLinkPrefix(name string) string {
	if strings.HasPrefix(name, "#/components/schemas/") {
		return ReplaceBackslash(name)
	}
	return ReplaceBackslash("#/components/schemas/" + name)
}

// GenSchemaObjectID for generating a schema object id
func GenSchemaObjectID(typeName string) string {
	typeNameParts := strings.Split(typeName, ".")
	return typeNameParts[len(typeNameParts)-1]
}

// ReplaceBackslash with forward slash
func ReplaceBackslash(origin string) string {
	return strings.ReplaceAll(origin, "\\", "/")
}
