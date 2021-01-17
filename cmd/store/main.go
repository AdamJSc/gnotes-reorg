package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"reorg/pkg/app"
	"reorg/pkg/domain"
	"time"
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

	var inPath string
	if err := parseFlag(&inPath); err != nil {
		return fmt.Errorf("cannot parse flag: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(15)*time.Second)
	defer cancel()

	if err := fs.DirExists(ctx, inPath); err != nil {
		return fmt.Errorf("cannot find %s: %w", inPath, err)
	}

	var manifestPath string
	if err := fs.ParseAbsPath(ctx, &manifestPath, inPath, "manifest.json"); err != nil {
		return fmt.Errorf("cannot parse absolute path: %w", err)
	}

	log.Printf("scanning directory: %s", inPath)
	files, err := fs.GetChildPaths(
		ctx,
		inPath,
		&domain.IsNotDir{},
		&domain.IsJSON{},
		&domain.IsNotName{BaseNames: []string{"manifest.json"}},
	)
	if err != nil {
		return err
	}

	if len(files) == 0 {
		return fmt.Errorf("no json files found in parent: %s", inPath)
	}

	log.Printf("%d note files to move", len(files))

	log.Println("parsing notes from files...")
	notes, err := domain.ParseNotesFromPaths(files)
	if err != nil {
		return fmt.Errorf("cannot parse notes: %w", err)
	}

	log.Println("parsing manifest from file...")
	manifest, err := domain.ParseManifestFromPath(manifestPath)
	if err != nil {
		return fmt.Errorf("cannot parse manifest: %w", err)
	}

	if len(manifest) == 0 {
		return errors.New("manifest is empty :'(")
	}

	if len(notes) != len(manifest) {
		log.Printf("WARNING: mismatched source length: %d notes: %d manifest entries", len(notes), len(manifest))
		if !app.Cont() {
			return errors.New("aborted")
		}
	}

	notes = domain.FilterNotesByManifest(notes, manifest, true)

	log.Printf("%d notes moving to storage", len(notes))
	if !app.Cont() {
		return errors.New("aborted")
	}

	store := &domain.StubStorage{}

	log.Println("begin move to storage...")
	if err := domain.MoveNotesToStorage(notes, manifest, store); err != nil {
		return fmt.Errorf("moving notes failed: %w", err)
	}

	return nil
}

// parseFlag parses and sanity checks the required flag
func parseFlag(i *string) error {
	flagI := flag.String("i", "", "relative path to directory of cleaned files with manifest")
	flag.Parse()

	if flagI == nil {
		return errors.New("-i is missing")
	}

	valI := *flagI

	if valI == "" {
		return errors.New("-i is empty")
	}

	*i = valI

	return nil
}
