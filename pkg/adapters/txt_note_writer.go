package adapters

import (
	"fmt"
	"os"
	"reorg/pkg/domain"
	"strings"
)

// TxtNoteWriter writes a Note as a text file
type TxtNoteWriter struct {
	domain.NoteWriter
	SubDir string // represents sub-directory to write note to
	Files  *domain.FileSystemService
}

// Write implements domain.NoteWriter
func (t *TxtNoteWriter) Write(n domain.Note) error {
	parentDir := n.ParentDir

	var err error

	if t.SubDir != "" {
		parentDir, err = joinAndCreate(t.Files, []string{parentDir, t.SubDir})
		if err != nil {
			return err
		}
	}

	if n.Category != "" {
		parentDir, err = joinAndCreate(t.Files, []string{parentDir, n.Category})
		if err != nil {
			return err
		}
	}

	// create parent directory if it doesn't already exist
	if err := t.Files.DirExists(parentDir); err != nil {
		if err := t.Files.MakeDir(parentDir); err != nil {
			return fmt.Errorf("cannot make directory %s: %w", parentDir, err)
		}
	}

	// save note
	filePath, err := generateAbsFilePath(t.Files, parentDir, n.Filename(), "txt", n.Index)
	if err != nil {
		return fmt.Errorf("cannot generate file path: %w", err)
	}

	content := []byte(n.Content)

	if err := t.Files.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("cannot write note with id %s: %w", n.ID, err)
	}

	return nil
}

// joinAndCreate joins the provided file parts and attempts to create the path as a directory
func joinAndCreate(f *domain.FileSystemService, parts []string) (string, error) {
	path := strings.Join(parts, string(os.PathSeparator))

	if err := f.DirExists(path); err != nil {
		if err := f.MakeDir(path); err != nil {
			return "", fmt.Errorf("cannot make directory %s: %w", path, err)
		}
	}

	return path, nil
}

// generateAbsFilePath generates a file path from the provided arguments
func generateAbsFilePath(f *domain.FileSystemService, dir, name, ext string, i int) (string, error) {
	var suffix string
	if i > 0 {
		suffix = fmt.Sprintf("_%d", i)
	}

	fileNameWithExt := fmt.Sprintf("%s%s.%s", name, suffix, ext)

	absPath, err := f.ParseAbsPath(dir, fileNameWithExt)
	if err != nil {
		return "", fmt.Errorf("cannot parse absolute path: %w", err)
	}

	return absPath, nil
}
