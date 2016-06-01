package marc

import (
	"bufio"
	"bytes"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"strconv"
	"unicode/utf8"
)

// leaderTemplate is used to fill in default values in leader,
// unless defined in record. (Usually applies to LineMARC,
// which omits the leader.)
var leaderTemplate = []byte("     c   a22        4500")

// Format represents a MARC serialization format
type Format int

// Supported serialization formats for encoding and decoding
const (
	unknown  Format = iota // Unparsable
	MARC                   // Standard binary MARC (ISO2709)
	LineMARC               // Line mode MARC (ex: NORMARC)
	MARCXML                // MarcXchange (ISO25577)
)

// String returns a string representation of a Format.
func (f Format) String() string {
	switch f {
	case unknown:
		return "Unknown MARC format"
	case MARC:
		return "Standard MARC (ISO2709)"
	case LineMARC:
		return "Line-MARC"
	case MARCXML:
		return "MarcXchange (ISO25577)"
	default:
		panic("unreachable")
	}
}

// DetectFormat tries to detect the MARC encoding of the given byte slice. It
// detects one of LineMARC/MARC/MARCXML, or otherwise unknown.
func DetectFormat(data []byte) Format {
	// Find the first non-whitespace byte
	i := 0
	for ; i < len(data) && isWS(data[i]); i++ {
	}
	switch data[i] {
	case '<':
		return MARCXML
	case '*': // TODO also '^' ?
		return LineMARC
	default:
		if data[i] >= '0' && data[i] <= '9' {
			return MARC
		}
		return unknown
	}
}

func isWS(b byte) bool {
	switch b {
	case '\t', '\n', '\x0c', '\r', ' ':
		return true
	}
	return false
}

type Encoder struct {
	w      *bufio.Writer
	xmlEnc *xml.Encoder
	f      Format
}

func (enc *Encoder) Encode(r Record) (err error) {
	// TODO revise this writer solution
	type writer interface {
		io.Writer
		io.ByteWriter
	}
	var writeByte = func(w writer, b byte) int {
		if err != nil {
			return 0
		}
		err = w.WriteByte(b)
		return 1
	}
	var writeString = func(w writer, s string) int {
		if err != nil {
			return 0
		}
		var n int
		n, err = io.WriteString(w, s)
		return n
	}
	var oneChar = func(s string) byte {
		if len(s) == 0 {
			return ' '
		}
		return s[0]
	}

	switch enc.f {
	case MARCXML:
		return enc.xmlEnc.Encode(r)
	case LineMARC:
		writeString(enc.w, "*000")
		writeString(enc.w, r.Leader)
		writeByte(enc.w, '\n')
		for _, f := range r.CtrlFields {
			writeByte(enc.w, '*')
			writeString(enc.w, f.Tag)
			writeString(enc.w, f.Value)
			writeByte(enc.w, '\n')
		}
		for _, f := range r.DataFields {
			writeByte(enc.w, '*')
			writeString(enc.w, f.Tag)
			writeByte(enc.w, oneChar(f.Ind1))
			writeByte(enc.w, oneChar(f.Ind2))
			for _, s := range f.SubFields {
				writeByte(enc.w, '$')
				writeByte(enc.w, oneChar(s.Code))
				writeString(enc.w, s.Value)
			}
			writeByte(enc.w, '\n')
		}
		writeString(enc.w, "^\n")
		return err
	case MARC:
		const (
			fs = '' // field separator
			ss = '' // subfield separator
			rt = '' // record terminator
		)
		var (
			head bytes.Buffer // leader + directory
			body bytes.Buffer // control fields + data fields
			p    = 0          // position in body
		)
		for _, f := range r.CtrlFields {
			start := p
			p += writeString(&body, f.Value)
			p += writeByte(&body, fs)

			writeString(&head, f.Tag) // TODO make sure Tag is 3 chars
			writeString(&head, fmt.Sprintf("%04d", len(f.Value)+1))
			writeString(&head, fmt.Sprintf("%05d", start))
		}
		for _, f := range r.DataFields {
			start := p
			p += writeString(&body, f.Ind1) // TODO make sure Ind1 is 1 char
			p += writeString(&body, f.Ind2) // TODO make sure Ind2 is 1 char
			p += writeByte(&body, ss)
			writeString(&head, f.Tag) // TODO make sure Tag is 3 chars
			for i, sf := range f.SubFields {
				p += writeString(&body, sf.Code) // TODO make sure Code is 1 char
				p += writeString(&body, sf.Value)
				if i < len(f.SubFields)-1 {
					p += writeByte(&body, ss)
				}
			}
			p += writeByte(&body, fs)
			writeString(&head, fmt.Sprintf("%04d", p-start))
			writeString(&head, fmt.Sprintf("%05d", start))
		}
		writeByte(&head, fs)
		writeByte(&body, rt)
		// We copy the computed size, even if allready present in leader
		size := 24 + len(head.Bytes()) + len(body.Bytes())

		if size > 99999 {
			return fmt.Errorf("record is bigger than max supported size in binary MARC (99999): %d", size)
		}
		//fmt.Printf("leader: %s computed: %d\n", r.Leader[0:5], size)
		//fmt.Println(head.String())
		//fmt.Println(body.String())
		_, err = enc.w.WriteString(fmt.Sprintf("%05d", size))
		if err != nil {
			return err
		}
		_, err = enc.w.WriteString(r.Leader[5:])
		if err != nil {
			return err
		}
		_, err = enc.w.Write(head.Bytes())
		if err != nil {
			return err
		}
		_, err = enc.w.Write(body.Bytes())
		return err
	default:
		panic("Encode Unknown")
	}
}

func (enc *Encoder) Flush() error {
	if enc.w == nil {
		return nil
	}
	return enc.w.Flush()
}

func NewEncoder(w io.Writer, f Format) *Encoder {
	switch f {
	case MARCXML:
		return &Encoder{xmlEnc: xml.NewEncoder(w), f: f}
	default:
		return &Encoder{w: bufio.NewWriter(w), f: f}
	}
}

// Decoder parses MARC records from an input stream.
type Decoder struct {
	r      *bufio.Reader
	xmlDec *xml.Decoder
	input  []byte
	pos    int // position in input
	f      Format
}

// NewDecoder returns a new Decoder using the given reader and format.
func NewDecoder(r io.Reader, f Format) *Decoder {
	switch f {
	case LineMARC:
		return &Decoder{r: bufio.NewReader(r), f: f}
	case MARCXML:
		return &Decoder{xmlDec: xml.NewDecoder(r), f: f}
	default:
		return &Decoder{r: bufio.NewReader(r), f: f}
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
	case MARCXML:
		var r Record
		for {
			t, _ := d.xmlDec.Token()
			if t == nil {
				break
			}
			switch elem := t.(type) {
			case xml.StartElement:
				if elem.Name.Local == "record" {
					err := d.xmlDec.DecodeElement(&r, &elem)
					return r, err
				}
			}
		}
		return r, io.EOF
	default:
		return d.decodeMARC()
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
	// Some records might include the ^ characters, notably in the leader,
	// so we check to make sure we reached a record terminator
	// TODO flag the record for replacement of ^ with space in leader and control fields
	for d.input[len(d.input)-2] != '\n' {
		// Most likely it's a leader or control field 008 where spaces
		// are indicated with ^, so we read to the end of the line.
		b, err := d.r.ReadBytes('\n')
		if err != nil {
			return r, err
		}
		d.input = append(d.input, b...)
		// Read to next terminator (hopefully)
		b, err = d.r.ReadBytes(0x5E)
		if err != nil {
			return r, err
		}
		d.input = append(d.input, b...)
	}

	d.pos = 0

	if d.peek() == '\n' {
		d.pos++
	}

	leader := make([]byte, 24)

	for d.next() == '*' {
		s := d.pos // keep track of start of tag
		if bytes.HasPrefix(d.input[d.pos:], []byte("00")) {
			d.pos += 3
			if len(d.input) < d.pos {
				return r, nil
			}
			// Parse controlfield

			f := CField{Tag: string(d.input[s:d.pos])}
			if d.consumeUntil('\n') {
				if d.input[s+2] == '0' {
					// controlfield 000 = leader
					copy(leader, d.input[s+3:d.pos])
				} else {
					f.Value = string(d.input[s+3 : d.pos])
					r.CtrlFields = append(r.CtrlFields, f)
				}
				// consume and ignore \n
				d.pos++
			}

			continue
		}
		// Parse datafield

		// consume last 3 chars tag + 2 chars indicators
		d.pos += 5
		if len(d.input) < d.pos {
			return r, nil
		}

		f := DField{
			Tag:  string(d.input[s : s+3]),
			Ind1: string(d.input[s+3 : s+4]),
			Ind2: string(d.input[s+4 : s+5]),
		}
		// parse subfields
		for d.next() == '$' {
			sf := SubField{Code: string(d.next())}
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
		r.DataFields = append(r.DataFields, f)
	}

	// replace spaces with chars from leader template
	for i, c := range leader {
		if c == '\x00' {
			leader[i] = leaderTemplate[i]
		}
	}
	r.Leader = string(leader)

	return r, nil
}

func (d *Decoder) decodeMARC() (Record, error) {
	const recordTerminator = ''
	var r Record

	b, err := d.r.ReadBytes(recordTerminator)
	if err != nil && len(b) == 0 {
		return r, err
	}
	if len(b) < 24 {
		return r, io.EOF
	}

	r.Leader = string(b[0:24])
	size, err := strconv.Atoi(r.Leader[0:5])
	if err != nil {
		return r, errors.New("leader pos 0:5 not an integer")
	}
	if size != len(b) {
		return r, fmt.Errorf("leader reports size %d; actual size is %d\n", size, len(b))
	}

	// leader+directory length
	ll, err := strconv.Atoi(r.Leader[12:17])
	if err != nil {
		return r, errors.New("leader pos 12:17 not an integer")
	}
	p := 24 // position
	for p < ll-1 {
		if bytes.HasPrefix(b[p:], []byte("00")) {
			// control field
			fl, err := strconv.Atoi(string(b[p+3 : p+7]))
			if err != nil {
				return r, errors.New("directory item field length not an integer")
			}
			fs, err := strconv.Atoi(string(b[p+7 : p+12]))
			if err != nil {
				return r, errors.New("directory item field starting position not an integer")
			}
			if ll+fs+fl-1 > size {
				return r, errors.New("directory item starting position/length out of bounds")
			}
			f := CField{Tag: string(b[p : p+3]), Value: string(b[ll+fs : ll+fs+fl-1])}
			r.CtrlFields = append(r.CtrlFields, f)
		} else {
			// data field
			fl, err := strconv.Atoi(string(b[p+3 : p+7]))
			if err != nil {
				return r, errors.New("directory item field length not an integer")
			}
			fs, err := strconv.Atoi(string(b[p+7 : p+12]))
			if err != nil {
				return r, errors.New("directory item field starting position not an integer")
			}
			if ll+fs+fl-1 > size {
				return r, errors.New("directory item starting position/length out of bounds")
			}
			f := DField{
				Tag:  string(b[p : p+3]),
				Ind1: string(b[ll+fs : ll+fs+1]),
				Ind2: string(b[ll+fs+1 : ll+fs+2]),
			}
			// parse subfields
			for _, s := range bytes.Split(b[ll+fs+2:ll+fs+fl-1], []byte("")) {
				if len(s) > 1 {
					f.SubFields = append(f.SubFields,
						SubField{Code: string(s[:1]), Value: string(s[1:])})
				}
			}
			r.DataFields = append(r.DataFields, f)
		}
		p += 12
	}

	return r, nil
}
