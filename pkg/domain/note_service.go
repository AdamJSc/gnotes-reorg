package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
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

// ParseFromDirs parses Notes from the provided directory paths
func (ns *NoteService) ParseFromDirs(ctx context.Context, paths []string) ([]*Note, error) {
	var notes []*Note

	for _, p := range paths {
		contentPath, err := ns.fs.Abs(p, "content.html")
		if err != nil {
			return nil, fmt.Errorf("cannot get absolute file path: %w", err)
		}

		id := ns.fs.Base(p)

		n, err := ns.parseFromFile(ctx, contentPath, id)
		if err != nil {
			return nil, fmt.Errorf("cannot parse note from file %s: %w", p, err)
		}

		notes = append(notes, n)
	}

	return notes, nil
}

// WriteToDir outputs the provided notes to individual files within the provided output path
func (ns *NoteService) WriteToDir(ctx context.Context, notes []*Note, path string) (n int, e error) {
	defer func() {
		if e != nil {
			if err := ns.cleanupPath(path); err != nil {
				e = fmt.Errorf("cleanup failed: %s: original error: %w", err.Error(), e)
			}
		}
	}()

	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	// clean slate
	if err := ns.cleanupPath(path); err != nil {
		return 0, fmt.Errorf("cannot remove path %s: %w", path, err)
	}

	// get path info
	info, err := ns.fs.Stat(path)
	if err != nil {
		// attempt to create directory
		if err := ns.fs.Mkdir(path, 0755); err != nil {
			return 0, fmt.Errorf("cannot make directory %s: %w", path, err)
		}
		// re-retrieve info
		info, err = ns.fs.Stat(path)
		if err != nil {
			return 0, fmt.Errorf("cannot retrieve info for directory %s: %w", path, err)
		}
	}

	// ensure that path is a directory
	if !info.IsDir() {
		return 0, fmt.Errorf("not a directory: %s", path)
	}

	errCh := make(chan error, 1)
	noteCh := make(chan *Note, len(notes))

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
			go func(n *Note, idx int) {
				defer func() {
					<-sem
				}()
				// save note
				filePath, err := ns.generateFilePath(path, n.filename(), "json", idx)
				if err != nil {
					errCh <- fmt.Errorf("cannot generate file path: %w", err)
					return
				}
				buf := bytes.NewBuffer(nil)
				if err := json.NewEncoder(buf).Encode(&n); err != nil {
					errCh <- fmt.Errorf("cannot parse json: %w", err)
					return
				}
				if err := ioutil.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
					errCh <- fmt.Errorf("note %s: %w", n.ID, err)
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

// parseFromFile parses a Note from the provided file path
func (ns *NoteService) parseFromFile(ctx context.Context, path, id string) (*Note, error) {
	b, err := ns.fs.ReadFile(path)
	if err != nil {
		return nil, err
	}

	content := string(b)
	sanitised, err := sanitiseInput(content)
	if err != nil {
		return nil, err
	}

	n := &Note{
		ID:           id,
		OriginalPath: path,
	}

	if err := parseTitle(sanitised, &n.Title); err != nil {
		return nil, err
	}
	if err := parseTimestamp(sanitised, &n.Timestamp); err != nil {
		return nil, err
	}
	if err := parseNoteContent(sanitised, &n.Content); err != nil {
		return nil, err
	}

	return n, nil
}

// cleanupPath removes the provided path and all descendents
func (ns *NoteService) cleanupPath(outPath string) error {
	return ns.fs.RemoveAll(outPath)
}

// generateFilePath generates a file path from the provided arguments
func (ns *NoteService) generateFilePath(dir, base, ext string, i int) (string, error) {
	var suffix string
	if i > 0 {
		suffix = fmt.Sprintf("_%d", i)
	}

	fileName := fmt.Sprintf("%s%s.%s", base, suffix, ext)

	fullPath, err := ns.fs.Abs(dir, fileName)
	if err != nil {
		return "", fmt.Errorf("cannot parse absolute path: %w", err)
	}

	return fullPath, nil
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
