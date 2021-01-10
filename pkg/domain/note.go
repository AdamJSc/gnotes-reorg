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
	ID        string    `json:"id"`
	Path      string    `json:"originalPath"`
	Title     string    `json:"title"`
	Timestamp time.Time `json:"timestamp"`
	Content   string    `json:"content"`
}

// MarshalJSON implements custom marshaler on Note struct
func (n *Note) MarshalJSON() ([]byte, error) {
	type noteAlias Note
	var payload = struct {
		Filename string `json:"filename"`
		noteAlias
	}{
		Filename:  n.filename("json"),
		noteAlias: noteAlias(*n),
	}

	return json.Marshal(payload)
}

// filename returns a generated filename
func (n Note) filename(ext string) string {
	title := strings.ToLower(n.Title)
	baseName := sanitize.BaseName(title)
	fileName := fmt.Sprintf("%s_%s", n.Timestamp.Format("2006-01-02"), baseName)
	if len(fileName) > maxFnameTitleLen {
		fileName = fmt.Sprintf("%s__", fileName[:maxFnameTitleLen])
	}
	return fmt.Sprintf("%s.%s", fileName, ext)
}
