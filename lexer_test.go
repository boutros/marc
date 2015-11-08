package marc

import (
	"bytes"
	"fmt"
	"reflect"
	"testing"
)

func lexTokens(input string) []tokenType {
	lex := newLineLexer(bytes.NewBufferString(input))
	var res []tokenType
	for {
		tok := lex.Next()
		res = append(res, tok.typ)
		if tok.typ == tokenEOF {
			break
		}
	}
	return res
}

func lexValues(input string) []string {
	lex := newLineLexer(bytes.NewBufferString(input))
	var res []string
outer:
	for {
		tok := lex.Next()
		switch tok.typ {
		case tokenEOF:
			break outer
		case tokenERROR:
			res = append(res, fmt.Sprintf("%s: %q", lex.Error(), tok.value))
		default:
			res = append(res, tok.value)
		}
	}
	return res
}

func TestLineLexerTokens(t *testing.T) {
	var tests = []struct {
		input string
		want  []tokenType
	}{
		{"", []tokenType{tokenEOF}},
		{"abc", []tokenType{tokenValue, tokenEOF}},
		{"100", []tokenType{tokenValue, tokenEOF}},
		{"*001", []tokenType{tokenCtrlTag, tokenEOF}},
		{"*009", []tokenType{tokenCtrlTag, tokenEOF}},
		{"*100  ", []tokenType{tokenTag, tokenEOF}},
		{"*245  ", []tokenType{tokenTag, tokenEOF}},
		{"*100_1", []tokenType{tokenTag, tokenEOF}},
		{"*100  $aa\n*101  $ab", []tokenType{tokenTag, tokenSubField, tokenValue, tokenTag, tokenSubField, tokenValue, tokenEOF}},
		{"*000 01307nam0 2200349 I 450", []tokenType{tokenCtrlTag, tokenValue, tokenEOF}},
	}

	for _, test := range tests {
		tokens := lexTokens(test.input)
		if !reflect.DeepEqual(tokens, test.want) {
			t.Errorf("lineLexer lexing %q got %v; want %v",
				test.input, tokens, test.want)
		}
	}
}

func TestLineLexerValue(t *testing.T) {
	var tests = []struct {
		input string
		want  []string
	}{
		{"", nil},
		{"abc", []string{"abc"}},
		{"100", []string{"100"}},
		{"*001", []string{"001"}},
		{"*009", []string{"009"}},
		{"*24510", []string{"24510"}},
		{"*1001_$aSandburg, Carl$d1878-1967", []string{"1001_", "a", "Sandburg, Carl", "d", "1878-1967"}},
		{"*100  $aa\n*101  $bb", []string{"100  ", "a", "a", "101  ", "b", "b"}},
		{"*000xyz", []string{"000", "xyz"}},
	}

	for _, test := range tests {
		tokens := lexValues(test.input)
		if !reflect.DeepEqual(tokens, test.want) {
			t.Errorf("lineLexer lexing %q got %q; want %v",
				test.input, tokens, test.want)
		}
	}
}
