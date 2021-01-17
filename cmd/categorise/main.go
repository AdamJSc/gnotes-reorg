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

	var inPath string
	if err := parseFlag(&inPath); err != nil {
		return fmt.Errorf("cannot parse flag: %w", err)
	}

	ctx := context.Background()

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

	log.Println("removing notes already processed...")
	notes = domain.FilterNotesByManifest(notes, manifest, false)

	log.Println("sorting notes...")
	notes = domain.SortNotesByFilenameDesc(notes)

	log.Printf("%d note files to categorise", len(notes))

	if !app.Cont() {
		return errors.New("aborted")
	}

	log.Println("begin requesting categories...")
	if err := domain.RequestCategories(notes, manifest, manifestPath); err != nil {
		return fmt.Errorf("cannot request categories: %w", err)
	}

	return nil
}

// parseFlag parses and sanity checks the required flag
func parseFlag(i *string) error {
	flagI := flag.String("i", "", "relative path to directory of cleaned files")
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
