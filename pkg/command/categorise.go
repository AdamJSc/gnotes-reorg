package command

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"reorg/pkg/domain"
	"strings"
)

// abridgeLen defines the number of Note content lines to render as a preview when specifying a category
const abridgedLen = 5

// defaultCategory defines the category to use when user input is empty
const defaultCategory = "_none"

// Categorise represents our categorise command
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

	notes, err := c.parseNotesFromFiles(files)
	if err != nil {
		return fmt.Errorf("cannot parse notes: %w", err)
	}

	log.Println("parsing manifest from file...")

	manifestPath, err := c.Files.ParseAbsPath(c.inPath, "manifest.json")
	if err != nil {
		return fmt.Errorf("cannot parse manifest path: %w", err)
	}

	manifest, err := c.Notes.ParseManifestFromPath(manifestPath)
	if err != nil {
		return fmt.Errorf("cannot parse manifest: %w", err)
	}

	log.Println("removing notes already processed...")

	notes = c.Notes.FilterNotesByManifest(notes, manifest, false)

	log.Println("sorting notes by filename descending (most recent timestamp first)...")

	notes = c.Notes.SortNotesByFilenameDesc(notes)

	log.Printf("%d note files to categorise", len(notes))

	if !cont() {
		return errors.New("aborted")
	}

	log.Println("begin requesting categories...")

	if err := c.requestCategories(notes, manifest); err != nil {
		return fmt.Errorf("cannot request categories: %w", err)
	}

	log.Printf("finished categorising %d notes", len(notes))

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

// parseNotesFromFiles parses Notes from the files at the provided paths
func (c *Categorise) parseNotesFromFiles(paths []string) ([]domain.Note, error) {
	var notes []domain.Note

	for _, p := range paths {
		n, err := c.Notes.ParseFromFile(p)
		if err != nil {
			return nil, fmt.Errorf("cannot parse note from file %s: %w", p, err)
		}

		notes = append(notes, n)
	}

	return notes, nil
}

// requestCategories requests categories for each of the provided Notes in turn
func (c *Categorise) requestCategories(notes []domain.Note, manifest domain.NoteManifest) error {
	for _, n := range notes {
		cat := requestCategory(n, true)

		if err := manifest.Set(n.Filename(), cat); err != nil {
			return fmt.Errorf("cannot set manifest value: %w", err)
		}

		if err := c.Notes.SaveManifest(manifest); err != nil {
			return fmt.Errorf("cannot save manifest: %w", err)
		}
	}

	return nil
}

// requestCategory outputs the provided Note to console and returns the subsequent user input
func requestCategory(n domain.Note, abridged bool) string {
	content := n.Content
	if abridged == true {
		lines := strings.Split(content, "\n")
		if len(lines) > abridgedLen {
			content = strings.Join(lines[:abridgedLen], "\n")
		}
	}

	fmt.Printf("%s %s:\n%s\n", n.Timestamp.Format("2006-01-02"), n.Title, content)
	fmt.Print("> category? [type `f` for full] ")

	s := bufio.NewScanner(os.Stdin)
	s.Scan()
	inp := s.Text()
	switch inp {
	case "f":
		// render full content
		return requestCategory(n, false)
	case "":
		inp = defaultCategory
	}
	return inp
}
