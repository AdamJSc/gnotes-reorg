package adapters

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"reorg/pkg/domain"
	"strings"
)

// JSONNoteWriter writes a Note as a JSON file
type JSONNoteWriter struct {
	domain.NoteWriter
	SubDir string // represents sub-directory to write note to
	Files  *domain.FileSystemService
}

// Write implements domain.NoteWriter
func (j *JSONNoteWriter) Write(n domain.Note) error {
	parentDir := n.ParentDir

	if j.SubDir != "" {
		parentDir = strings.Join([]string{parentDir, j.SubDir}, string(os.PathSeparator))
	}

	if n.Category != "" {
		parentDir = strings.Join([]string{parentDir, n.Category}, string(os.PathSeparator))
	}

	// save note
	filePath, err := generateAbsFilePath(j.Files, parentDir, n.Filename(), "json", n.Index)
	if err != nil {
		return fmt.Errorf("cannot generate file path: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(&n); err != nil {
		return fmt.Errorf("cannot parse json: %w", err)
	}

	if err := j.Files.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("cannot write note with id %s: %w", n.ID, err)
	}

	return nil
}
