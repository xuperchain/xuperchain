// Copyright 2014 dong<ddliuhb@gmail.com>.
// Licensed under the MIT license.
//
// Motto - Modular Javascript environment.
package motto

import (
	"errors"
	"io/ioutil"
	"path/filepath"
	"strings"

	"github.com/robertkrimen/otto"
)

// ModuleLoader is declared to load a module.
type ModuleLoader func(*Motto) (otto.Value, error)

// Create module loader from javascript source code.
//
// When the loader is called, the javascript source is executed in Motto.
//
// "pwd" indicates current working directory, which might be used to search for
// modules.
func CreateLoaderFromSource(source, filename string) ModuleLoader {
	pwd := filepath.Dir(filename)
	return func(vm *Motto) (otto.Value, error) {
		// Wraps the source to create a module environment
		source = "(function(module) {var require = module.require;var exports = module.exports;var __dirname = module.__dirname;" + source + "\n})"

		// Provide the "require" method in the module scope.
		jsRequire := func(call otto.FunctionCall) otto.Value {
			jsModuleName := call.Argument(0).String()

			moduleValue, err := vm.Require(jsModuleName, pwd)
			if err != nil {
				jsException(vm, "Error", "motto: "+err.Error())
			}

			return moduleValue
		}

		jsModule, _ := vm.Object(`({exports: {}})`)
		jsModule.Set("require", jsRequire)
		jsModule.Set("__dirname", pwd)
		jsExports, _ := jsModule.Get("exports")

		// Run the module source, with "jsModule" as the "module" variable, "jsExports" as "this"(Nodejs capable).
		var src *otto.Script
		var err error
		var sourceMap string
		if vm.SourceMapEnabled {
			sourceMapFilename := filename + ".map"
			if ok, _ := isFile(sourceMapFilename); ok {
				if content, err := ioutil.ReadFile(sourceMapFilename); err == nil {
					sourceMap = string(content)
				}
			}
		}
		if sourceMap != "" {
			src, err = vm.CompileWithSourceMap(filename, source, sourceMap)
		} else {
			src, err = vm.Compile(filename, source)
		}
		if err != nil {
			return otto.UndefinedValue(), err
		}
		srcValue, err := vm.Otto.Run(src)
		if err != nil {
			return otto.UndefinedValue(), err
		}
		moduleReturn, err := srcValue.Call(jsExports, jsModule)
		if err != nil {
			return otto.UndefinedValue(), err
		}

		var moduleValue otto.Value
		if !moduleReturn.IsUndefined() {
			moduleValue = moduleReturn
			jsModule.Set("exports", moduleValue)
		} else {
			moduleValue, _ = jsModule.Get("exports")
		}

		return moduleValue, nil
	}
}

// Create module loader from javascript file.
//
// Filename can be a javascript file or a json file.
func CreateLoaderFromFile(filename string) ModuleLoader {
	return func(vm *Motto) (otto.Value, error) {
		source, err := ioutil.ReadFile(filename)

		if err != nil {
			return otto.UndefinedValue(), err
		}

		// load json
		if filepath.Ext(filename) == ".json" {
			return vm.Call("JSON.parse", nil, string(source))
		}

		return CreateLoaderFromSource(string(source), filename)(vm)
	}
}

// Find a file module by name.
//
// If name starts with "." or "/", we search the module in the according locations
// (name and name.js and name.json).
//
// Otherwise we search the module in the "node_modules" sub-directory of "pwd" and
// "paths"
//
// It basicly follows the rules of Node.js module api: http://nodejs.org/api/modules.html
func FindFileModule(name, pwd string, paths []string) (string, error) {
	if len(name) == 0 {
		return "", errors.New("Empty module name")
	}

	add := func(choices []string, name string) []string {
		ext := filepath.Ext(name)
		if ext != ".js" && ext != ".json" {
			choices = append(choices, name+".js", name+".json")
		}
		choices = append(choices, name)
		return choices
	}

	var choices []string
	if name[0] == '.' || filepath.IsAbs(name) {
		if name[0] == '.' {
			name = filepath.Join(pwd, name)
		}

		choices = add(choices, name)
	} else {
		if pwd != "" {
			choices = add(choices, filepath.Join(pwd, "node_modules", name))
		}

		for _, v := range paths {
			choices = add(choices, filepath.Join(v, name))
		}
	}

	for _, v := range choices {
		ok, err := isDir(v)
		if err != nil {
			return "", err
		}

		if ok {
			packageJsonFilename := filepath.Join(v, "package.json")
			ok, err := isFile(packageJsonFilename)
			if err != nil {
				return "", err
			}

			var entryPoint string
			if ok {
				entryPoint, err = parsePackageEntryPoint(packageJsonFilename)
				if err != nil {
					return "", err
				}
			}

			if entryPoint == "" {
				entryPoint = "./index.js"
			}

			if !strings.HasPrefix(entryPoint, ".") {
				entryPoint = "./" + entryPoint
			}
			return FindFileModule(entryPoint, v, paths)
		}

		ok, err = isFile(v)
		if err != nil {
			return "", err
		}

		if ok {
			return filepath.Abs(v)
		}
	}

	return "", errors.New("Module not found: " + name)
}
