package app

import (
	"bufio"
	"fmt"
	"os"
)

func Cont() bool {
	fmt.Print("> continue? [Y/n] ")

	s := bufio.NewScanner(os.Stdin)
	if s.Scan(); s.Text() != "Y" {
		return false
	}

	return true
}
