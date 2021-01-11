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

// noteManifest maps a note filename to its category
type noteManifest map[string]string

// ParseManifestFromPath parse the manifest from the provided path
func ParseManifestFromPath(manifestPath string) (noteManifest, error) {
	manifest := make(noteManifest)

	// read contents of manifest file
	payload, err := ioutil.ReadFile(manifestPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("cannot parse existing manifest file %s: %w", manifestPath, err)
		}
		if _, err := os.Create(manifestPath); err != nil {
			return nil, fmt.Errorf("cannot create manifest file %s: %w", manifestPath, err)
		}
		return manifest, nil
	}

	// check if file contents are empty
	if len(payload) == 0 {
		return manifest, nil
	}

	// parse manifest
	r := bytes.NewReader(payload)
	if err := json.NewDecoder(r).Decode(&manifest); err != nil {
		return nil, fmt.Errorf("cannot json decode manifest %s: %w", manifestPath, err)
	}

	return manifest, nil
}

// ParseNotesFromPaths returns the notes whose payloads are stored in files at the provided paths
func ParseNotesFromPaths(paths []string) ([]Note, error) {
	var notes []Note

	for _, p := range paths {
		// read contents of file at path
		payload, err := ioutil.ReadFile(p)
		if err != nil {
			return nil, fmt.Errorf("cannot parse file as note %s: %w", p, err)
		}

		// parse json
		r := bytes.NewReader(payload)
		var n Note
		if err := json.NewDecoder(r).Decode(&n); err != nil {
			return nil, fmt.Errorf("cannot json decode file payload as note %s: %w", p, err)
		}

		notes = append(notes, n)
	}

	return notes, nil
}

// FilterNotesNotInManifest returns the provided Note objects that do not appear in the provided manifest
func FilterNotesNotInManifest(notes []Note, manifest noteManifest) []Note {
	var retained []Note

	for _, n := range notes {
		// keep note to process if it doesn't already appear in the manifest
		if _, ok := manifest[n.filename()]; !ok {
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
