package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAddSchemaRefLinkPrefix(t *testing.T) {
	assert.Equal(t, "#/components/schemas/Object", AddSchemaRefLinkPrefix("Object"))
	assert.Equal(t, "#/components/schemas/Object", AddSchemaRefLinkPrefix("#/components/schemas/Object"))
}

func TestGenSchemaObjectID(t *testing.T) {
	tests := map[string]struct {
		typeName string
		want     string
	}{
		"identified type name": {
			typeName: "Object",
			want:     "Object",
		},
		"prefixed type name": {
			typeName: "test.Object",
			want:     "Object",
		},
		"repo prefixed type name": {
			typeName: "user.goas.Object",
			want:     "Object",
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, GenSchemaObjectID(tc.typeName))
		})
	}
}
