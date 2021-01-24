package main

import (
	"flag"
	"reorg/pkg/adapters"
	"reorg/pkg/command"
	"reorg/pkg/domain"
)

func main() {
	osfs := &adapters.OsFileSystem{}

	command.Run(&command.Store{
		InPath: parseFlag(),
		Files:  domain.NewFileSystemService(osfs),
		Notes:  domain.NewNoteService(osfs),
	})
}

// parseFlag parses the required flag
func parseFlag() string {
	i := flag.String("i", "", "relative path to directory of cleaned files and manifest")

	flag.Parse()

	return *i
}
