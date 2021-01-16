package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reorg/pkg/app"
	"reorg/pkg/domain"
	"reorg/pkg/fs"
	"strings"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("failed: %s", err.Error())
	}
	log.Println("process complete!")
}

// run executes our business logic
func run() error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flagI := flag.String("i", "", "relative path to gnotes export directory")
	flagO := flag.String("o", "", "relative path to output directory for cleaned notes")
	flag.Parse()

	inPath, err := fs.ParseDirFlag(flagI)
	if err != nil {
		return fmt.Errorf("cannot parse -i: %w", err)
	}

	inPath = strings.Join([]string{inPath, "Other"}, string(os.PathSeparator))
	if err := fs.DirExists(inPath); err != nil {
		return fmt.Errorf("input path: %w", err)
	}

	outPath, err := fs.ParseDirFlag(flagO)
	if err != nil {
		return fmt.Errorf("cannot parse -o: %w", err)
	}

	outPath = strings.Join([]string{outPath, "output"}, string(os.PathSeparator))
	outPath, err = filepath.Abs(outPath)
	if err != nil {
		return fmt.Errorf("absolute out path failed: %w", err)
	}

	log.Printf("scanning directory: %s", inPath)
	dirs, err := fs.GetChildPaths(inPath, true)
	if err != nil {
		return err
	}

	if len(dirs) == 0 {
		return fmt.Errorf("no directories found in parent: %s", inPath)
	}

	log.Printf("%d directories to search for note files", len(dirs))

	if !app.Cont() {
		return errors.New("aborted")
	}

	log.Println("parsing notes...")

	notes, err := domain.ParseRawNotes(dirs, outPath)
	if err != nil {
		return err
	}

	log.Printf("parsed %d notes\n", len(notes))
	log.Printf("writing to directory: %s", outPath)
	log.Println("this will reset its existing contents")

	if !app.Cont() {
		return errors.New("aborted")
	}

	n, err := domain.WriteNotes(context.Background(), notes, outPath)
	if err != nil {
		return err
	}

	log.Printf("finished writing %d notes\n", n)

	return nil
}
