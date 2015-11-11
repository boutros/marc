package marc

import (
	"bufio"
	//"fmt"
	"bytes"
	"io"
	"unicode/utf8"
)

// Format represents a MARC serialization format
type Format int

// Supported serialization formats for encoding and decoding
const (
	MARC     Format = iota // Standard binary MARC (ISO2709)
	LineMARC               // Line mode MARC (ex: NORMARC)
	MARCXML                // MarcXchange (ISO25577)
)

// Decoder parses MARC records from an input stream.
type Decoder struct {
	r     *bufio.Reader
	input []byte
	pos   int // position in input
	f     Format
}

// NewDecoder returns a new Decoder using the given reader and format.
func NewDecoder(r io.Reader, f Format) *Decoder {
	return &Decoder{
		r: bufio.NewReader(r),
		f: f,
	}
}

// DecodeAll consumes the input stream and returns all decoded records.
// If there is an error, it will return, together with the succesfully
// parsed MARC records up til then.
func (d *Decoder) DecodeAll() ([]Record, error) {
	res := []Record{}
	for r, err := d.Decode(); err != io.EOF; r, err = d.Decode() {
		if err != nil {
			return res, err
		}
		res = append(res, r)
	}
	return res, nil
}

func (d *Decoder) Decode() (Record, error) {
	switch d.f {
	case LineMARC:
		return d.decodeLineMARC()
	default:
		panic("TODO")
	}
}

func (d *Decoder) next() rune {
	ch, size := utf8.DecodeRune(d.input[d.pos:])
	d.pos += size
	return ch
}

func (d *Decoder) peek() rune {
	ch, _ := utf8.DecodeRune(d.input[d.pos:])
	return ch
}

func (d *Decoder) consumeUntil(r rune) bool {
	for {
		ch, size := utf8.DecodeRune(d.input[d.pos:])
		if size == 0 {
			return false
		}
		if ch == r {
			break
		}
		d.pos += size
	}
	return true
}

func (d *Decoder) consumeUntilOr(r1, r2 rune) bool {
	for {
		ch, size := utf8.DecodeRune(d.input[d.pos:])
		if size == 0 {
			return false
		}
		if ch == r1 || ch == r2 {
			break
		}
		d.pos += size
	}
	return true
}

func (d *Decoder) nextN(n int) string {
	start := d.pos
	for n > 0 {
		_, size := utf8.DecodeRune(d.input[d.pos:])
		d.pos += size
		n--
	}
	return string(d.input[start:d.pos])
}

func (d *Decoder) decodeLineMARC() (r Record, err error) {
	if d.input, err = d.r.ReadBytes(0x5E); err != nil {
		return r, err
	}
	d.pos = 0

	// Parse leader? Not in LineMARC?
	// TODO research

	if d.peek() == '\n' {
		d.pos++
	}

	for d.next() == '*'  {
		s := d.pos // keep track of start of tag
		if bytes.HasPrefix(d.input[d.pos:], []byte("00")){
			d.pos+=3
			if len(d.input) < d.pos {
				return r, nil
			}
			// Parse controlfield

			f := cField{Tag: string(d.input[s:d.pos])}

			if d.consumeUntil('\n') {
				f.Value = string(d.input[s+3 : d.pos])
				// consume and ignore \n
				d.pos++
			}
			r.ctrlFields = append(r.ctrlFields, f)
			continue
		}
		// Parse datafield

		// consume last 3 chars tag + 2 chars indicators
		d.pos+=5
		if len(d.input) < d.pos {
			return r, nil
		}

		f := dField{
			Tag:  string(d.input[s : s+3]),
			Ind1: string(d.input[s+3 : s+4]),
			Ind2: string(d.input[s+4 : s+5]),
		}
		// parse subfields
		for d.next() == '$' {
			sf := subField{Code: string(d.next())}
			s = d.pos // keep track of subfield start
			if d.consumeUntilOr('$', '\n') {
				sf.Value = string(d.input[s:d.pos])
				if d.peek() == '\n' {
					f.SubFields = append(f.SubFields, sf)
					d.pos++
					break
				}
			}
			f.SubFields = append(f.SubFields, sf)
		}
		r.dataFields = append(r.dataFields, f)
	}

	return r, nil
}
