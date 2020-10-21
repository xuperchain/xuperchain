package keys

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const (
	DefaultHost     = "localhost"
	DefaultPort     = "10997"
	DefaultHashType = "sha256"
	DefaultKeysDir  = ".keys"
	TestPort        = "0"
)

func returnDataDir(dir string) (string, error) {
	dir = path.Join(dir, "data")
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return dir, checkMakeDataDir(dir)
}

func returnNamesDir(dir string) (string, error) {
	dir = path.Join(dir, "names")
	dir, err := filepath.Abs(dir)
	if err != nil {
		return "", err
	}
	return dir, checkMakeDataDir(dir)
}

//----------------------------------------------------------------
func writeKey(keyDir string, addr, keyJson []byte) ([]byte, error) {
	dir, err := returnDataDir(keyDir)
	if err != nil {
		return nil, fmt.Errorf("failed to get keys dir: %v", err)
	}
	if err := WriteKeyFile(addr, dir, keyJson); err != nil {
		return nil, err
	}
	return addr, nil
}

//----------------------------------------------------------------
// manage names for keys

func coreNameAdd(keysDir, name, addr string) error {
	namesDir, err := returnNamesDir(keysDir)
	if err != nil {
		return err
	}
	dataDir, err := returnDataDir(keysDir)
	if err != nil {
		return err
	}
	if _, err := os.Stat(path.Join(dataDir, addr+".json")); err != nil {
		return fmt.Errorf("unknown key %s", addr)
	}
	return ioutil.WriteFile(path.Join(namesDir, name), []byte(addr), 0600)
}

func coreNameList(keysDir string) (map[string]string, error) {
	dir, err := returnNamesDir(keysDir)
	if err != nil {
		return nil, err
	}
	names := make(map[string]string)
	fs, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, f := range fs {
		b, err := ioutil.ReadFile(path.Join(dir, f.Name()))
		if err != nil {
			return nil, err
		}
		names[f.Name()] = string(b)
	}
	return names, nil
}

func coreNameRm(keysDir string, name string) error {
	dir, err := returnNamesDir(keysDir)
	if err != nil {
		return err
	}
	return os.Remove(path.Join(dir, name))
}

func coreNameGet(keysDir, name string) (string, error) {
	dir, err := returnNamesDir(keysDir)
	if err != nil {
		return "", err
	}
	b, err := ioutil.ReadFile(path.Join(dir, name))
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func checkMakeDataDir(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		err = os.MkdirAll(dir, 0700)
		if err != nil {
			return err
		}
	}
	return nil
}

// return addr from name or addr
func getNameAddr(keysDir, name, addr string) (string, error) {
	if name == "" && addr == "" {
		return "", fmt.Errorf("at least one of name or addr must be provided")
	}

	// name takes precedent if both are given
	var err error
	if name != "" {
		addr, err = coreNameGet(keysDir, name)
		if err != nil {
			return "", err
		}
	}
	return strings.ToUpper(addr), nil
}
