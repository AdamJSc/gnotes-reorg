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

	c.inPath, err = c.Files.ParseAbsPath(ctx, c.inPath, "Other")
	if err != nil {
		return fmt.Errorf("cannot parse absolute path %s: %w", c.inPath, err)
	}

	c.outPath, err = c.Files.ParseAbsPath(ctx, c.outPath)
	if err != nil {
		return fmt.Errorf("cannot parse absolute path %s: %w", c.outPath, err)
	}

	if err := c.Files.DirExists(ctx, c.inPath); err != nil {
		return fmt.Errorf("cannot find %s: %w", c.inPath, err)
	}

	if err := c.Files.DirExists(ctx, c.outPath); err != nil {
		if err := c.Files.MakeDir(ctx, c.outPath); err != nil {
			return fmt.Errorf("cannot create directory %s: %w", c.outPath, err)
		}
	}

	log.Printf("scanning directory: %s", c.inPath)

	dirs, err := c.Files.GetChildPaths(ctx, c.inPath, &domain.IsDir{})
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

	notes, err := c.Notes.ParseFromDirs(ctx, dirs)
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
