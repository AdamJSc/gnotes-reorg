package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

const (
	titlePrefix      = "Back"
	createTimePrefix = "Create Time: "
	modTimePrefix    = "Modify Time: "
	tsFormat         = "02/01/2006 15:04"
	headerLines      = 5
)

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

		n.Path = p

		notes = append(notes, n)
	}

	return notes, nil
}
