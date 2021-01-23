package domain

import (
	"context"
	"errors"
	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

// NoteService provides note-related functionality
type NoteService struct {
	fs FileSystem
}

// ParseFromRawFile parses a Note from raw source at the provided file path
func (ns *NoteService) ParseFromRawFile(path, id string) (Note, error) {
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

// WriteNotes writes the provided notes using the provided NoteWriter
func (ns *NoteService) WriteNotes(ctx context.Context, notes []Note, nw NoteWriter) (int, error) {
	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	errCh := make(chan error, 1)
	noteCh := make(chan Note, len(notes))

	go func() {
		sem := make(chan struct{}, 50) // ensure no more than 50 concurrent operations
		for idx, n := range notes {
			// check whether operation has ended
			select {
			case <-ctxWithCancel.Done():
				return
			default:
			}

			// otherwise continue for current note
			sem <- struct{}{}
			go func(n Note, idx int) {
				defer func() {
					<-sem
				}()

				if err := nw.Write(noteWithIndex{note: n, idx: idx}); err != nil {
					errCh <- fmt.Errorf("error writing note %d: %w", idx, err)
					return
				}

				noteCh <- n
			}(n, idx)
		}
	}()

	total := len(notes)
	count := 0

	for {
		select {
		case err := <-errCh:
			cancel()
			return 0, fmt.Errorf("failed to create note: %w", err)
		case <-noteCh:
			count++
			log.Printf("written note %d/%d", count, total)
			if count == total {
				return count, nil
			}
		}
	}
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
