package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// NoteWriter defines the required behaviour for writing a Note
type NoteWriter interface {
	Write(n noteWithIndex) error
}

// JSONNoteWriter writes a Note as a JSON file
type JSONNoteWriter struct {
	NoteWriter
	Files *FileSystemService
}

// Write implements NoteWriter
func (j *JSONNoteWriter) Write(ni noteWithIndex) error {
	// save note
	filePath, err := generateAbsFilePath(j.Files, ni.note.ParentDir, ni.note.Filename(), "json", ni.idx)
	if err != nil {
		return fmt.Errorf("cannot generate file path: %w", err)
	}

	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(&ni.note); err != nil {
		return fmt.Errorf("cannot parse json: %w", err)
	}

	if err := j.Files.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("cannot write note with id %s: %w", ni.note.ID, err)
	}

	return nil
}

// TxtNoteWriter writes a Note as a text file
type TxtNoteWriter struct {
	NoteWriter
	Files *FileSystemService
}

// Write implements NoteWriter
func (t *TxtNoteWriter) Write(ni noteWithIndex) error {
	// save note
	filePath, err := generateAbsFilePath(t.Files, ni.note.ParentDir, ni.note.Filename(), "txt", ni.idx)
	if err != nil {
		return fmt.Errorf("cannot generate file path: %w", err)
	}

	content := []byte(ni.note.Content)

	if err := t.Files.WriteFile(filePath, content, 0644); err != nil {
		return fmt.Errorf("cannot write note with id %s: %w", ni.note.ID, err)
	}

	return nil
}

// generateAbsFilePath generates a file path from the provided arguments
func generateAbsFilePath(fs *FileSystemService, dir, name, ext string, i int) (string, error) {
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
