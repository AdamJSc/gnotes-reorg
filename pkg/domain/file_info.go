package domain

// FileInfo defines the behaviours of a file info object
type FileInfo interface {
	IsDir() bool
	Name() string
}
