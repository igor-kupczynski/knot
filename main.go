package main

import (
	"os"

	"github.com/igor-kupczynski/knot/internal/cmd"
)

func main() {
	os.Exit(cmd.Execute())
}
