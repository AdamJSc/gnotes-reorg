package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reorg/pkg/domain"
	"reorg/pkg/fs"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("failed: %s", err.Error())
	}
}

// run executes our business logic
func run() error {
	flagI := flag.String("i", "", "input directory, relative to cwd")
	flagO := flag.String("o", "", "output directory, relative to cwd")
	flag.Parse()

	inpPath, outPath, err := fs.ParseIO(flagI, flagO)
	if err != nil {
		return err
	}

	outPath = fmt.Sprintf("%s/output", outPath)
	outPath, err = filepath.Abs(outPath)
	if err != nil {
		return fmt.Errorf("absolute out path failed: %w", err)
	}

	log.Printf("scanning directory: %s", inpPath)
	dirs, err := fs.GetSubDirsFromPath(inpPath)
	if err != nil {
		return err
	}

	if len(dirs) == 0 {
		return fmt.Errorf("no directories found in parent: %s", inpPath)
	}

	log.Printf("%d directories to search for note files", len(dirs))

	if !cont() {
		return errors.New("aborted")
	}

	log.Println("parsing notes...")

	notes, err := domain.ParseNotes(dirs, outPath)
	if err != nil {
		return err
	}

	log.Printf("parsed %d notes\n", len(notes))
	log.Printf("writing to directory: %s", outPath)
	log.Println("this will reset its existing contents")

	if !cont() {
		return errors.New("aborted")
	}

	if err := domain.WriteNotes(context.Background(), notes, outPath); err != nil {
		return err
	}

	log.Printf("finished writing %d notes\n", len(notes))
	log.Println("process complete!")

	return nil
}

func cont() bool {
	fmt.Print("continue? [Y/n] ")

	s := bufio.NewScanner(os.Stdin)
	if s.Scan(); s.Text() != "Y" {
		return false
	}

	return true
}
