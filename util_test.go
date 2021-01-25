package main

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsMainFile(t *testing.T) {
	dir, _ := os.Getwd()
	assert.True(t, IsMainFile(fmt.Sprintf("%s/main.go", dir)))
	assert.False(t, IsMainFile(fmt.Sprintf("%s/oas.go", dir)))
}
