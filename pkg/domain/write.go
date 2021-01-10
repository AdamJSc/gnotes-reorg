package domain

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"sync"
)

// WriteNotes handles the output of our notes to individual files within the output path
func WriteNotes(ctx context.Context, notes []Note, outPath string) (e error) {
	defer func() {
		if e != nil {
			if err := cleanupPath(outPath); err != nil {
				e = fmt.Errorf("cleanup failed: %w", err)
			}
		}
	}()

	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	// clean slate
	if err := cleanupPath(outPath); err != nil {
		return fmt.Errorf("cannot remove path %s: %w", outPath, err)
	}

	// get path info
	info, err := os.Stat(outPath)
	if err != nil {
		// attempt to create directory
		if err := os.Mkdir(outPath, 0755); err != nil {
			return fmt.Errorf("cannot make directory %s: %w", outPath, err)
		}
		// re-retrieve info
		info, err = os.Stat(outPath)
		if err != nil {
			return fmt.Errorf("cannot retrieve info for directory %s: %w", outPath, err)
		}
	}

	// ensure that path is a directory
	if !info.IsDir() {
		return fmt.Errorf("not a directory: %s", outPath)
	}

	errCh := make(chan error, 1)
	doneCh := make(chan struct{}, 1)

	go func() {
		sem := make(chan struct{}, 50) // ensure no more than 50 concurrent operations
		wg := &sync.WaitGroup{}
		wg.Add(len(notes))

		for _, n := range notes {
			// check whether operation has ended
			select {
			case <-ctxWithCancel.Done():
				return
			default:
			}
			// otherwise continue for current note
			sem <- struct{}{}
			go func(n Note) {
				defer func() {
					<-sem
					wg.Done()
				}()

				// save note
				filePath := strings.Join([]string{outPath, n.filename(".json")}, string(os.PathSeparator))
				buf := bytes.NewBuffer(nil)
				if err := json.NewEncoder(buf).Encode(&n); err != nil {
					errCh <- fmt.Errorf("cannot parse json: %w", err)
					return
				}
				if err := ioutil.WriteFile(filePath, buf.Bytes(), 0644); err != nil {
					errCh <- fmt.Errorf("note %s: %w", n.ID, err)
					return
				}
			}(n)
		}

		wg.Wait()
		doneCh <- struct{}{}
	}()

	select {
	case err := <-errCh:
		cancel()
		return fmt.Errorf("failed to create note: %w", err)
	case <-doneCh:
		return nil
	}
}

// cleanupPath removes the provided path and all descendents
func cleanupPath(outPath string) error {
	return os.RemoveAll(outPath)
}
