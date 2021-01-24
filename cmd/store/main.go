package main

import (
	"flag"
	"log"
	"reorg/pkg/adapters"
	"reorg/pkg/command"
	"reorg/pkg/domain"
)

func main() {
	osfs := &adapters.OsFileSystem{}
	filesService := domain.NewFileSystemService(osfs)

	i, f, g := parseFlags()

	var wr domain.NoteWriter

	switch {
	case f == g:
		log.Fatal("must specify destination either file system or google")
	case f:
		wr = &adapters.TxtNoteWriter{SubDir: "categorised", Files: filesService}
	case g:
		wr = &adapters.GoogleStorageNoteWriter{}
	}

	command.Run(&command.Store{
		InPath: i,
		Writer: wr,
		Files:  filesService,
		Notes:  domain.NewNoteService(osfs),
	})
}

// parseFlags parses the required flags
func parseFlags() (string, bool, bool) {
	i := flag.String("i", "", "relative path to directory of cleaned files and manifest")
	f := flag.Bool("f", false, "destination file system <input_path>/categorised")
	g := flag.Bool("g", false, "destination google storage")

	flag.Parse()

	return *i, *f, *g
}
