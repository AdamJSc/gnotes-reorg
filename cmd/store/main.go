package main

import (
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

	flagI := flag.String("i", "", "relative path to directory of cleaned files with manifest")
	flag.Parse()

	inPath, err := fs.ParseDirFlag(flagI)
	if err != nil {
		return fmt.Errorf("cannot parse -i: %w", err)
	}

	manifestPath := strings.Join([]string{inPath, "manifest.json"}, string(os.PathSeparator))
	manifestPath, err = filepath.Abs(manifestPath)
	if err != nil {
		return fmt.Errorf("absolute manifest path failed: %w", err)
	}

	log.Printf("scanning directory: %s", inPath)
	files, err := fs.GetChildPaths(inPath, false)
	if err != nil {
		return err
	}

	var jsonFiles []string
	for _, f := range files {
		if strings.HasSuffix(f, ".json") && filepath.Base(f) != "manifest.json" {
			jsonFiles = append(jsonFiles, f)
		}
	}

	if len(jsonFiles) == 0 {
		return fmt.Errorf("no json files found in parent: %s", inPath)
	}

	log.Printf("%d note files to move", len(jsonFiles))

	log.Println("parsing notes from files...")
	notes, err := domain.ParseNotesFromPaths(jsonFiles)
	if err != nil {
		return fmt.Errorf("cannot parse notes: %w", err)
	}

	log.Println("parsing manifest from file...")
	manifest, err := domain.ParseManifestFromPath(manifestPath)
	if err != nil {
		return fmt.Errorf("cannot parse manifest: %w", err)
	}

	if len(notes) != len(manifest) {
		return fmt.Errorf("mismatch source length: %d notes: %d manifest entries", len(notes), len(manifest))
	}

	log.Printf("%d notes moving to storage", len(notes))
	if !app.Cont() {
		return errors.New("aborted")
	}

	log.Println("begin move to storage...")
	if err := domain.MoveNotesToStorage(notes, manifest); err != nil {
		return fmt.Errorf("moving notes failed: %w", err)
	}

	return nil
}
