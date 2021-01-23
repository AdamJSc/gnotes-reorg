package adapters

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"reorg/pkg/domain"
	"strings"
)

// OsFileSystem implements FileSystem for the local filesystem
type OsFileSystem struct {
	domain.FileSystem
}

// ReadFile implements FileSystem.ReadFile()
func (o *OsFileSystem) ReadFile(path string) ([]byte, error) {
	return ioutil.ReadFile(path)
}

// ReadFile implements FileSystem.ReadFile()
func (o *OsFileSystem) WriteFile(path string, data []byte, perm uint32) error {
	return ioutil.WriteFile(path, data, os.FileMode(perm))
}

// ReadDir implements FileSystem.ReadDir()
func (o *OsFileSystem) ReadDir(path string) ([]domain.FileInfo, error) {
	infos, err := ioutil.ReadDir(path)
	if err != nil {
		return nil, fmt.Errorf("os.ReadDir() failed: %w", err)
	}

	var fis []domain.FileInfo

	for _, i := range infos {
		fis = append(fis, &OsFileInfo{fi: i})
	}

	return fis, nil
}

// DirExists implements FileSystem.DirExists()
func (o *OsFileSystem) DirExists(path string) error {
	info, err := o.Stat(path)
	if err != nil {
		if o.IsNotExist(err) {
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

// IsNotExist implements FileSystem.IsNotExist()
func (o *OsFileSystem) IsNotExist(err error) bool {
	// os.IsNotExist() will not work, only checks most recent type
	// oserror package is internal so not importable
	// do it the old-school way...
	if err == nil {
		return false
	}

	return strings.HasSuffix(err.Error(), "no such file or directory")
}

// Stat implements FileSystem.Stat()
func (o *OsFileSystem) Stat(path string) (domain.FileInfo, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, fmt.Errorf("os.Stat() failed: %w", err)
	}

	return &OsFileInfo{fi: info}, nil
}

// Mkdir implements FileSystem.Mkdir()
func (o *OsFileSystem) Mkdir(path string, perm uint32) error {
	return os.Mkdir(path, os.FileMode(perm))
}

// RemoveAll implements FileSystem.RemoveAll()
func (o *OsFileSystem) RemoveAll(path string) error {
	return os.RemoveAll(path)
}

// Abs implements FileSystem.Abs()
func (o *OsFileSystem) Abs(pathParts ...string) (string, error) {
	joined := strings.Join(pathParts, string(os.PathSeparator))
	return filepath.Abs(joined)
}

// Base implements FileSystem.Base()
func (o *OsFileSystem) Base(path string) string {
	return filepath.Base(path)
}
