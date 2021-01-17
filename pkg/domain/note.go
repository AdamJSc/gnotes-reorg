package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/kennygrant/sanitize"
	"github.com/microcosm-cc/bluemonday"
)

const maxFnameTitleLen = 30

// Note represents a single Note
type Note struct {
	ID           string    `json:"id"`           // numeric gnotes id
	Path         string    `json:"-"`            // parent directory of note once cleaned (inflated, not stored)
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

// NoteService provides note-related functionality
type NoteService struct {
	fs FileSystem
}

// ParseFromDirs parses Notes from the provided directory paths
func (ns *NoteService) ParseFromDirs(paths []string) ([]Note, error) {
	var notes []Note

	for _, p := range paths {
		contentPath, err := ns.fs.Abs(p, "content.html")
		if err != nil {
			return nil, fmt.Errorf("cannot get absolute file path: %w", err)
		}

		id := ns.fs.Base(p)

		n, err := ns.parseFromFile(contentPath, id)
		if err != nil {
			return nil, fmt.Errorf("cannot parse note from file %s: %w", p, err)
		}

		notes = append(notes, n)
	}

	return notes, nil
}

// parseFromFile parses a Note from the provided file path
func (ns *NoteService) parseFromFile(path, id string) (Note, error) {
	b, err := ns.fs.ReadFile(path)
	if err != nil {
		return Note{}, err
	}

	content := string(b)
	sanitised, err := sanitiseInput(content)
	if err != nil {
		return Note{}, err
	}

	n := Note{
		ID:           id,
		OriginalPath: path,
	}

	if err := parseTitle(sanitised, &n.Title); err != nil {
		return Note{}, err
	}
	if err := parseTimestamp(sanitised, &n.Timestamp); err != nil {
		return Note{}, err
	}
	if err := parseNoteContent(sanitised, &n.Content); err != nil {
		return Note{}, err
	}

	return n, nil
}

// NewNoteService returns a new NoteService using the provided FileSystem
func NewNoteService(fs FileSystem) *NoteService {
	return &NoteService{fs: fs}
}

// sanitiseInput sanitises the provided input string
func sanitiseInput(inp string) (string, error) {
	var findReplace = func(inp string, fr map[string]string) string {
		filtered := inp
		for f, r := range fr {
			rgx, err := regexp.Compile(f)
			if err != nil {
				log.Fatal(err)
			}
			filtered = string(rgx.ReplaceAll([]byte(filtered), []byte(r)))
			filtered = strings.Trim(filtered, " \n")
		}
		return filtered
	}

	filtered := inp

	// initial character replacements
	filtered = findReplace(filtered, map[string]string{
		"&ensp;":   " ",
		"&quot;":   "\"",
		"&amp;":    "&",
		"<br>":     "\n",
		"<p(.*?)>": "\n\n",
	})

	// strip tags
	filtered = bluemonday.StripTagsPolicy().Sanitize(filtered)

	// add non-sanitised characters back in
	filtered = findReplace(filtered, map[string]string{
		"&#39;": "'",
		"&#34;": "\"",
	})

	return fmt.Sprintf("%s\n", filtered), nil
}

// parseTitle parses title from a string of sanitised file contents
func parseTitle(inp string, t *string) error {
	lines, err := parseLines(inp)
	if err != nil {
		return err
	}

	tLine := lines[0]
	parts := strings.Split(tLine, titlePrefix)

	title := parts[0]
	if len(parts) == 2 {
		title = parts[1]
	}

	title = strings.ToLower(title)
	*t = strings.Trim(title, " \n")

	return nil
}

// parseTimestamp parses modified date from a string of sanitised file contents
func parseTimestamp(inp string, t *time.Time) error {
	lines, err := parseLines(inp)
	if err != nil {
		return err
	}

	cLine := lines[2]
	mLine := lines[3]

	if !strings.HasPrefix(cLine, createTimePrefix) {
		return errors.New("cannot locate created timestamp")
	}
	if !strings.HasPrefix(mLine, modTimePrefix) {
		return errors.New("cannot locate modified timestamp")
	}

	parts := strings.Split(mLine, modTimePrefix)
	if len(parts) != 2 {
		return errors.New("cannot locate timestamp within modified line")
	}

	ts, err := time.Parse(tsFormat, strings.Trim(parts[1], " \n"))
	if err != nil {
		return err
	}

	loc, err := time.LoadLocation("Europe/London")
	if err != nil {
		return err
	}

	*t = time.Date(ts.Year(), ts.Month(), ts.Day(), ts.Hour(), ts.Minute(), 0, 0, loc)

	return nil
}

// parseNoteContent parses note content from a string of sanitised file contents
func parseNoteContent(inp string, c *string) error {
	lines, err := parseLines(inp)
	if err != nil {
		return err
	}

	contentLines := lines[headerLines-1:]
	content := strings.Join(contentLines, "\n")
	*c = strings.TrimLeft(content, " \n")

	return nil
}

// parseLines parses the provided input into the lines
func parseLines(inp string) ([]string, error) {
	lines := strings.Split(inp, "\n")
	if len(lines) < 6 {
		return nil, errors.New("not enough lines")
	}

	return lines, nil
}
