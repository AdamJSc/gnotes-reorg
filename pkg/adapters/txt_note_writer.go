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
	Files *domain.FileSystemService
}

// Write implements domain.NoteWriter
func (t *TxtNoteWriter) Write(n domain.Note) error {
	parentDir := n.ParentDir

	if n.Category != "" {
		parentDir = strings.Join([]string{n.Category, n.ParentDir}, string(os.PathSeparator))
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

// generateAbsFilePath generates a file path from the provided arguments
func generateAbsFilePath(fs *domain.FileSystemService, dir, name, ext string, i int) (string, error) {
	var suffix string
	if i > 0 {
		suffix = fmt.Sprintf("_%d", i)
	}

	fileNameWithExt := fmt.Sprintf("%s%s.%s", name, suffix, ext)

	absPath, err := fs.ParseAbsPath(dir, fileNameWithExt)
	if err != nil {
		return "", fmt.Errorf("cannot parse absolute path: %w", err)
	}

	return absPath, nil
}
