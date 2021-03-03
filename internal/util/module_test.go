package util

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"
	"testing"

	"github.com/leonelquinteros/gotext"

	"github.com/stretchr/testify/assert"
)

// TestModulePath_Get tests get abs path from modulepath
func TestModulePath_Get(t *testing.T) {
	gotext.Configure("../../locales", "en", "default")
	modulePath := ModulePath("../../")
	path, err := modulePath.Get()
	assert.NoError(t, err)
	assert.Contains(t, path, "github.com")

	modulePath = ""
	path, err = modulePath.Get()
	assert.Error(t, err)
	assert.Equal(t, "", path)
}

// TestModulePath_CheckPathExists check if module path exists and is valid
func TestModulePath_CheckPathExists(t *testing.T) {
	gotext.Configure("../../locales", "en", "default")
	path, _ := os.Getwd()
	path, _ = filepath.Abs(fmt.Sprintf("%s/../../", path))
	tests := map[string]struct {
		modulePath string
		want       string
		wantErr    error
	}{
		"valid module path": {
			path,
			path,
			nil,
		},
		"module path is not a directory": {
			fmt.Sprintf("%s/main.go", path),
			"",
			fmt.Errorf(gotext.Get("error.io.expected-directory", "module path")),
		},
		"module path does not exist": {
			fmt.Sprintf("%s/does_not_exist/", path),
			"",
			&os.PathError{
				Op:   "stat",
				Path: fmt.Sprintf("%s/does_not_exist", path),
				Err:  syscall.ENOENT,
			},
		},
		"module path is missing a go.mod file": {
			fmt.Sprintf("%s/internal/", path),
			fmt.Sprintf("%s/internal", path),
			nil,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			modulePath := ModulePath(tc.modulePath)
			path, err := modulePath.CheckPathExists()
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, path)
		})
	}
}

func TestModulePath_CheckGoModExists(t *testing.T) {
	gotext.Configure("../../locales", "en", "default")
	path, _ := os.Getwd()
	path, _ = filepath.Abs(fmt.Sprintf("%s/../../", path))
	tests := map[string]struct {
		modulePath string
		want       string
		wantErr    error
	}{
		"valid module path": {
			path,
			fmt.Sprintf("%s/go.mod", path),
			nil,
		},
		"module path is not a directory": {
			fmt.Sprintf("%s/main.go", path),
			"",
			fmt.Errorf("cannot get information of %s: stat %s: not a directory",
				fmt.Sprintf("%s/main.go/go.mod", path),
				fmt.Sprintf("%s/main.go/go.mod", path)),
		},
		"module path does not exist": {
			fmt.Sprintf("%s/does_not_exist/", path),
			"",
			&os.PathError{
				Op:   "stat",
				Path: fmt.Sprintf("%s/does_not_exist/go.mod", path),
				Err:  syscall.ENOENT,
			},
		},
		"module path is missing a go.mod file": {
			fmt.Sprintf("%s/internal/", path),
			"",
			&os.PathError{
				Op:   "stat",
				Path: fmt.Sprintf("%s/internal/go.mod", path),
				Err:  syscall.ENOENT,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			modulePath := ModulePath(tc.modulePath)
			path, _, err := modulePath.CheckGoModExists()
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, path)
		})
	}
}

func TestModulePath_CheckMainFilePathExists(t *testing.T) {
	gotext.Configure("../../locales", "en", "default")
	path, _ := os.Getwd()
	path, _ = filepath.Abs(fmt.Sprintf("%s/../../", path))
	tests := map[string]struct {
		modulePath   string
		mainFilePath string
		want         string
		wantErr      error
	}{
		"main file path not supplied": {
			path,
			"",
			fmt.Sprintf("%s/main.go", path),
			nil,
		},
		"invalid main file path supplied": {
			path,
			fmt.Sprintf("%s/internal/main.go", path),
			"",
			&os.PathError{
				Op:   "stat",
				Path: fmt.Sprintf("%s/internal/main.go", path),
				Err:  syscall.ENOENT,
			},
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			modulePath := ModulePath(tc.modulePath)
			path, err := modulePath.CheckMainFilePathExists(tc.mainFilePath)
			assert.Equal(t, tc.wantErr, err)
			assert.Equal(t, tc.want, path)
		})
	}
}

// TestGetModulePath test
func TestIsMainFile(t *testing.T) {
	gotext.Configure("../../locales", "en", "default")
	dir, _ := os.Getwd()
	ok, err := IsMainFile(fmt.Sprintf("%s/../../main.go", dir))
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
