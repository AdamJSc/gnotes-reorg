package adapters

import (
	"log"
	"reorg/pkg/domain"
)

// GoogleStorageNoteWriter writes notes to Google Storage
type GoogleStorageNoteWriter struct {
	domain.NoteWriter
}

// Write implements domain.NoteWriter
func (g *GoogleStorageNoteWriter) Write(n domain.Note) error {
	// TODO: implement me
	log.Printf("google storage stub: note %s", n.Filename())
	return nil
}
