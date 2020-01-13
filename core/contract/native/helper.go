package native

import (
	"os"
	"path/filepath"
)

// RelPathOfCWD 返回工作目录的相对路径
func RelPathOfCWD(rootpath string) (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}
	socketPath, err := filepath.Rel(cwd, rootpath)
	if err != nil {
		return "", err
	}
	return socketPath, nil
}
