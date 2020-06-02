package native

import (
	"os"
	"path/filepath"
	"strings"
)

// NormalizeSockPath make unix socket path as shorter as possiable
func NormalizeSockPath(s string) string {
	if !filepath.IsAbs(s) {
		return s
	}

	wd, _ := os.Getwd()
	if !strings.HasPrefix(s, wd) {
		return s
	}
	relpath, _ := filepath.Rel(wd, s)
	return relpath
}
