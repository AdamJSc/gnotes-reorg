package domain

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"
	"sync"
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

// ParseNotes handles the parsing of our input paths into Notes
func ParseNotes(inpPaths []string, outPath string) ([]Note, error) {
	var notes []Note

	// gather notes with id and path
	for _, p := range inpPaths {
		root := strings.Split(p, "/")
		notes = append(notes, Note{
			ID:   root[len(root)-1],
			Path: getContentPath(p),
		})
	}

	// enrich notes with content and date
	notes, err := enrichNotes(notes)
	if err != nil {
		return nil, err
	}

	return notes, nil
}

// getContentPath returns the name of the content file within the provided dir path
func getContentPath(dir string) string {
	return fmt.Sprintf("%s/content.html", dir)
}

// enrichNotes parses contents of each note's path and returns the note objects with this data attached
func enrichNotes(notes []Note) ([]Note, error) {
	var (
		errCh    = make(chan error, 1)
		resultCh = make(chan []Note, 1)
	)

	go func() {
		var enriched []Note
		noteCh := make(chan Note, len(notes))

		wg := &sync.WaitGroup{}
		wg.Add(len(notes))

		for _, n := range notes {
			go func(n Note) {
				defer func() {
					wg.Done()
				}()
				if err := enrichNote(&n); err != nil {
					errCh <- fmt.Errorf("note %s: %w", n.ID, err)
					return
				}
				noteCh <- n
			}(n)
		}

		wg.Wait()
		close(noteCh)

		for n := range noteCh {
			enriched = append(enriched, n)
		}
		resultCh <- enriched
	}()

	for {
		select {
		case err := <-errCh:
			return nil, err
		case result := <-resultCh:
			return result, nil
		}
	}
}

// enrichNote enriches the provided note
func enrichNote(n *Note) error {
	b, err := ioutil.ReadFile(n.Path)
	if err != nil {
		return err
	}

	content := string(b)
	sanitised, err := sanitiseInput(content)
	if err != nil {
		return err
	}

	if err := parseTitle(sanitised, &n.Title); err != nil {
		return err
	}

	if err := parseTimestamp(sanitised, &n.Timestamp); err != nil {
		return err
	}

	if err := parseNoteContent(sanitised, &n.Content); err != nil {
		return err
	}

	return nil
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
