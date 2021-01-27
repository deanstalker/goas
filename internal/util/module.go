package util

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// GetModulePath fetches the path to the go.mod file
func GetModulePath(path string) (string, error) {
	if path == "" {
		path, _ = os.Getwd()
	}

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	return modfile.ModulePath(data), nil
}

// CheckModulePathExists ensure the path to the go.mod file exists
func CheckModulePathExists(modulePath string) (string, error) {
	// check modulePath is exist
	modulePath, _ = filepath.Abs(modulePath)
	moduleInfo, err := os.Stat(modulePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", err
		}
		return "", fmt.Errorf("cannot get information of %s: %s", modulePath, err)
	}
	if !moduleInfo.IsDir() {
		return "", fmt.Errorf("modulePath should be a directory")
	}

	return modulePath, nil
}

// CheckGoModExists will check if the go.mod file exists
func CheckGoModExists(modulePath string) (string, os.FileInfo, error) {
	goModFilePath := filepath.Join(modulePath, "go.mod")
	goModFileInfo, err := os.Stat(goModFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil, err
		}
		return "", nil, fmt.Errorf("cannot get information of %s: %s", goModFilePath, err)
	}
	if goModFileInfo.IsDir() {
		return "", nil, fmt.Errorf("%s should be a file", goModFilePath)
	}

	return goModFilePath, goModFileInfo, nil
}

// CheckMainFilePathExists will check to see if the main.go file exists in the main file path
func CheckMainFilePathExists(mainFilePath, modulePath string) (string, error) {
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
			return "", fmt.Errorf("cannot get information of %s: %s", mainFilePath, err)
		}
		if mainFileInfo.IsDir() {
			return "", fmt.Errorf("mainFilePath should not be a directory")
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
