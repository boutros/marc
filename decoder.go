package marc

import (
	"fmt"
	"io"
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
	lex lexer
}

// NewDecoder returns a new Decoder using the given reader and format.
func NewDecoder(r io.Reader, f Format) *Decoder {
	switch f {
	case LineMARC:
		return &Decoder{lex: newLineLexer(r)}
	default:
		panic("TODO")
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
	var r Record
	var tok token

	// todo parse leader

	// parse control fields
	for tok = d.lex.Next(); tok.typ == tokenCtrlTag; tok = d.lex.Next() {
		f := cField{Tag: tok.value}
		tok = d.lex.Next()
		if tok.typ != tokenValue {
			break
		}
		f.Field = tok.value
		r.ctrlFields = append(r.ctrlFields, f)
	}
	switch tok.typ {
	case tokenERROR:
		line, col := d.lex.Pos()
		return r, fmt.Errorf("%d:%d %v: %q", line, col, d.lex.Error(), tok.value)
	case tokenTerminator:
		return d.Decode()
	case tokenEOF:
		return r, io.EOF
	case tokenTag:
		// go on to parse data fields
	default:
		fmt.Printf("%v", tok)
		panic("TODO")
	}

	// parse data fields (tok is allready a tag)
datafields:
	for {
		f := dField{
			Tag:  tok.value[0:3],
			Ind1: tok.value[3:4],
			Ind2: tok.value[4:5],
		}

		// parse subFields
		for {
			tok = d.lex.Next()
			if tok.typ != tokenSubField {
				break
			}
			sf := subField{Code: tok.value}
			tok = d.lex.Next()
			if tok.typ != tokenValue {
				break
			}
			sf.Value = tok.value
			f.SubFields = append(f.SubFields, sf)
		}
		r.dataFields = append(r.dataFields, f)

		switch tok.typ {
		case tokenTag:
			continue
		case tokenEOF, tokenTerminator:
			break datafields
		default:
			fmt.Printf("%s %s", tok.typ, tok.value)
			panic("TODO")
		}

	}
	//fmt.Printf("%v", r)
	return r, nil
}
