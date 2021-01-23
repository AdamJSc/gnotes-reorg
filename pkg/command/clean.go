package command

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"reorg/pkg/domain"
	"time"
)

// Clean represents out clean command
type Clean struct {
	Files *domain.FileSystemService
	Notes *domain.NoteService
	runner
	inPath  string
	outPath string
}

// Run implements Runner
func (c *Clean) Run() error {
	if err := c.parseFlags(); err != nil {
		return fmt.Errorf("cannot parse flags: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(5)*time.Second)
	defer cancel()

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
		return fmt.Errorf("cannot find %s: %w", c.inPath, err)
	}

	if err := c.Files.DirExists(c.outPath); err != nil {
		if err := c.Files.MakeDir(c.outPath); err != nil {
			return fmt.Errorf("cannot create directory %s: %w", c.outPath, err)
		}
	}

	log.Printf("scanning directory: %s", c.inPath)

	dirs, err := c.Files.GetChildPaths(c.inPath, &domain.IsDir{})
	if err != nil {
		return err
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

	log.Printf("parsed %d notes\n", len(notes))
	log.Printf("writing to directory: %s", c.outPath)
	log.Println("this will reset its existing contents")

	if !cont() {
		return errors.New("aborted")
	}

	n, err := c.Notes.WriteToDir(ctx, notes, c.outPath)
	if err != nil {
		return err
	}

	log.Printf("finished writing %d notes\n", n)

	return nil
}

// parseFlags parses and sanity checks the required flags
func (c *Clean) parseFlags() error {
	flagI := flag.String("i", "", "relative path to gnotes export directory")
	flagO := flag.String("o", "", "relative path to output directory for cleaned notes")
	flag.Parse()

	if flagI == nil {
		return errors.New("-i is missing")
	}
	if flagO == nil {
		return errors.New("-o is missing")
	}

	valI := *flagI
	valO := *flagO

	if valI == "" {
		return errors.New("-i is empty")
	}
	if valO == "" {
		return errors.New("-o is empty")
	}

	c.inPath = valI
	c.outPath = valO

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
