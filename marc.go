package marc

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"sort"
	"strings"
	"unicode/utf8"
)

const colWidth = 70

// Record represents a bibliographic record, serializable to MARC formt
type Record struct {
	XMLName    xml.Name `xml:"record"`
	Leader     string   `xml:"leader"` // 24 chars
	CtrlFields CFields  `xml:"controlfield"`
	DataFields DFields  `xml:"datafield"`
}

type CField struct {
	Tag   string `xml:"tag,attr"`  // 3 chars
	Value string `xml:",chardata"` // if Tag == "000"; 40 chars
}

type DField struct {
	Tag       string    `xml:"tag,attr"`  // 3 chars
	Ind1      string    `xml:"ind1,attr"` // 1 char
	Ind2      string    `xml:"ind2,attr"` // 1 char
	SubFields SubFields `xml:"subfield"`
}

type SubField struct {
	Code  string `xml:"code,attr"` // 1 char
	Value string `xml:",chardata"`
}

type CFields []CField
type DFields []DField
type SubFields []SubField

// Len satisfies the Sort interface for CFields.
func (f CFields) Len() int { return len(f) }

// Swap satisfies the Sort interface for CFields.
func (f CFields) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

// Less satisfies the Sort interface for CFields.
func (f CFields) Less(i, j int) bool {
	if f[i].Tag < f[j].Tag {
		return true
	}
	if f[i].Value < f[j].Value {
		return true
	}
	return false
}

// Len satisfies the Sort interface for DFields.
func (f DFields) Len() int { return len(f) }

// Swap satisfies the Sort interface for DFields.
func (f DFields) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

// Less satisfies the Sort interface for DFields.
func (f DFields) Less(i, j int) bool {
	if f[i].Tag < f[j].Tag {
		return true
	}
	if f[i].Ind1 < f[j].Ind1 {
		return true
	}
	if f[i].Ind2 < f[j].Ind2 {
		return true
	}
	return false
}

// Len satisfies the Sort interface for SubFields.
func (f SubFields) Len() int { return len(f) }

// Swap satisfies the Sort interface for SubFields.
func (f SubFields) Swap(i, j int) { f[i], f[j] = f[j], f[i] }

// Less satisfies the Sort interface for SubFields.
func (f SubFields) Less(i, j int) bool {
	if f[i].Code < f[j].Code {
		return true
	}
	if f[i].Value < f[j].Value {
		return true
	}

	return false
}

// Eq tests for Record equality.
func (r Record) Eq(other Record) bool {
	if r.Leader != other.Leader {
		return false
	}
	if len(r.CtrlFields) != len(other.CtrlFields) {
		return false
	}
	sort.Sort(r.CtrlFields)
	sort.Sort(other.CtrlFields)
	for i, f := range r.CtrlFields {
		if other.CtrlFields[i].Tag != f.Tag {
			return false
		}
	}
	return true
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
	for _, c := range r.CtrlFields {
		fmt.Fprintf(w, "%s%s%s %s\n", bold, c.Tag, reset, c.Value)
	}
	for _, d := range r.DataFields {
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
