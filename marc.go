package marc

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

const colWidth = 70

// Record represents a bibliographic record, serializable to MARC formt
type Record struct {
	leader     [24]byte
	ctrlFields []cField
	dataFields []dField
}

type cField struct {
	Tag   string // 3 chars
	Value string // if Tag == "000"; 40 chars
}

type dField struct {
	Tag       string // 3 chars
	Ind1      string // 1 char
	Ind2      string // 1 char
	SubFields []subField
}

type subField struct {
	Code  string // 1 char
	Value string
}

// DumpTo dumps a Record to the give writer
func (r Record) DumpTo(w io.Writer, colors bool) {
	bold, reset, faint, green := "", "", "", ""
	if colors {
		bold = "\x1b[1m"
		reset = "\x1b[0m"
		faint = "\x1b[2m"
		green = "\x1b[32m"
	}

	orBlank := func(s string) string {
		if len(s) == 0 || s == " " {
			return "_"
		}
		return s
	}
	fmt.Fprintf(w, "leader\n")
	for _, c := range r.ctrlFields {
		fmt.Fprintf(w, "%s%s%s %s\n", bold, c.Tag, reset, c.Value)
	}
	for _, d := range r.dataFields {
		fmt.Fprintf(w, "%s%s %s%s%s%s ",
			bold, d.Tag, faint, orBlank(d.Ind1), orBlank(d.Ind2), reset)

		var b bytes.Buffer
		for _, s := range d.SubFields {
			fmt.Fprintf(&b, "|%s %s ", s.Code, s.Value)
		}
		fields := strings.Fields(b.String())

		// current rune-count in line
		c := 0

		for _, f := range fields {
			wlen := utf8.RuneCountInString(f)
			if c+wlen > colWidth {
				// Wrap to new line and indent
				w.Write([]byte("\n       "))
				c = 0
				if f[0] != '|' {
					// Not subfield code; indent along with start of text in above line
					w.Write([]byte("   "))
					c += 2
				}
			}
			if f[0] == '|' {
				// subfield code, with color escape
				fmt.Fprintf(w, "%s%s%s", green, f, reset)
			} else {
				// subfield value
				w.Write([]byte(f))
			}

			// Write again space stripped by strings.Fields
			w.Write([]byte(" "))
			c += wlen + 1
		}
		fmt.Fprintf(w, "\n")
	}
	fmt.Fprintf(w, "\n")
}
