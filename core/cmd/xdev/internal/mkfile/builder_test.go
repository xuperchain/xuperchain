package mkfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackageParse(t *testing.T) {
	loader := NewLoader()
	pkgpath := filepath.Join("testdata", "m1")
	pkg, err := loader.Load(pkgpath)
	if err != nil {
		t.Fatal(err)
	}

	builder := NewBuilder()
	err = builder.Parse(pkg)
	if err != nil {
		t.Fatal(err)
	}
	err = builder.GenerateMakeFile(os.Stderr)
	if err != nil {
		t.Fatal(err)
	}
}
