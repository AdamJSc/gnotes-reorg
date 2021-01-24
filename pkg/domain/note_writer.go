package domain

// NoteWriter defines the required behaviour for writing a Note
type NoteWriter interface {
	Write(n Note) error
}
