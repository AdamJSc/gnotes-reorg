package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
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

	log.Printf("scanning directory: %s", inpPath)
	dirs, err := getSubDirsFromPath(inpPath)
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

// getSubDirsFromPath returns fully-qualified paths to all sub-directories within the provided parent directory
func getSubDirsFromPath(parent string) ([]string, error) {
	dirs := []string{}
	infos, err := ioutil.ReadDir(parent)
	if err != nil {
		return nil, err
	}
	for _, info := range infos {
		if info.IsDir() {
			dirs = append(dirs, fmt.Sprintf("%s/%s", parent, info.Name()))
		}
	}
	return dirs, nil
}
