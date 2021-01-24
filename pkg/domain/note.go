package domain

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
)

const maxFnameTitleLen = 30

// Note represents a single Note
type Note struct {
	ID           string    `json:"id"`           // numeric gnotes id
	Index        int       `json:"-"`            // index of note within a slice
	ParentDir    string    `json:"-"`            // parent directory of note once cleaned (inflated, not stored)
	Category     string    `json:"-"`            // category of note (inflated, not stored)
	OriginalPath string    `json:"originalPath"` // original full-qualified path to note html source file
	Title        string    `json:"title"`        // title of the note
	Timestamp    time.Time `json:"timestamp"`    // timestamp of the note
	Content      string    `json:"content"`      // content of the note
}

// MarshalJSON implements custom marshaler on Note struct
func (n *Note) MarshalJSON() ([]byte, error) {
	type noteAlias Note
	var payload = struct {
		Filename string `json:"filename"`
		noteAlias
	}{
		Filename:  n.Filename(),
		noteAlias: noteAlias(*n),
	}

	return json.Marshal(payload)
}

// Filename returns a generated filename
func (n Note) Filename() string {
	title := strings.ToLower(n.Title)
	baseName := sanitize.BaseName(title)
	fileName := fmt.Sprintf("%s_%s", n.Timestamp.Format("2006-01-02"), baseName)
	if len(fileName) > maxFnameTitleLen {
		fileName = fmt.Sprintf("%s__", fileName[:maxFnameTitleLen])
	}
	return fileName
}

// NoteManifest maps a note filename to its category
type NoteManifest struct {
	path    string
	content map[string]string
}

// Len returns count of filenames with categories
func (nm *NoteManifest) Len() int {
	return len(nm.content)
}

// Set assigns the provided Note to the manifest
func (nm *NoteManifest) Set(n Note) error {
	if nm.content == nil {
		nm.content = make(map[string]string)
	}

	filename := n.Filename()

	if nm.HasCat(filename) {
		return fmt.Errorf("filename %s already has category", filename)
	}

	nm.content[filename] = n.Category

	return nil
}

// HasCat returs true if existing filename has a category
func (nm *NoteManifest) HasCat(filename string) bool {
	if nm.content == nil {
		nm.content = make(map[string]string)
	}

	_, ok := nm.content[filename]

	return ok
}

// MarshalJSON implements json.Marshaler
func (nm *NoteManifest) MarshalJSON() ([]byte, error) {
	return json.Marshal(nm.content)
}

// UnmarshalJSON implements json.Unmarshaler
func (nm *NoteManifest) UnmarshalJSON(b []byte) error {
	var content map[string]string

	if err := json.Unmarshal(b, &content); err != nil {
		return err
	}

	nm.content = content

	return nil
}
