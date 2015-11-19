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
		fmt.Fprintf(os.Stderr, "Usage: marcheck <marcdatabase>")
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

	dec := marc.NewDecoder(f, marc.LineMARC)
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
