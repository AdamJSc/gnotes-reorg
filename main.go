package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"reorg/domain/fs"
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

	log.Printf("scanning directory: %s", inpPath)
	dirs, err := fs.GetSubDirsFromPath(inpPath)
	if err != nil {
		return err
	}

	if len(dirs) == 0 {
		return fmt.Errorf("no directories found in parent: %s", inpPath)
	}

	log.Printf("%d directories to search for note files", len(dirs))
	log.Printf("writing to directory: %s", outPath)
	fmt.Print("continue? [Y/n] ")

	s := bufio.NewScanner(os.Stdin)
	if s.Scan(); s.Text() != "Y" {
		return errors.New("unable to proceed")
	}

	log.Println("do work...")

	return nil
}
