package command

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reorg/pkg/adapters"
	"reorg/pkg/domain"
	"time"
)

// Store represents our store command
type Store struct {
	runner
	InPath string
	Files  *domain.FileSystemService
	Notes  *domain.NoteService
}

// Run implements Runner
func (s *Store) Run() error {
	if err := s.validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	var err error

	s.InPath, err = s.Files.ParseAbsPath(s.InPath)
	if err != nil {
		return fmt.Errorf("cannot parse absolute path %s: %w", s.InPath, err)
	}

	if err := s.Files.DirExists(s.InPath); err != nil {
		return fmt.Errorf("cannot find directory %s: %w", s.InPath, err)
	}

	log.Printf("scanning directory: %s", s.InPath)

	files, err := s.Files.GetChildPaths(
		s.InPath,
		&domain.IsNotDir{},
		&domain.IsJSON{},
		&domain.IsNotName{BaseNames: []string{manifestFileName}},
	)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no json files found in parent: %s", s.InPath)
	}

	log.Println("parsing notes from files...")

	notes, err := s.Notes.ParseFromFiles(files)
	if err != nil {
		return fmt.Errorf("cannot parse notes: %w", err)
	}

	log.Println("parsing manifest from file...")

	manifestPath, err := s.Files.ParseAbsPath(s.InPath, manifestFileName)
	if err != nil {
		return fmt.Errorf("cannot parse manifest path: %w", err)
	}

	manifest, err := s.Notes.ParseManifestFromPath(manifestPath)
	if err != nil {
		return fmt.Errorf("cannot parse manifest: %w", err)
	}

	ml := manifest.Len()

	if ml == 0 {
		return errors.New("manifest is empty :'(")
	}

	if len(notes) != ml {
		log.Printf("WARNING: mismatched source length: %d notes: %d manifest entries", len(notes), ml)
		if !cont() {
			return errors.New("aborted")
		}
	}

	notes = s.Notes.FilterNotesByManifest(notes, manifest, true)

	log.Printf("%d notes moving to storage", len(notes))

	if !cont() {
		return errors.New("aborted")
	}

	// TODO: flag to determine instance of writer
	wr := &adapters.GoogleStorageNoteWriter{}

	log.Println("begin move to storage...")

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()

	if _, err := s.Notes.WriteNotes(ctx, notes, wr); err != nil {
		return fmt.Errorf("moving notes failed: %w", err)
	}

	return nil
}

// validate sanity checks the input variables
func (s *Store) validate() error {
	if s.InPath == "" {
		return errors.New("-i is empty")
	}

	return nil
}
