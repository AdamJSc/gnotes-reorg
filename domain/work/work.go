package work

import (
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"sync"
	"time"
)

// note represents a single note
type note struct {
	id      string
	path    string
	date    time.Time
	content string
}

// Do does our grunt work
func Do(inpPaths []string, outPath string) error {
	var notes []note

	// gather notes with id and path
	for _, p := range inpPaths {
		root := strings.Split(p, "/")
		notes = append(notes, note{
			id:   root[len(root)-1],
			path: p,
		})
	}

	// enrich notes with content and date
	notes, err := enrichNotes(notes)
	if err != nil {
		return err
	}

	// output
	total := len(notes)
	for i, n := range notes {
		log.Printf("note #%d/%d: %+v\n", i, total, n)
	}

	log.Printf("finished summarising %d notes\n", total)

	return nil
}

// enrichNotes parses contents of each note's path and returns the note objects with this data attached
func enrichNotes(notes []note) ([]note, error) {
	var (
		errCh    = make(chan error, 1)
		resultCh = make(chan []note, 1)
	)

	go func() {
		var enriched []note
		noteCh := make(chan note, len(notes))

		wg := &sync.WaitGroup{}
		wg.Add(len(notes))

		for _, n := range notes {
			go func(n note) {
				defer func() {
					wg.Done()
				}()
				if err := enrichNote(&n); err != nil {
					errCh <- fmt.Errorf("note %s: %w", n.id, err)
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
func enrichNote(n *note) error {
	b, err := ioutil.ReadFile(n.path)
	if err != nil {
		return err
	}

	// TODO: sanitise content
	n.content = string(b)

	return nil
}

// getContentPath returns the name of the content file within the provided dir path
func getContentPath(dir string) string {
	return fmt.Sprintf("%s/content.html", dir)
}
