package work

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

// note represents a single note
type note struct {
	date    time.Time
	content string
}

// Do does our grunt work
func Do(inpPaths []string, outPath string) error {
	// result channels
	noteCh := make(chan *note, len(inpPaths))
	errCh := make(chan error, len(inpPaths))

	// concurrency objects
	sem := make(chan struct{}, 50)
	wg := &sync.WaitGroup{}

	// parse notes
	for i, path := range inpPaths {
		fpath := fmt.Sprintf("%s/content.html", path)
		wg.Add(1)
		sem <- struct{}{} // block
		go func(i int, fpath string) {
			defer func() {
				wg.Done()
				<-sem // unblock
			}()
			info, err := os.Stat(fpath)
			if err != nil {
				// cannot get file info
				errCh <- fmt.Errorf("inpPath #%d: %s", i, err.Error())
				return
			}
			// parse note
			n, err := parseNoteFromFile(fpath)
			if err != nil {
				errCh <- fmt.Errorf("error in %s: %s", info.Name(), err.Error())
				return
			}
			// send note to channel
			noteCh <- n
		}(i, fpath)
	}

	wg.Wait()
	close(noteCh)
	close(errCh)

	// output channel contents
	noteCount := 0
	noteTotal := len(noteCh)
	for n := range noteCh {
		noteCount++
		log.Printf("note #%d/%d: %+v\n", noteCount, noteTotal, n)
	}
	errCount := 0
	errTotal := len(errCh)
	for err := range errCh {
		errCount++
		log.Printf("error #%d/%d: %+v\n", errCount, errTotal, err)
	}
	log.Printf("finished summarising %d notes and %d errors\n", noteTotal, errTotal)

	return nil
}

// parseNoteFromFile parses a note from the given file path
func parseNoteFromFile(path string) (*note, error) {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	// TODO: parse properly, just add full content of file as a placeholder for now
	return &note{
		content: string(b),
	}, nil
}
