package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"regexp"
	"sort"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

const (
	titlePrefix      = "Back"
	createTimePrefix = "Create Time: "
	modTimePrefix    = "Modify Time: "
	tsFormat         = "02/01/2006 15:04"
	headerLines      = 5
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

// ParseFromFile parses a Note from the provided source file path
func (ns *NoteService) ParseFromFile(path string) (Note, error) {
	payload, err := ns.fs.ReadFile(path)
	if err != nil {
		return Note{}, fmt.Errorf("cannot parse file %s as note: %w", path, err)
	}

	// parse json
	r := bytes.NewReader(payload)
	var n Note
	if err := json.NewDecoder(r).Decode(&n); err != nil {
		return Note{}, fmt.Errorf("cannot json decode payload in file %s as note: %w", path, err)
	}

	n.ParentDir = path

	return n, nil
}

// ParseManifestFromPath returns a noteManifest parsed from the provided path
func (ns *NoteService) ParseManifestFromPath(path string) (NoteManifest, error) {
	m := NoteManifest{path: path}

	// read contents of manifest file
	payload, err := ns.fs.ReadFile(path)
	if err != nil {
		if !ns.fs.IsNotExist(err) {
			return NoteManifest{}, fmt.Errorf("cannot parse existing manifest file %s: %w", path, err)
		}
		return m, nil
	}

	// check if file contents are empty
	if len(payload) == 0 {
		return m, nil
	}

	// parse manifest
	if err := json.Unmarshal(payload, &m); err != nil {
		return NoteManifest{}, fmt.Errorf("cannot json decode payload at %s as manifest: %w", path, err)
	}

	return m, nil
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

// FilterNotesByManifest returns the provided Notes based on the provided manifest
//
// If keepIfPresent is true, Notes will be retained that are present in the manifest,
// otherwise they will be retained if they are not present in the manifest.
func (ns *NoteService) FilterNotesByManifest(notes []Note, m NoteManifest, keepIfPresent bool) []Note {
	var retained []Note

	for _, n := range notes {
		if m.IsSet(n.Filename()) == keepIfPresent {
			retained = append(retained, n)
		}
	}

	return retained
}

// SortNotesByFilenameDesc sorts the provided notes ordered descending by filename
func (ns *NoteService) SortNotesByFilenameDesc(notes []Note) []Note {
	sort.SliceStable(notes, func(i, j int) bool {
		n1 := notes[i]
		n2 := notes[j]
		return strings.Compare(n1.Filename(), n2.Filename()) > 0
	})

	return notes
}

// SaveManifest saves the provided manifest
func (ns *NoteService) SaveManifest(m NoteManifest) error {
	b, err := json.Marshal(&m)
	if err != nil {
		return fmt.Errorf("cannot json encode manifest: %w", err)
	}

	if err := ns.fs.WriteFile(m.path, b, 0644); err != nil {
		return fmt.Errorf("cannot write to file %s: %w", m.path, err)
	}

	return nil
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
