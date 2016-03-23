package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/boutros/marc"
)

type tagStats map[string]int

type stats []tagStats

func newStats() stats {
	s := make([]tagStats, 1000)
	for i := 0; i < 1000; i++ {
		s[i] = make(tagStats)
	}
	return s
}

func (s stats) countRecord(r marc.Record) error {
	for _, df := range r.DataFields {
		i, err := strconv.Atoi(df.Tag)
		if err != nil {
			return err
		}

		for _, sf := range df.SubFields {
			s[i][sf.Code] = s[i][sf.Code] + 1
		}
	}
	return nil
}

func (s stats) Dump() {
	var (
		bold  = "\x1b[1m"
		reset = "\x1b[0m"
		//faint = "\x1b[2m"
	)
	for i, e := range s {
		if len(e) == 0 {
			continue
		}
		fmt.Printf("%03d  ", i)
		for code, count := range e {
			fmt.Printf("%s%s%s: %d ", bold, code, reset, count)
		}
		fmt.Println()
	}
}

func main() {
	log.SetFlags(0)
	log.SetPrefix("marcstats: ")

	tags := flag.String("t", "", "generate statistics from these tag contents (ex '100e,700a' )")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Usage: marcstats -t<tags> <marcdatabase>\n")
		os.Exit(1)
	}

	f, err := os.Open(flag.Args()[0])
	if err != nil {
		log.Fatal(err)
	}

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
	stats := newStats()
	var filterTags []string
	if *tags != "" {
		filterTags = strings.Split(*tags, ",")
		for _, t := range filterTags {
			if len(t) != 4 {
				log.Fatalf("wrong tag filter format: %q (should be a three-digit number plus one character, ex. \"100e\")", t)
			}
		}
	}
	filterStats := make(map[string]tagStats)

	for rec, err := dec.Decode(); err != io.EOF; rec, err = dec.Decode() {
		if err != nil {
			log.Fatal(err)
		}
		if len(filterTags) > 0 {
			filterCounts(filterTags, filterStats, rec)
		} else {
			if err := stats.countRecord(rec); err != nil {
				log.Println(err)
			}
		}
		c++
	}
	fmt.Printf("Done in %s\n", time.Now().Sub(start))
	fmt.Printf("Number of records: %d\n", c)
	if len(filterTags) > 0 {
		for tag, e := range filterStats {
			fmt.Println(tag)
			ranks := rankByWordCount(e)
			for _, p := range ranks {
				fmt.Printf("\t%6d %s\n", p.Value, p.Key)
			}
			fmt.Println()
		}
	} else {
		stats.Dump()
	}
}

func filterCounts(filterTags []string, filterStats map[string]tagStats, rec marc.Record) {
	for _, df := range rec.DataFields {
		for _, ft := range filterTags {
			if df.Tag == ft[0:3] {
				for _, sf := range df.SubFields {
					if sf.Code == ft[3:] {
						if _, ok := filterStats[ft]; !ok {
							filterStats[ft] = make(tagStats)
						}
						filterStats[ft][sf.Value] = filterStats[ft][sf.Value] + 1
					}
				}
			}
		}
	}
}

func rankByWordCount(wordFrequencies map[string]int) PairList {
	pl := make(PairList, len(wordFrequencies))
	i := 0
	for k, v := range wordFrequencies {
		pl[i] = Pair{k, v}
		i++
	}
	sort.Sort(sort.Reverse(pl))
	return pl
}

type Pair struct {
	Key   string
	Value int
}

type PairList []Pair

func (p PairList) Len() int           { return len(p) }
func (p PairList) Less(i, j int) bool { return p[i].Value < p[j].Value }
func (p PairList) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }
