package mkfile

import (
	"container/list"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	PkgDescFile = "xdev.toml"
)

type Dependency struct {
	Path string
}

type PackgeDesc struct {
	Package struct {
		Name   string
		Author string
	}
	Dependencies map[string]Dependency

	path string
}

type Loader struct {
	root string
}

func NewLoader() *Loader {
	return &Loader{}
}

// Load parse and load all package's dep packages
func (l *Loader) Load(pkgpath string) (*Package, error) {
	if !filepath.IsAbs(pkgpath) {
		pkgpath, _ = filepath.Abs(pkgpath)
	}
	l.root = pkgpath
	entryDesc, err := l.parsePackageDesc(pkgpath)
	if err != nil {
		return nil, err
	}
	entryDesc.path = pkgpath
	pkgs := make(map[string]*Package)

	// 广度搜索，剔除已经加载的package
	queue := list.New()
	queue.PushBack(entryDesc)
	for queue.Len() != 0 {
		desc := queue.Remove(queue.Front()).(*PackgeDesc)
		if _, ok := pkgs[desc.Package.Name]; ok {
			continue
		}

		pkgs[desc.Package.Name] = &Package{
			Name: desc.Package.Name,
			Path: desc.path,
		}

		deps, err := l.parsePackageDeps(desc)
		if err != nil {
			return nil, err
		}

		for _, dep := range deps {
			queue.PushBack(dep)
		}
	}

	depPkgs := make([]*Package, 0, len(pkgs))
	for _, pkg := range pkgs {
		if pkg.Name == entryDesc.Package.Name {
			continue
		}
		depPkgs = append(depPkgs, pkg)
	}
	return &Package{
		Name: entryDesc.Package.Name,
		Path: pkgpath,
		Deps: depPkgs,
	}, nil
}

func (l *Loader) parsePackageDesc(pkgpath string) (*PackgeDesc, error) {
	pkgDescFile := filepath.Join(pkgpath, PkgDescFile)
	var desc PackgeDesc
	_, err := toml.DecodeFile(pkgDescFile, &desc)
	if err != nil {
		return nil, err
	}
	return &desc, nil
}

func (l *Loader) parsePackageDeps(pkg *PackgeDesc) ([]*PackgeDesc, error) {
	ret := make([]*PackgeDesc, 0, len(pkg.Dependencies))
	for name, dep := range pkg.Dependencies {
		if name == MainPackage {
			return nil, errors.New("can not use main package as dependency")
		}

		depFullPath := dep.Path
		// 如果依赖路径为空，则默认从entry package的vendor目录下寻找同名的package
		if depFullPath == "" {
			depFullPath = filepath.Join(l.root, "vendor", name)
		}
		if !filepath.IsAbs(depFullPath) {
			depFullPath, _ = filepath.Abs(filepath.Join(pkg.path, depFullPath))
		}
		depDescFile := filepath.Join(depFullPath, PkgDescFile)
		if _, err := os.Stat(depDescFile); err != nil {
			return nil, fmt.Errorf("package %s not found", name)
		}
		desc, err := l.parsePackageDesc(depFullPath)
		if err != nil {
			return nil, err
		}
		if desc.Package.Name != name {
			return nil, fmt.Errorf("mismatched package name %s:%s", desc.Package.Name, name)
		}
		desc.path = depFullPath
		ret = append(ret, desc)
	}
	return ret, nil
}
