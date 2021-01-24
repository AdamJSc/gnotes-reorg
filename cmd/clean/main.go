package main

import (
	"flag"
	"reorg/pkg/adapters"
	"reorg/pkg/command"
	"reorg/pkg/domain"
)

func main() {
	osfs := &adapters.OsFileSystem{}

	i, o, j, t := parseFlags()

	command.Run(&command.Clean{
		InPath:  i,
		OutPath: o,
		JSONOut: j,
		TxtOut:  t,
		Files:   domain.NewFileSystemService(osfs),
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
