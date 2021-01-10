package fs

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
)

// GetChildPaths returns fully-qualified paths to all child paths within the provided parent directory
func GetChildPaths(parent string, dirs bool) ([]string, error) {
	paths := []string{}
	infos, err := ioutil.ReadDir(parent)
	if err != nil {
		return nil, err
	}
	for _, info := range infos {
		if info.IsDir() == dirs {
			paths = append(paths, fmt.Sprintf("%s/%s", parent, info.Name()))
		}
	}
	return paths, nil
}

// ParseDirFlag validates and returns the value of the provided flag as a formatted directory path
func ParseDirFlag(flag *string) (string, error) {
	// must not be empty
	if flag == nil {
		return "", errors.New("flag must not be nil")
	}

	val := *flag
	if val == "" {
		return "", errors.New("flag value must not be empty")
	}

	// must exist
	if err := DirExists(val); err != nil {
		return "", err
	}

	return val, nil
}

// DirExists returns an error if the provided path does not exist as a directory
func DirExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return err
	}

	// must represent a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	return nil
}
