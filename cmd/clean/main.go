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

	inpPath, outPath, err := parseIO(flagI, flagO)
	if err != nil {
		return err
	}

	outPath = fmt.Sprintf("%s/output", outPath)
	outPath, err = filepath.Abs(outPath)
	if err != nil {
		return fmt.Errorf("absolute out path failed: %w", err)
	}

	log.Printf("scanning directory: %s", inpPath)
	dirs, err := fs.GetChildPaths(inpPath, true)
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

// parseIO validates and returns the provided input and output flags as formatted directory paths
func parseIO(i, o *string) (string, string, error) {
	inp := *i
	out := *o

	// input and output paths must be supplied
	if inp == "" {
		return "", "", errors.New("-i flag required")
	}
	if out == "" {
		return "", "", errors.New("-o flag required")
	}

	// input path must exist
	info, err := os.Stat(inp)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf("input path does not exist: %s", inp)
		}
		return "", "", err
	}

	// input path must represent a directory
	if !info.IsDir() {
		return "", "", fmt.Errorf("input path is not a directory: %s", inp)
	}

	// if output path does not exist, attempt to create it as a directory
	info, err = os.Stat(out)
	if err != nil {
		if !os.IsNotExist(err) {
			return "", "", err
		}
		if err := os.Mkdir(out, os.ModeDir); err != nil {
			return "", "", err
		}
	}

	// output path must represent a directory
	info, err = os.Stat(out)
	if err != nil {
		return "", "", err
	}
	if !info.IsDir() {
		return "", "", fmt.Errorf("output path is not a directory: %s", inp)
	}

	// fully-qualified input path includes a sub-directory
	fullInp := fmt.Sprintf("%s/Other", inp)

	// fully-qualified input path must exist
	info, err = os.Stat(fullInp)
	if err != nil {
		if os.IsNotExist(err) {
			return "", "", fmt.Errorf("input path sub-directory does not exist: %s", fullInp)
		}
		return "", "", err
	}

	// fully-qualified input path must represent a directory
	if !info.IsDir() {
		return "", "", fmt.Errorf("input path sub-directory is not a directory: %s", fullInp)
	}

	return fullInp, out, nil
}

func cont() bool {
	fmt.Print("continue? [Y/n] ")

	s := bufio.NewScanner(os.Stdin)
	if s.Scan(); s.Text() != "Y" {
		return false
	}

	return true
}
