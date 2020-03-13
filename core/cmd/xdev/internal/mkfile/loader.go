package mkfile

import (
	"container/list"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/BurntSushi/toml"
)

const (
	PkgDescFile     = "xdev.toml"
	ModDescFile     = "mod.toml"
	SelfPackageName = "self"
	XchainPackage   = "xchain"
)

// DependencyDesc 描述了依赖的package信息
type DependencyDesc struct {
	Name    string
	Path    string
	Modules []string
}

// PackgeDesc 描述了package的信息
type PackgeDesc struct {
	Package struct {
		Name   string
		Author string
	}
	Dependencies []DependencyDesc
}

// ModuleDesc 描述了module的信息
type ModuleDesc struct {
	Dependencies []DependencyDesc
}

type moduleNode struct {
	name    string
	pkgPath string
	pkgDesc *PackgeDesc

	deps []DependencyDesc
}

func (m *moduleNode) fullName() string {
	return m.pkgDesc.Package.Name + "/" + m.name
}

func (m *moduleNode) path() string {
	return filepath.Join(m.pkgPath, "src", m.name)
}

// Loader 分析package依赖，并确保所有依赖都已经在本地存在
type Loader struct {
	root       string
	xdevRoot   string
	searchPath []string
}

// NewLoader instance Loader
func NewLoader() *Loader {
	return &Loader{}
}

func (l *Loader) WithXROOT(xroot string) *Loader {
	l.xdevRoot = xroot
	return l
}

func (l *Loader) WithSearchPath(paths []string) *Loader {
	l.searchPath = paths
	return l
}

// Load parse and load all package's dep packages
func (l *Loader) Load(pkgpath string, addons []DependencyDesc) (*Package, error) {
	if !filepath.IsAbs(pkgpath) {
		pkgpath, _ = filepath.Abs(pkgpath)
	}
	l.root = pkgpath
	entryDesc, err := ParsePackageDesc(pkgpath)
	if err != nil {
		return nil, err
	}

	entryMod := ""
	deps := append([]DependencyDesc{}, entryDesc.Dependencies...)
	deps = append(deps, addons...)
	rootNode := &moduleNode{
		name:    entryMod,
		pkgPath: pkgpath,
		pkgDesc: entryDesc,
		deps:    deps,
	}

	nodes := make(map[string]*moduleNode)
	// 广度搜索，剔除已经加载的module
	queue := list.New()
	queue.PushBack(rootNode)
	for queue.Len() != 0 {
		node := queue.Remove(queue.Front()).(*moduleNode)
		modFullName := node.fullName()
		if _, ok := nodes[modFullName]; ok {
			continue
		}
		nodes[modFullName] = node

		deps, err := l.parseModuleDeps(node)
		if err != nil {
			return nil, err
		}

		for _, dep := range deps {
			queue.PushBack(dep)
		}
	}

	entryPkg := &Package{
		Name: entryDesc.Package.Name,
		Path: pkgpath,
		Deps: make(map[string]*Package),
		Modules: []Module{
			{
				Name: entryMod,
				Path: rootNode.path(),
			},
		},
	}

	modules := make([]string, 0, len(nodes))
	for mod := range nodes {
		modules = append(modules, mod)
	}
	sort.Strings(modules)

	for _, mod := range modules {
		node := nodes[mod]
		if node.pkgDesc.Package.Name == entryDesc.Package.Name {
			if node.name != "" {
				entryPkg.Modules = append(entryPkg.Modules, Module{
					Name: node.name,
					Path: node.path(),
				})
			}
			continue
		}
		pkg, ok := entryPkg.Deps[node.pkgDesc.Package.Name]
		if !ok {
			pkg = &Package{
				Name: node.pkgDesc.Package.Name,
				Path: node.pkgPath,
			}
			entryPkg.Deps[node.pkgDesc.Package.Name] = pkg
		}
		pkg.Modules = append(pkg.Modules, Module{
			Name: node.name,
			Path: node.path(),
		})
	}

	return entryPkg, nil
}

func ParsePackageDesc(pkgpath string) (*PackgeDesc, error) {
	pkgDescFile := filepath.Join(pkgpath, PkgDescFile)
	var desc PackgeDesc
	_, err := toml.DecodeFile(pkgDescFile, &desc)
	if err != nil {
		return nil, err
	}
	return &desc, nil
}

func ParseModuleDesc(modpath string) (*ModuleDesc, error) {
	modDescFile := filepath.Join(modpath, ModDescFile)
	var desc ModuleDesc
	_, err := toml.DecodeFile(modDescFile, &desc)
	if err != nil {
		return nil, err
	}
	return &desc, nil
}

func pathExists(p string) bool {
	_, err := os.Stat(p)
	return err == nil
}

func (l *Loader) loadPackage(wd string, desc *DependencyDesc) (string, error) {
	if desc.Name == SelfPackageName {
		return wd, nil
	}
	if desc.Name == XchainPackage {
		return l.xdevRoot, nil
	}
	var depFullPath string
	searchPath := append([]string{}, l.searchPath...)
	searchPath = append(searchPath, filepath.Join(l.root, "vendor"))

	// 如果依赖路径为空，项目vendor和全局搜索目录里面寻找
	if desc.Path == "" {
		for _, path := range searchPath {
			depFullPath = filepath.Join(path, desc.Name)
			if pathExists(depFullPath) {
				return depFullPath, nil
			}
		}
	}
	if filepath.IsAbs(desc.Path) && pathExists(desc.Path) {
		return desc.Path, nil
	}

	depFullPath = filepath.Join(wd, desc.Path)
	if pathExists(depFullPath) {
		return filepath.Abs(depFullPath)
	}
	return "", fmt.Errorf("package '%s' not found, search path:%v, pwd:%s", desc.Name, searchPath, wd)
}

func (l *Loader) parseModuleDeps(node *moduleNode) (map[string]*moduleNode, error) {
	modules := make(map[string]*moduleNode)
	for _, dep := range node.deps {
		if dep.Name == MainPackage {
			return nil, errors.New("can not use main package as dependency")
		}

		depPkgPath, err := l.loadPackage(node.pkgPath, &dep)
		if err != nil {
			return nil, err
		}

		depDescFile := filepath.Join(depPkgPath, PkgDescFile)
		if _, err := os.Stat(depDescFile); err != nil {
			return nil, fmt.Errorf("%s for package '%s' not found, pkg path:%s", PkgDescFile, dep.Name, depPkgPath)
		}
		desc, err := ParsePackageDesc(depPkgPath)
		if err != nil {
			return nil, err
		}
		if dep.Name != SelfPackageName && desc.Package.Name != dep.Name {
			return nil, fmt.Errorf("mismatched package name %s:%s", desc.Package.Name, dep.Name)
		}
		packageMod := &moduleNode{
			name:    "",
			pkgPath: depPkgPath,
			pkgDesc: desc,
			deps:    desc.Dependencies,
		}
		modules[packageMod.fullName()] = packageMod

		for _, mod := range dep.Modules {
			modpath := filepath.Join(depPkgPath, "src", mod)
			if _, err := os.Stat(modpath); err != nil {
				return nil, fmt.Errorf("module %s in package %s not found", mod, dep.Name)
			}
			modNode := &moduleNode{
				name:    mod,
				pkgPath: depPkgPath,
				pkgDesc: desc,
			}
			moddesc, err := ParseModuleDesc(modpath)
			if err == nil {
				modNode.deps = moddesc.Dependencies
			}

			fullName := modNode.fullName()
			if _, ok := modules[fullName]; ok {
				continue
			}
			modules[fullName] = modNode
		}
	}
	return modules, nil
}
