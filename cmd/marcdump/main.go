package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/boutros/marc"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("marcdump: ")
	var (
		useColors = flag.Bool("color", true, "use colored terminal output")
		filter    = flag.String("filter", "", "only print specified fields, ex.: 100b,245a")
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: marcdump [options...] file\n\nOptions:\n")
		flag.PrintDefaults()
	}

	if len(os.Args) < 2 || strings.HasPrefix(os.Args[len(os.Args)-1], "-") {
		flag.Usage()
		os.Exit(1)
	}

	flag.Parse()

	f, err := os.Open(os.Args[len(os.Args)-1])
	if err != nil {
		log.Fatal(err)
	}
	dec := marc.NewDecoder(f, marc.LineMARC)
	for r, err := dec.Decode(); err != io.EOF; r, err = dec.Decode() {
		if err != nil {
			log.Fatal(err)
		}
		r.DumpTo(os.Stdout, *useColors)
	}

	log.Println(*useColors, *filter)
}
