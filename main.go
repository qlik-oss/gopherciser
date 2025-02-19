package main

import (
	_ "net/http/pprof"

	"github.com/qlik-oss/gopherciser/cmd"
)

// Compile documentation data to be used by GUI and for markdown generation
//go:generate go run ./generatedocs/cmd/compiledocs

func main() {
	cmd.Execute()
}
