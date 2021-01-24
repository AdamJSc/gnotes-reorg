package command

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"reorg/pkg/adapters"
	"reorg/pkg/domain"
	"time"
)

// Store represents our store command
type Store struct {
	runner
	Files  *domain.FileSystemService
	Notes  *domain.NoteService
	inPath string
}

// Run implements Runner
func (s *Store) Run() error {
	if err := s.parseFlag(); err != nil {
		return fmt.Errorf("cannot parse flag: %w", err)
	}

	var err error

	s.inPath, err = s.Files.ParseAbsPath(s.inPath)
	if err != nil {
		return fmt.Errorf("cannot parse absolute path %s: %w", s.inPath, err)
	}

	if err := s.Files.DirExists(s.inPath); err != nil {
		return fmt.Errorf("cannot find directory %s: %w", s.inPath, err)
	}

	log.Printf("scanning directory: %s", s.inPath)

	files, err := s.Files.GetChildPaths(
		s.inPath,
		&domain.IsNotDir{},
		&domain.IsJSON{},
		&domain.IsNotName{BaseNames: []string{manifestFileName}},
	)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no json files found in parent: %s", s.inPath)
	}

	log.Println("parsing notes from files...")

	notes, err := s.Notes.ParseFromFiles(files)
	if err != nil {
		return fmt.Errorf("cannot parse notes: %w", err)
	}

	log.Println("parsing manifest from file...")

	manifestPath, err := s.Files.ParseAbsPath(s.inPath, manifestFileName)
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

// parseFlag parses and sanity checks the required flag
func (s *Store) parseFlag() error {
	i := flag.String("i", "", "relative path to directory of cleaned files with manifest")
	flag.Parse()

	s.inPath = *i

	if s.inPath == "" {
		return errors.New("-i is empty")
	}

	return nil
}
