package main

import (
	"reorg/pkg/command"
	"reorg/pkg/domain"
)

func main() {
	osfs := &domain.OsFileSystem{}

	command.Run(&command.Clean{
		Files: domain.NewFileSystemService(osfs),
		Notes: domain.NewNoteService(osfs),
	})
}
