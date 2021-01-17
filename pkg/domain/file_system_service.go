package domain

import (
	"context"
	"fmt"
)

// FileSystemService provides file system related functionality
type FileSystemService struct {
	fs FileSystem
}

// GetChildPaths returns fully-qualified paths to all child paths within the provided parent directory.
//
// Will return only paths to directories if onlyDirs is true, otherwise only paths to files
func (f *FileSystemService) GetChildPaths(ctx context.Context, parent string, validators ...FileValidator) ([]string, error) {
	paths := []string{}

	infos, err := f.fs.ReadDir(parent)
	if err != nil {
		return nil, err
	}

	for _, info := range infos {
		if fileIsValid(info, validators) {
			paths = append(paths, fmt.Sprintf("%s/%s", parent, info.Name()))
		}
	}

	return paths, nil
}

// DirExists returns an error if the provided path does not exist as a directory
func (f *FileSystemService) DirExists(ctx context.Context, path string) error {
	return f.fs.DirExists(path)
}

// MakeDir attempts to make the directory at the given path
func (f *FileSystemService) MakeDir(ctx context.Context, path string) error {
	return f.fs.Mkdir(path, 0755)
}

// ParseAbsPath parses the absolute path of the provided components and assigns top val
func (f *FileSystemService) ParseAbsPath(ctx context.Context, val *string, parts ...string) error {
	abs, err := f.fs.Abs(parts...)
	if err != nil {
		return fmt.Errorf("absolute path failed: %w", err)
	}

	*val = abs

	return nil
}

// NewFileSystemService returns a new FileSystemService using the provided FileSystem
func NewFileSystemService(fs FileSystem) *FileSystemService {
	return &FileSystemService{fs: fs}
}
