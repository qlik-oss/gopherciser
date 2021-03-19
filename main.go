package main

import (
	_ "net/http/pprof"

	_ "github.com/qlik-oss/gopherciser/elastic"

	"github.com/qlik-oss/gopherciser/cmd"
)

// Compile documentation data to be used by GUI and for markdown generation
//go:generate go run ./generatedocs/cmd/compiledocs

// Generate markdown files
//go:generate go run ./generatedocs/cmd/generatemarkdown --output ./docs/settingup.md

func main() {
	cmd.Execute()
}
