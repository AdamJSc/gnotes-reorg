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
	Path         string    `json:"-"`            // parent directory of note once cleaned
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
		Filename:  n.filename(),
		noteAlias: noteAlias(*n),
	}

	return json.Marshal(payload)
}

// filename returns a generated filename
func (n Note) filename() string {
	title := strings.ToLower(n.Title)
	baseName := sanitize.BaseName(title)
	fileName := fmt.Sprintf("%s_%s", n.Timestamp.Format("2006-01-02"), baseName)
	if len(fileName) > maxFnameTitleLen {
		fileName = fmt.Sprintf("%s__", fileName[:maxFnameTitleLen])
	}
	return fileName
}
