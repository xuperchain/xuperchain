// Copyright 2014 dong<ddliuhb@gmail.com>.
// Licensed under the MIT license.
//
// Motto - Modular Javascript environment.
package motto

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

func isDir(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return fi.IsDir(), nil
}

func isFile(path string) (bool, error) {
	fi, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}

	return !fi.IsDir(), nil
}

type packageInfo struct {
	Main string `json:"main"`
}

func parsePackageEntryPoint(path string) (string, error) {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return "", err
	}

	var info packageInfo
	err = json.Unmarshal(bytes, &info)
	if err != nil {
		return "", err
	}

	return info.Main, nil
}

// Throw a javascript error, see https://github.com/robertkrimen/otto/issues/17
func jsException(vm *Motto, errorType, msg string) {
	value, _ := vm.Call("new "+errorType, nil, msg)
	panic(value)
}
