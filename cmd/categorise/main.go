package main

import (
	"reorg/pkg/adapters"
	"reorg/pkg/command"
	"reorg/pkg/domain"
)

func main() {
	osfs := &adapters.OsFileSystem{}

	command.Run(&command.Categorise{
		Files: domain.NewFileSystemService(osfs),
		Notes: domain.NewNoteService(osfs),
	})
}
