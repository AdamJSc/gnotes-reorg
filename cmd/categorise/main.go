package main

import (
	"flag"
	"reorg/pkg/adapters"
	"reorg/pkg/command"
	"reorg/pkg/domain"
)

func main() {
	osfs := &adapters.OsFileSystem{}

	command.Run(&command.Categorise{
		InPath: parseFlag(),
		Files:  domain.NewFileSystemService(osfs),
		Notes:  domain.NewNoteService(osfs),
	})
}

// parseFlag parses the required flag
func parseFlag() string {
	i := flag.String("i", "", "relative path to directory of cleaned files")

	flag.Parse()

	return *i
}
