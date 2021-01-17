package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"reorg/pkg/app"
	"reorg/pkg/domain"
)

func main() {
	fs := domain.NewFileSystemService(&domain.OsFileSystem{})

	if err := run(fs); err != nil {
		log.Fatalf("failed: %s", err.Error())
	}

	log.Println("process complete!")
}

// run executes our business logic
func run(fs *domain.FileSystemService) error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	var inPath, outPath string
	if err := parseFlags(&inPath, &outPath); err != nil {
		return fmt.Errorf("cannot parse flags: %w", err)
	}

	if err := fs.DirExists(inPath); err != nil {
		return fmt.Errorf("cannot find %s: %w", inPath, err)
	}

	if err := fs.ParseAbsPath(&inPath, inPath, "Other"); err != nil {
		return fmt.Errorf("cannot parse absolute path: %w", err)
	}

	if err := fs.ParseAbsPath(&outPath, outPath, "output"); err != nil {
		return fmt.Errorf("cannot parse absolute path: %w", err)
	}

	log.Printf("scanning directory: %s", inPath)
	dirs, err := fs.GetChildPaths(inPath, &domain.IsDir{})
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

// parseFlags parses and sanity checks the required flags
func parseFlags(i, o *string) error {
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

	*i = valI
	*o = valO

	return nil
}
