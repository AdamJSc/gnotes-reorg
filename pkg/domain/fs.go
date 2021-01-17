package domain

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// FileSystemService provides file system related functionality
type FileSystemService struct {
	fs FileSystem
}

// GetChildPaths returns fully-qualified paths to all child paths within the provided parent directory.
//
// Will return only paths to directories if onlyDirs is true, otherwise only paths to files
func (f *FileSystemService) GetChildPaths(parent string, validators ...FileValidator) ([]string, error) {
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
func (f *FileSystemService) DirExists(path string) error {
	return f.fs.DirExists(path)
}

// ParseAbsPath parses the absolute path of the provided components and assigns top val
func (f *FileSystemService) ParseAbsPath(val *string, parts ...string) error {
	joined := strings.Join(parts, string(os.PathSeparator))

	abs, err := f.fs.Abs(joined)
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

// FileSystem implements the behaviours of a file system
type FileSystem interface {
	ReadDir(path string) ([]FileInfo, error)
	DirExists(path string) error
	Abs(path string) (string, error)
}

// OsFileSystem implements FileSystem for the local filesystem
type OsFileSystem struct{}

// ReadDir implements FileSystem.ReadDir()
func (o *OsFileSystem) ReadDir(path string) ([]FileInfo, error) {
	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("os.ReadDir() failed: %w", err)
	}

	var fis []FileInfo

	for _, i := range infos {
		fis = append(fis, &OsFileInfo{fi: i})
	}

	return fis, nil
}

// DirExists implements FileSystem.DirExists()
func (o *OsFileSystem) DirExists(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("path does not exist: %s", path)
		}
		return fmt.Errorf("os.Stat() failed: %w", err)
	}

	// must represent a directory
	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", path)
	}

	return nil
}

// Abs implements FileSystem.Abs()
func (o *OsFileSystem) Abs(path string) (string, error) {
	return filepath.Abs(path)
}

// FileInfo defines the behaviours of a file info object
type FileInfo interface {
	IsDir() bool
	Name() string
}

// OsFileInfo implements FileInfo for the local file system
type OsFileInfo struct {
	fi os.FileInfo
}

// IsDir implements FileInfo.IsDir()
func (o *OsFileInfo) IsDir() bool {
	return o.fi.IsDir()
}

// IsDir implements FileInfo.Name()
func (o *OsFileInfo) Name() string {
	return o.fi.Name()
}

// FileValidator defines the behaviour of a file validator
type FileValidator interface {
	Valid(f FileInfo) bool
}

// IsDir defines a file validator that checks whether file info refers to a path that is a directory
type IsDir struct{}

// Valid implements FileValidator.Valid()
func (i *IsDir) Valid(f FileInfo) bool {
	return f.IsDir()
}

// IsNotDir defines a file validator that checks whether file info refers to a path that is not a directory
type IsNotDir struct{}

// Valid implements FileValidator.Valid()
func (i *IsNotDir) Valid(f FileInfo) bool {
	return !f.IsDir()
}

// IsJSON defines a file validator that checks for JSON files
type IsJSON struct{}

// Valid implements FileValidator.Valid()
func (i *IsJSON) Valid(f FileInfo) bool {
	return strings.HasSuffix(f.Name(), ".json")
}

// IsNotName defines a file validator that checks for file name
type IsNotName struct {
	BaseNames []string
}

// Valid implements FileValidator.Valid()
func (i *IsNotName) Valid(f FileInfo) bool {
	for _, b := range i.BaseNames {
		if b == f.Name() {
			return false
		}
	}
	return true
}

// fileIsValid runs the provided file through the provided file validators and returns true if all of them pass
func fileIsValid(f FileInfo, validators []FileValidator) bool {
	for _, v := range validators {
		if !v.Valid(f) {
			return false
		}
	}
	return true
}
