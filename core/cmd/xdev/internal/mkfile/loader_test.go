package mkfile

import (
	"path/filepath"
	"testing"
)

func TestLoader(t *testing.T) {
	loader := NewLoader()
	pkgpath := filepath.Join("testdata", "m1")
	pkg, err := loader.Load(pkgpath)
	if err != nil {
		t.Fatal(err)
	}
	if pkg.Name != "m1" {
		t.Fatal("not equal")
	}
	if len(pkg.Deps) != 0 {
		t.Fatal("not zero")
	}
}
