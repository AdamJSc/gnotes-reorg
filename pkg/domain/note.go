package domain

import (
	"fmt"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
)

const maxFnameTitleLen = 30

// Note represents a single Note
type Note struct {
	id        string
	path      string
	title     string
	timestamp time.Time
	content   string
}

// filename returns a generated filename
func (n Note) filename() string {
	title := strings.ToLower(n.title)
	baseName := sanitize.BaseName(title)
	fileName := fmt.Sprintf("%s_%s", n.timestamp.Format("2006-01-02"), baseName)
	if len(fileName) > maxFnameTitleLen {
		fileName = fmt.Sprintf("%s__", fileName[:maxFnameTitleLen])
	}
	return fmt.Sprintf("%s.txt", fileName)
}
