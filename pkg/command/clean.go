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

// Clean represents our clean command
type Clean struct {
	runner
	Files   *domain.FileSystemService
	Notes   *domain.NoteService
	inPath  string
	outPath string
	jsonOut bool
	txtOut  bool
}

// Run implements Runner
func (c *Clean) Run() error {
	if err := c.parseFlags(); err != nil {
		return fmt.Errorf("cannot parse flags: %w", err)
	}

	var err error

	c.inPath, err = c.Files.ParseAbsPath(c.inPath, "Other")
	if err != nil {
		return fmt.Errorf("cannot parse absolute path %s: %w", c.inPath, err)
	}

	c.outPath, err = c.Files.ParseAbsPath(c.outPath)
	if err != nil {
		return fmt.Errorf("cannot parse absolute path %s: %w", c.outPath, err)
	}

	if err := c.Files.DirExists(c.inPath); err != nil {
		return fmt.Errorf("cannot find directory %s: %w", c.inPath, err)
	}

	log.Printf("scanning directory: %s", c.inPath)

	dirs, err := c.Files.GetChildPaths(c.inPath, &domain.IsDir{})
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if len(dirs) == 0 {
		return fmt.Errorf("no directories found in parent: %s", c.inPath)
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

	notes = enrichNotesWithParentDir(notes, c.outPath)

	log.Printf("parsed %d notes\n", len(notes))
	log.Printf("writing to directory: %s", c.outPath)
	log.Println("this will reset its existing contents")

	if !cont() {
		return errors.New("aborted")
	}

	// clear output directory
	if err := c.Files.RemoveAll(c.outPath); err != nil {
		return fmt.Errorf("cannot remove directory %s: %w", c.outPath, err)
	}

	// recreate output directory
	if err := c.Files.MakeDir(c.outPath); err != nil {
		return fmt.Errorf("cannot create directory %s: %w", c.outPath, err)
	}

	var wr domain.NoteWriter
	switch {
	case c.jsonOut:
		wr = &adapters.JSONNoteWriter{Files: c.Files}
	case c.txtOut:
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

// parseFlags parses and sanity checks the required flags
func (c *Clean) parseFlags() error {
	i := flag.String("i", "", "relative path to gnotes export directory")
	o := flag.String("o", "", "relative path to output directory for cleaned notes")
	j := flag.Bool("json", false, "output cleaned notes as json files")
	t := flag.Bool("txt", false, "output cleaned notes as txt files")
	flag.Parse()

	c.inPath = *i
	c.outPath = *o
	c.jsonOut = *j
	c.txtOut = *t

	if c.inPath == "" {
		return errors.New("-i is empty")
	}
	if c.outPath == "" {
		return errors.New("-o is empty")
	}
	if c.txtOut == c.jsonOut {
		return errors.New("please specify either -json or -txt")
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
