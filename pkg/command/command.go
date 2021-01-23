package command

import (
	"bufio"
	"fmt"
	"log"
	"os"
)

// Run invokes the provided Runner and handles the resulting error
func Run(r runner) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("running...")

	if err := r.Run(); err != nil {
		panic(err)
	}

	log.Println("complete!")
}

// runner defines the behaviour of a command runner
type runner interface {
	Run() error
}

// cont prompts the user for confirmation to continue
func cont() bool {
	fmt.Print("> continue? [Y/n] ")

	s := bufio.NewScanner(os.Stdin)
	if s.Scan(); s.Text() != "Y" {
		return false
	}

	return true
}
