package util

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/leonelquinteros/gotext"

	"golang.org/x/mod/modfile"
)

type ModulePath string

func (m ModulePath) Get() (string, error) {
	path := string(m)
	if path == "" {
		path, _ = os.Getwd()
	}

	path = fmt.Sprintf("%s/go.mod", path)

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return modfile.ModulePath(data), nil
}

func (m ModulePath) CheckPathExists() (string, error) {
	modulePath := string(m)
	modulePath, _ = filepath.Abs(modulePath)
	moduleInfo, err := os.Stat(modulePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", err
		}
		return "", fmt.Errorf(gotext.Get("error.io.stat-error", modulePath, err))
	}

	if !moduleInfo.IsDir() {
		return "", fmt.Errorf(gotext.Get("error.io.expected-directory", "module path"))
	}

	return modulePath, nil
}

func (m ModulePath) CheckGoModExists() (string, os.FileInfo, error) {
	modulePath := string(m)
	goModFilePath := filepath.Join(modulePath, "go.mod")
	goModFileInfo, err := os.Stat(goModFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, err
		}
		return "", nil, fmt.Errorf(gotext.Get("error.io.stat-error", goModFilePath, err))
	}
	if goModFileInfo.IsDir() {
		return "", nil, fmt.Errorf(gotext.Get("error.io.expected-file", goModFilePath))
	}

	return goModFilePath, goModFileInfo, nil
}

func (m ModulePath) CheckMainFilePathExists(mainFilePath string) (string, error) {
	modulePath := string(m)

	if mainFilePath == "" {
		fns, err := filepath.Glob(filepath.Join(modulePath, "*.go"))
		if err != nil {
			return "", err
		}
		for _, fn := range fns {
			ok, err := IsMainFile(fn)
			if err != nil {
				return "", err
			}
			if ok {
				mainFilePath = fn
				break
			}
		}
	} else {
		mainFileInfo, err := os.Stat(mainFilePath)
		if err != nil {
			if os.IsNotExist(err) {
				return "", err
			}
			return "", fmt.Errorf(gotext.Get("error.io.stat-error", mainFilePath, err))
		}
		if mainFileInfo.IsDir() {
			return "", fmt.Errorf(gotext.Get("error.io.expected-file", mainFilePath))
		}
	}

	return mainFilePath, nil
}

// IsMainFile checks if the main.go file is in the nominated path
func IsMainFile(path string) (bool, error) {
	f, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer f.Close()

	var isMainPackage, hasMainFunc bool

	bs := bufio.NewScanner(f)
	for bs.Scan() {
		l := bs.Text()
		if !isMainPackage && strings.HasPrefix(l, "package main") {
			isMainPackage = true
		}
		if !hasMainFunc && strings.HasPrefix(l, "func main()") {
			hasMainFunc = true
		}
		if isMainPackage && hasMainFunc {
			break
		}
	}
	if bs.Err() != nil {
		return false, bs.Err()
	}

	return isMainPackage && hasMainFunc, nil
}
