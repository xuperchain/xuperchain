package mkfile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestPackageParse(t *testing.T) {
	loader := NewLoader()
	pkgpath := filepath.Join("testdata", "pkg1")
	pkg, err := loader.Load(pkgpath, nil)
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

	err = builder.GenerateCompileCommands(os.Stderr)
	if err != nil {
		t.Fatal(err)
	}
}
