package mkfile

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoader(t *testing.T) {
	loader := NewLoader()
	pkgpath := filepath.Join("testdata", "pkg1")
	pkg, err := loader.Load(pkgpath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Name != "pkg1" {
		t.Fatal("not equal")
	}

	wd, _ := os.Getwd()
	expect := &Package{
		Name: "pkg1",
		Path: filepath.Join(wd, "testdata", "pkg1"),
		Modules: []Module{
			Module{Name: "", Path: filepath.Join(wd, "testdata", "pkg1", "src")},
		},
		Deps: map[string]*Package{
			"pkg2": &Package{
				Name: "pkg2",
				Path: filepath.Join(wd, "testdata", "pkg2"),
				Modules: []Module{
					Module{Name: "", Path: filepath.Join(wd, "testdata", "pkg2", "src")},
					Module{Name: "m1", Path: filepath.Join(wd, "testdata", "pkg2", "src", "m1")},
					Module{Name: "m2", Path: filepath.Join(wd, "testdata", "pkg2", "src", "m2")},
					Module{Name: "m3", Path: filepath.Join(wd, "testdata", "pkg2", "src", "m3")},
				},
			},
			"pkg3": &Package{
				Name: "pkg3",
				Path: filepath.Join(wd, "testdata", "pkg3"),
				Modules: []Module{
					Module{Name: "", Path: filepath.Join(wd, "testdata", "pkg3", "src")},
				},
			},
			"pkg4": &Package{
				Name: "pkg4",
				Path: filepath.Join(wd, "testdata", "pkg1", "vendor", "pkg4"),
				Modules: []Module{
					Module{Name: "", Path: filepath.Join(wd, "testdata", "pkg1", "vendor", "pkg4", "src")},
				},
			},
		},
	}
	if !reflect.DeepEqual(pkg, expect) {
		t.Logf("%#v", pkg)
		t.Fatalf("%#v", expect)
	}
}

func TestAddon(t *testing.T) {
	loader := NewLoader()
	pkgpath := filepath.Join("testdata", "pkg3")
	pkg, err := loader.Load(pkgpath, []DependencyDesc{
		{
			Name:    "pkg2",
			Path:    filepath.Join("..", "pkg2"),
			Modules: []string{"m3"},
		},
	})
	if err != nil {
		t.Fatal(err)
	}

	wd, _ := os.Getwd()
	expect := &Package{
		Name: "pkg3",
		Path: filepath.Join(wd, "testdata", "pkg3"),
		Modules: []Module{
			Module{Name: "", Path: filepath.Join(wd, "testdata", "pkg3", "src")},
		},
		Deps: map[string]*Package{
			"pkg2": &Package{
				Name: "pkg2",
				Path: filepath.Join(wd, "testdata", "pkg2"),
				Modules: []Module{
					Module{Name: "", Path: filepath.Join(wd, "testdata", "pkg2", "src")},
					Module{Name: "m3", Path: filepath.Join(wd, "testdata", "pkg2", "src", "m3")},
				},
			},
		},
	}
	if !reflect.DeepEqual(pkg, expect) {
		t.Logf("%#v", pkg)
		t.Fatalf("%#v", expect)
	}
}
