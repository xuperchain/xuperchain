package mkfile

import (
	"reflect"
	"testing"
)

func TestPrefixPath(t *testing.T) {
	paths := []string{
		"/home",
		"/home/work",
		"/lib/a",
		"/lib/b",
		"/lib/a/b",
	}
	ret := prefixPaths(paths)
	expect := []string{
		"/home",
		"/lib/a",
		"/lib/b",
	}
	if !reflect.DeepEqual(ret, expect) {
		t.Fatalf("ret:%v, expect:%v", ret, expect)
	}
}
