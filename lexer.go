package marc

import (
	"bufio"
	"io"
	"unicode/utf8"
)

const newline = 0x0D
const eof = rune(0)

type lexer interface {
	// Next returns the next token in the stream.
	Next() token
	Error() string
	// Pos returns the line number and column of the token returned by Next()
	Pos() (int, int)
}

type tokenType int

type token struct {
	typ   tokenType
	value string
}

const (
	// special tokens
	tokenERROR tokenType = iota // io/encoding errors
	tokenEOF                    // end of stream

	// regular tokens
	tokenCtrlTag  // 3 chars
	tokenTag      // 5 chars (3 char for tag, 2 for indicators)
	tokenSubField // 1 char
	tokenValue    // value of subfield or control field
	tokenTerminator

// tokenLeader TODO
)

func (t tokenType) String() string {
	switch t {
	case tokenERROR:
		return "tokenERROR"
	case tokenEOF:
		return "tokenEOF"
	case tokenCtrlTag:
		return "tokenCtrlTag"
	case tokenTag:
		return "tokenTag"
	case tokenSubField:
		return "tokenSubField"
	case tokenValue:
		return "tokenValue"
	case tokenTerminator:
		return "tokenTerminator"
	default:
		panic("TODO")
	}
}

type lineLexer struct {
	r     *bufio.Reader
	input []byte // current record beeing lexed
	line  int    // TODO line number in input stream TODO?
	start int    // start of current token
	pos   int    // position in line
	err   string
}

func newLineLexer(r io.Reader) *lineLexer {
	return &lineLexer{r: bufio.NewReader(r)}

}

func (l *lineLexer) nextRune() rune {
	ch, size := utf8.DecodeRune(l.input[l.pos:])
	l.pos += size
	return ch
}

func (l *lineLexer) consume(n int) bool {
	ok := true
	for n > 0 {
		ch := l.nextRune()
		if ch == eof || ch == utf8.RuneError {
			ok = false
		}
		n--
	}
	return ok
}

func (l *lineLexer) consumeDigits(n int) bool {
	ok := true
	for n > 0 {
		ch := l.nextRune()
		if ch < '0' || ch > '9' {
			ok = false
		}
		n--
	}
	return ok
}

func (l *lineLexer) consumeUntil(r rune) {
	for {
		ch, size := utf8.DecodeRune(l.input[l.pos:])
		if ch == r || ch == '\n' || size == 0 {
			break
		}
		l.pos += size
	}
}

func (l *lineLexer) Next() token {
	if l.pos >= len(l.input) {
		line, err := l.r.ReadBytes(Terminator)
		switch err {
		case io.EOF:
			if len(line) == 0 {
				return token{tokenEOF, ""}
			}
			l.start, l.pos = 0, 0
			l.input = line
		case nil:
			l.start, l.pos = 0, 0
			l.input = line
		default:
			l.err = err.Error()
			return token{tokenERROR, ""}
		}
	}

	l.start = l.pos
	ch := l.nextRune()

	switch ch {
	case '^':
		return token{tokenTerminator, ""}
	case '\n':
		l.start = l.pos
		return l.Next()
	case utf8.RuneError:
		l.err = "invalid UTF-8 encoding"
		return token{tokenERROR, l.value()}
	case eof:
		// fill input buffer again
		return l.Next()
	case '*':
		l.start = l.pos
		if !l.consumeDigits(3) {
			l.err = "non-digit tag"
			return token{tokenERROR, l.value()}
		}
		if l.input[l.start] == '0' && l.input[l.start+1] == '0' {
			return token{tokenCtrlTag, l.value()}
		}
		if !l.consume(2) {
			return token{tokenEOF, ""}
		}
		if l.pos-l.start != 5 {
			// TODO add test
			l.err = "invalid tag"
			return token{tokenERROR, l.value()}
		}
		return token{tokenTag, l.value()}
	case Separator:
		l.start = l.pos
		ch := l.nextRune()
		if ch == eof {
			return l.Next()
		}
		return token{tokenSubField, l.value()}
	default:

		// lexing a value
		l.consumeUntil('$')
		return token{tokenValue, l.value()}
	}

}

func (l *lineLexer) Pos() (int, int) {
	return l.line, l.pos
}

func (l *lineLexer) Error() string {
	return l.err
}

func (l *lineLexer) value() string {
	return string(l.input[l.start:l.pos])
}
