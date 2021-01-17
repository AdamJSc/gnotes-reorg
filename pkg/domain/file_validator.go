package domain

import "strings"

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
