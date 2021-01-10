package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reorg/pkg/domain"
	"reorg/pkg/fs"
	"strings"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("failed: %s", err.Error())
	}
	log.Println("process complete!")
}

// run executes our business logic
func run() error {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	flagI := flag.String("i", "", "relative path to directory of cleaned files")
	flag.Parse()

	inPath, err := fs.ParseDirFlag(flagI)
	if err != nil {
		return fmt.Errorf("cannot parse -i: %w", err)
	}

	manifestPath := strings.Join([]string{inPath, "manifest.json"}, string(os.PathSeparator))
	manifestPath, err = filepath.Abs(manifestPath)
	if err != nil {
		return fmt.Errorf("absolute manifest path failed: %w", err)
	}

	log.Printf("scanning directory: %s", inPath)
	files, err := fs.GetChildPaths(inPath, false)
	if err != nil {
		return err
	}

	var jsonFiles []string
	for _, f := range files {
		if strings.HasSuffix(f, ".json") {
			jsonFiles = append(jsonFiles, f)
		}
	}

	if len(jsonFiles) == 0 {
		return fmt.Errorf("no json files found in parent: %s", inPath)
	}

	log.Printf("%d note files to categorise", len(jsonFiles))

	if !cont() {
		return errors.New("aborted")
	}

	log.Println("categorising notes...")
	if err := domain.CategoriseNotes(jsonFiles, manifestPath); err != nil {
		return fmt.Errorf("cannot categorise notes: %w", err)
	}

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
