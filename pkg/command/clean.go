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

// Clean represents our clean command
type Clean struct {
	runner
	InPath  string
	OutPath string
	JSONOut bool
	TxtOut  bool
	Files   *domain.FileSystemService
	Notes   *domain.NoteService
}

// Run implements Runner
func (c *Clean) Run() error {
	if err := c.validate(); err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	var err error

	c.InPath, err = c.Files.ParseAbsPath(c.InPath, "Other")
	if err != nil {
		return fmt.Errorf("cannot parse absolute path %s: %w", c.InPath, err)
	}

	c.OutPath, err = c.Files.ParseAbsPath(c.OutPath)
	if err != nil {
		return fmt.Errorf("cannot parse absolute path %s: %w", c.OutPath, err)
	}

	if err := c.Files.DirExists(c.InPath); err != nil {
		return fmt.Errorf("cannot find directory %s: %w", c.InPath, err)
	}

	log.Printf("scanning directory: %s", c.InPath)

	dirs, err := c.Files.GetChildPaths(c.InPath, &domain.IsDir{})
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if len(dirs) == 0 {
		return fmt.Errorf("no directories found in parent: %s", c.InPath)
	}

	log.Printf("%d directories to search for note files", len(dirs))

	if !cont() {
		return errors.New("aborted")
	}

	log.Println("parsing notes...")

	notes, err := c.parseNotesFromRawDirs(dirs)
	if err != nil {
		return err
	}

	notes = enrichNotesWithParentDir(notes, c.OutPath)

	log.Printf("parsed %d notes\n", len(notes))
	log.Printf("writing to directory: %s", c.OutPath)
	log.Println("this will reset its existing contents")

	if !cont() {
		return errors.New("aborted")
	}

	// clear output directory
	if err := c.Files.RemoveAll(c.OutPath); err != nil {
		return fmt.Errorf("cannot remove directory %s: %w", c.OutPath, err)
	}

	// recreate output directory
	if err := c.Files.MakeDir(c.OutPath); err != nil {
		return fmt.Errorf("cannot create directory %s: %w", c.OutPath, err)
	}

	var wr domain.NoteWriter
	switch {
	case c.JSONOut:
		wr = &adapters.JSONNoteWriter{Files: c.Files}
	case c.TxtOut:
		wr = &adapters.TxtNoteWriter{Files: c.Files}
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()

	n, err := c.Notes.WriteNotes(ctx, notes, wr)
	if err != nil {
		return err
	}

	log.Printf("finished writing %d notes\n", n)

	return nil
}

// validate sanity checks the input variables
func (c *Clean) validate() error {
	if c.InPath == "" {
		return errors.New("input path is empty")
	}
	if c.OutPath == "" {
		return errors.New("output path is empty")
	}
	if c.TxtOut == c.JSONOut {
		return errors.New("must specify output either json or txt")
	}

	return nil
}

// parseNotesFromRawDirs parses Notes from the provided raw directory paths
func (c *Clean) parseNotesFromRawDirs(paths []string) ([]domain.Note, error) {
	var notes []domain.Note

	for _, p := range paths {
		pathToRaw, err := c.Files.ParseAbsPath(p, "content.html")
		if err != nil {
			return nil, fmt.Errorf("cannot get absolute file path: %w", err)
		}

		// id is final directory name of provided directory path
		id := c.Files.ParseBase(p)

		n, err := c.Notes.ParseFromRawFile(pathToRaw, id)
		if err != nil {
			return nil, fmt.Errorf("cannot parse note from file %s: %w", p, err)
		}

		notes = append(notes, n)
	}

	return notes, nil
}

// enrichNotesWithParentDir enriches the provided Notes with the provided parent directory value
func enrichNotesWithParentDir(notes []domain.Note, parentDir string) []domain.Note {
	var enriched []domain.Note

	for _, n := range notes {
		n.ParentDir = parentDir
		enriched = append(enriched, n)
	}

	return enriched
}
