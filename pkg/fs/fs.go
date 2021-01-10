package fs

import (
	"fmt"
	"io/ioutil"
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
