package main

import (
	"fmt"
	noteapp "github.com/dmundt/note/internal/note"
	"os"
)

func main() {
	if err := noteapp.Run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
