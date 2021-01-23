package command

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"reorg/pkg/domain"
)

// Clean represents out clean command
type Categorise struct {
	Files *domain.FileSystemService
	Notes *domain.NoteService
	runner
	inPath string
}

// Run implements Runner
func (c *Categorise) Run() error {
	if err := c.parseFlag(); err != nil {
		return fmt.Errorf("cannot parse flag: %w", err)
	}

	if err := c.Files.DirExists(c.inPath); err != nil {
		return fmt.Errorf("cannot find directory %s: %w", c.inPath, err)
	}

	manifestPath, err := c.Files.ParseAbsPath(c.inPath, "manifest.json")
	if err != nil {
		return fmt.Errorf("cannot parse manifest path: %w", err)
	}

	log.Printf("scanning directory: %s", c.inPath)

	files, err := c.Files.GetChildPaths(
		c.inPath,
		&domain.IsNotDir{},
		&domain.IsJSON{},
		&domain.IsNotName{BaseNames: []string{"manifest.json"}},
	)
	if err != nil {
		return fmt.Errorf("validation error: %w", err)
	}

	if len(files) == 0 {
		return fmt.Errorf("no json files found in parent: %s", c.inPath)
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

	if !cont() {
		return errors.New("aborted")
	}

	log.Println("begin requesting categories...")

	if err := domain.RequestCategories(notes, manifest, manifestPath); err != nil {
		return fmt.Errorf("cannot request categories: %w", err)
	}

	return nil
}

// parseFlag parses and sanity checks the required flag
func (c *Categorise) parseFlag() error {
	i := flag.String("i", "", "relative path to directory of cleaned files")
	flag.Parse()

	c.inPath = *i

	if c.inPath == "" {
		return errors.New("-i is empty")
	}

	return nil
}
