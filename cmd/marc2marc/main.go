package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"os"

	"github.com/boutros/marc"
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("marc2marc: ")
}

func detectFormat(f *os.File) (marc.Format, error) {
	sniff := make([]byte, 64)
	_, err := f.Read(sniff)
	if err != nil {
		log.Fatal(err)
	}
	format := marc.DetectFormat(sniff)

	// rewind reader
	_, err = f.Seek(0, 0)
	if err != nil {
		log.Fatal(err)
	}

	switch format {
	case marc.MARC, marc.LineMARC, marc.MARCXML:
		return format, nil
	default:
		return format, errors.New("unknown MARC format")
	}
}

func main() {
	in := flag.String("i", "", "input file")
	f := flag.String("f", "", "output format: (m)arc, (l)ine-marc, marc(x)ml")

	flag.Parse()

	if *in == "" || *f == "" {
		flag.Usage()
		os.Exit(1)
	}

	inF, err := os.Open(*in)
	if err != nil {
		log.Fatal(err)
	}
	defer inF.Close()

	from, err := detectFormat(inF)
	if err != nil {
		log.Fatalf("%s: %v", inF.Name(), err)
	}

	var to marc.Format
	switch *f {
	case "m", "M":
		to = marc.MARC
	case "l", "L":
		to = marc.LineMARC
	case "x", "X":
		to = marc.MARCXML
	default:
		log.Println("illegal option for flag -f")
		flag.Usage()
		os.Exit(1)
	}

	if from == to {
		log.Println("nothing to do; input format same as output format")
		os.Exit(1)
	}

	dec := marc.NewDecoder(inF, from)
	enc := marc.NewEncoder(os.Stdout, to)

	for rec, err := dec.Decode(); err != io.EOF; rec, err = dec.Decode() {
		if err != nil {
			log.Println(err)
			continue
		}
		if err = enc.Encode(rec); err != nil {
			log.Println(err)
		}
	}
	enc.Flush()
}
