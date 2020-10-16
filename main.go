package main

import (
	"github.com/qlik-oss/gopherciser/cmd"
)

// Compile documentation data to be used by GUI and for markdown generation
//go:generate go run ./generatedocs/cmd/compiledocs

// Generate markdown files
//go:generate go run ./generatedocs/cmd/generatemarkdown --output ./docs/settingup.md

func main() {
	cmd.Execute()
}
