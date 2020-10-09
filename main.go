package main

import (
	"github.com/qlik-oss/gopherciser/cmd"
)

// Compile documentation data to be used by GUI and for markdown generation
//go:generate go run ./generatedocs/cmd/compiledocs/main.go

// Generate markdown files
//go:generate go run ./generatedocs/generate/generate.go --output ./docs/settingup.md

func main() {
	cmd.Execute()
}
