package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/boutros/marc"
)

func main() {
	log.SetFlags(0)
	log.SetPrefix("marccheck: ")
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: marcheck <marcdatabase>\n")
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}
	stats, err := os.Stat(f.Name())
	if err != nil {
		log.Fatal(err)
	}
	size := stats.Size()

	// Detect format
	sniff := make([]byte, 64)
	_, err = f.Read(sniff)
	if err != nil {
		log.Fatal(err)
	}
	format := marc.DetectFormat(sniff)
	switch format {
	case marc.MARC, marc.LineMARC, marc.MARCXML:
		break
	default:
		log.Fatal("Unknown MARC format")
	}

	// rewind reader
	_, err = f.Seek(0, 0)
	if err != nil {
		log.Fatal(err)
	}

	dec := marc.NewDecoder(f, format)
	c := 0
	start := time.Now()

	for _, err := dec.Decode(); err != io.EOF; _, err = dec.Decode() {
		if err != nil {
			log.Fatal(err)
		}
		c++
	}
	fmt.Printf("Done in %s\n", time.Now().Sub(start))
	fmt.Printf("Number of records: %d\n", c)
	fmt.Printf("Average parsing speed: %.2f MB/s", float64(size)/time.Now().Sub(start).Seconds()/1048576)
}
