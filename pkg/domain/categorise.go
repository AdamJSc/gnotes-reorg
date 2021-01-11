package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
)

// noteManifest maps a note filename to its category
type noteManifest map[string]string

// ParseManifestFromPath parse the manifest from the provided path
func ParseManifestFromPath(manifestPath string) (noteManifest, error) {
	var manifest noteManifest

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
func RequestCategories(notes []Note) error {
	// TODO: implement
	return nil
}
