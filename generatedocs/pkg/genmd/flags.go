package genmd

import (
	"flag"
	"os"
)

var output string

func handleFlags() {
	flagHelp := flag.Bool("help", false, "shows help")
	flag.StringVar(&output, "output", "generatedocs/generated/settingup.md", "path to output file")

	flag.Parse()

	if *flagHelp {
		flag.PrintDefaults()
		os.Exit(ExitCodeOk)
	}
}
