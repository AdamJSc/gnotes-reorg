package domain

import (
	"context"
	"errors"
	"fmt"
	"log"
)

// Storage defines our Store method
type Storage interface {
	Store(ctx context.Context, n Note) error
}

// StubStorage implements Storage as a stub
type StubStorage struct{ Storage }

// Store implements Store method on Storage interface
func (s *StubStorage) Store(ctx context.Context, n Note) error {
	log.Printf("processing %s...", n.filename())
	// TODO: implement me
	return nil
}

func MoveNotesToStorage(notes []Note, manifest noteManifest, store Storage) error {
	log.Println("moving files to storage...")

	errCh := make(chan error, 1)
	noteCh := make(chan Note, len(notes))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		sem := make(chan struct{}, 50) // allow max. 50 concurrent operations

		for _, n := range notes {
			sem <- struct{}{} // block

			go func(n Note) {
				defer func() {
					<-sem // unblock
				}()

				select {
				case <-ctx.Done():
					// show's over...
					return
				default:
				}

				// ok to process note...

				if err := populateNoteCategory(&n, manifest); err != nil {
					errCh <- fmt.Errorf("error populating category for note %s: %w", n.filename(), err)
					return
				}

				if err := store.Store(ctx, n); err != nil {
					errCh <- fmt.Errorf("error moving note %s to storage: %w", n.filename(), err)
					return
				}

				noteCh <- n
			}(n)
		}
	}()

	total := len(notes)
	count := 0

	for {
		select {
		case n := <-noteCh:
			count++
			log.Printf("success #%d/%d: %s", count, total, n.filename())
			if count == total {
				return nil
			}
		case err := <-errCh:
			cancel()
			return fmt.Errorf("operation cancelled: %w", err)
		}
	}
}

// populateNoteCategory adds a Category value to the provided Note
func populateNoteCategory(n *Note, manifest noteManifest) error {
	f := n.filename()
	cat, ok := manifest[f]
	if !ok {
		return errors.New("category not found")
	}
	n.Category = cat
	return nil
}
