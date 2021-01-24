package domain

// FileSystem implements the behaviours of a file system
type FileSystem interface {
	ReadFile(path string) ([]byte, error)
	WriteFile(path string, data []byte, perm uint32) error
	ReadDir(path string) ([]FileInfo, error)
	DirExists(path string) error
	IsNotExist(err error) bool
	Stat(path string) (FileInfo, error)
	Mkdir(path string, perm uint32) error
	RemoveAll(path string) error
	Abs(pathParts ...string) (string, error)
	Dir(path string) string
	Base(path string) string
}
