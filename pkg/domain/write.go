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
func WriteNotes(ctx context.Context, notes []Note, outPath string) (n int, e error) {
	defer func() {
		if e != nil {
			if err := cleanupPath(outPath); err != nil {
				e = fmt.Errorf("cleanup failed: %s: original error: %w", err.Error(), e)
			}
		}
	}()

	ctxWithCancel, cancel := context.WithCancel(ctx)
	defer cancel()

	// clean slate
	if err := cleanupPath(outPath); err != nil {
		return 0, fmt.Errorf("cannot remove path %s: %w", outPath, err)
	}

	// get path info
	info, err := os.Stat(outPath)
	if err != nil {
		// attempt to create directory
		if err := os.Mkdir(outPath, 0755); err != nil {
			return 0, fmt.Errorf("cannot make directory %s: %w", outPath, err)
		}
		// re-retrieve info
		info, err = os.Stat(outPath)
		if err != nil {
			return 0, fmt.Errorf("cannot retrieve info for directory %s: %w", outPath, err)
		}
	}

	// ensure that path is a directory
	if !info.IsDir() {
		return 0, fmt.Errorf("not a directory: %s", outPath)
	}

	errCh := make(chan error, 1)
	doneCh := make(chan struct{}, 1)

	count := 0
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
					wg.Done()
					<-sem
				}()
				// save note
				filePath, err := generateUniqueFilePath(outPath, n.filename(), "json", 0)
				if err != nil {
					errCh <- fmt.Errorf("cannot generate unique file path: %w", err)
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

				count++
			}(n)
		}

		wg.Wait()
		doneCh <- struct{}{}
	}()

	select {
	case err := <-errCh:
		cancel()
		return 0, fmt.Errorf("failed to create note: %w", err)
	case <-doneCh:
		return count, nil
	}
}

// cleanupPath removes the provided path and all descendents
func cleanupPath(outPath string) error {
	return os.RemoveAll(outPath)
}

// generateUniqueFilePath generates a unique file path from the provided arguments
func generateUniqueFilePath(dir, base, ext string, i int) (string, error) {
	if i > 50 {
		return "", fmt.Errorf("cannot increment %d times", i)
	}

	var suffix string
	if i > 0 {
		suffix = fmt.Sprintf("_%d", i)
	}

	fileName := fmt.Sprintf("%s%s.%s", base, suffix, ext)
	fullPath := strings.Join([]string{dir, fileName}, string(os.PathSeparator))

	_, err := os.Stat(fullPath)

	switch {
	case err == nil:
		// file already exists, increment suffix and try again
		return generateUniqueFilePath(dir, base, ext, i+1)
	case !os.IsNotExist(err):
		// something else went wrong
		return "", err
	default:
		return fullPath, nil
	}
}
