package util

import (
	"fmt"
	"io/ioutil"
	"os"

	"golang.org/x/mod/modfile"
)

func GetModulePath(path string) (string, error) {
	if path == "" {
		path, _ = os.Getwd()
	}

	data, err := ioutil.ReadFile(fmt.Sprintf("%s/go.mod", path))
	if err != nil {
		return "", err
	}

	return modfile.ModulePath(data), nil
}
