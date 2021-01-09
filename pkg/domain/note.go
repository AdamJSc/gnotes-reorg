package domain

import (
	"fmt"
	"strings"
	"time"
)

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
	title := strings.ReplaceAll(n.title, "/\\", "_")
	if len(title) > maxFnameTitleLen {
		title = fmt.Sprintf("%s___", title[:maxFnameTitleLen])
	}
	return fmt.Sprintf("%s %s.txt", n.timestamp.Format("2006-01-02"), title)
}
