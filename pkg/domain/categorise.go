package domain

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

const abridgedLen = 5

// FilterNotesByManifest returns the provided Note objects that do not appear in the provided manifest
func FilterNotesByManifest(notes []Note, manifest noteManifest, keepIfPresent bool) []Note {
	var retained []Note

	for _, n := range notes {
		_, ok := manifest[n.filename()]
		if keepIfPresent == ok {
			// retain if note appears in manifest
			retained = append(retained, n)
		}
	}

	return retained
}

// SortNotesByFilenameDesc orders the provided notes ordered descending by filename
func SortNotesByFilenameDesc(notes []Note) []Note {
	sort.SliceStable(notes, func(i, j int) bool {
		n1 := notes[i]
		n2 := notes[j]
		return strings.Compare(n1.filename(), n2.filename()) > 0
	})

	return notes
}

// RequestCategories requests categories for each of the provided Notes in turn
func RequestCategories(notes []Note, manifest noteManifest, manifestPath string) error {
	var err error
	for _, n := range notes {
		key := n.filename()
		cat := renderAndRequestCategory(n, true)

		manifest, err = applyToManifest(key, cat, manifest)
		if err != nil {
			return fmt.Errorf("cannot apply manifest value to key %s: %w", key, err)
		}

		if err := saveManifest(manifest, manifestPath); err != nil {
			return fmt.Errorf("cannot save manifest: %w", err)
		}
	}
	return nil
}

// renderAndRequestCategory outputs the provided Note to console and returns the subsequent user input
func renderAndRequestCategory(n Note, abridged bool) string {
	content := n.Content
	if abridged == true {
		lines := strings.Split(content, "\n")
		if len(lines) > abridgedLen {
			content = strings.Join(lines[:5], "\n")
		}
	}

	fmt.Printf("%s %s:\n%s\n", n.Timestamp.Format("2006-01-02"), n.Title, content)
	fmt.Print("> category? [type `f` for full] ")

	s := bufio.NewScanner(os.Stdin)
	s.Scan()
	inp := s.Text()
	if inp == "f" {
		return renderAndRequestCategory(n, false)
	}
	return inp
}

// applyToManifest applies the provided category to the provided manifest at the provided key
func applyToManifest(key, cat string, manifest noteManifest) (noteManifest, error) {
	if _, ok := manifest[key]; ok {
		return nil, fmt.Errorf("key already exists: %s", key)
	}
	manifest[key] = cat
	return manifest, nil
}

// saveManifest saves the provided manifest to the provided path
func saveManifest(manifest noteManifest, manifestPath string) error {
	buf := bytes.NewBuffer(nil)
	if err := json.NewEncoder(buf).Encode(&manifest); err != nil {
		return fmt.Errorf("cannot json encode manifest: %w", err)
	}

	if err := ioutil.WriteFile(manifestPath, buf.Bytes(), 0644); err != nil {
		return fmt.Errorf("cannot write to file %s: %w", manifestPath, err)
	}

	return nil
}
