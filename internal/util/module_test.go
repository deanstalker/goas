package util

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetModulePath test
func TestGetModulePath(t *testing.T) {
	path, err := GetModulePath("../../go.mod")
	assert.NoError(t, err)
	assert.Contains(t, path, "github.com")

	path, err = GetModulePath("")
	assert.Error(t, err)
	assert.Equal(t, "", path)
}

// TestGetModulePath test
func TestIsMainFile(t *testing.T) {
	dir, _ := os.Getwd()
	ok, err := IsMainFile(fmt.Sprintf("%s/../..//main.go", dir))
	if err != nil {
		assert.False(t, ok)
		assert.Error(t, err)
	}
	assert.True(t, ok)
	assert.NoError(t, err)

	ok, err = IsMainFile(fmt.Sprintf("%s/oas.go", dir))
	if err != nil {
		assert.False(t, ok)
		assert.Error(t, err)
	}
	assert.False(t, ok)
	assert.NoError(t, err)
}
