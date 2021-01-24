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

	i, o, j, t := parseFlags()

	var wr domain.NoteWriter

	switch {
	case j == t:
		log.Fatal("must specify output either json or txt")
	case t:
		wr = &adapters.TxtNoteWriter{Files: filesService}
	case j:
		wr = &adapters.JSONNoteWriter{Files: filesService}
	}

	command.Run(&command.Clean{
		InPath:  i,
		OutPath: o,
		Writer:  wr,
		Files:   filesService,
		Notes:   domain.NewNoteService(osfs),
	})
}

// parseFlags parses the required flags
func parseFlags() (string, string, bool, bool) {
	i := flag.String("i", "", "relative path to gnotes export directory")
	o := flag.String("o", "", "relative path to output directory for cleaned notes")
	j := flag.Bool("json", false, "output cleaned notes as json files")
	t := flag.Bool("txt", false, "output cleaned notes as txt files")

	flag.Parse()

	return *i, *o, *j, *t
}
