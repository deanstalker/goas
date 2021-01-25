package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetModulePath(t *testing.T) {
	path, err := GetModulePath("../../")
	assert.NoError(t, err)
	assert.Contains(t, path, "github.com")

	path, err = GetModulePath("")
	assert.Error(t, err)
	assert.Equal(t, "", path)
}
