package main

import (
	"os"

	"github.com/qba73/habit"
)

func main() {
	habit.RunCLI(os.Stdout, os.Stderr)
}
