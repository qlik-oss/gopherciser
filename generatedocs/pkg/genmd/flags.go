package genmd

import (
	"flag"
	"os"
)

var (
	output  string
	wiki    string
	verbose bool
)

func handleFlags() {
	flagHelp := flag.Bool("help", false, "shows help")
	flag.StringVar(&output, "output", "", "path to output file")
	flag.StringVar(&wiki, "wiki", "", "path to wiki folder")
	flag.BoolVar(&verbose, "verbose", false, "verbose print")

	flag.Parse()

	if *flagHelp {
		flag.PrintDefaults()
		os.Exit(ExitCodeOk)
	}
}
