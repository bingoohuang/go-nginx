package nginxconf

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"unicode"

	"github.com/pkg/errors"
)

type tokenType int

const (
	EOF tokenType = iota
	braceOpen
	braceClose
	semicolon
	Word
	Comment
)

// nolint:gochecknoglobals
var (
	tokenTypeName = map[tokenType]string{
		EOF:        "EOF",
		braceOpen:  "BRACE_OPEN",
		braceClose: "BRACE_CLOSE",
		semicolon:  "SEMICOLON",
		Word:       "WORD",
		Comment:    "COMMENT",
	}
)

func (t tokenType) String() string {
	return tokenTypeName[t]
}

type Token struct {
	Typ tokenType
	Lit string
}

func (t Token) String() string {
	return fmt.Sprintf("%s:%s", t.Typ, t.Lit)
}

// nolint:gochecknoglobals
var (
	EOFToken        = Token{Typ: EOF}
	BraceOpenToken  = Token{Typ: braceOpen, Lit: "{"}
	BraceCloseToken = Token{Typ: braceClose, Lit: "}"}
	SemicolonToken  = Token{Typ: semicolon, Lit: ";"}
)

type Scanner struct {
	r    *bufio.Reader
	line int
}

func NewScanner(content []byte) *Scanner {
	return &Scanner{
		r: bufio.NewReader(bytes.NewBuffer(content)),
	}
}

func (s *Scanner) read() (rune, error) {
	r, _, err := s.r.ReadRune()
	if r == '\n' {
		s.line++
	}

	return r, err
}

func (s *Scanner) unread() {
	_ = s.r.UnreadRune()
}

func (s *Scanner) Scan() Token {
	s.skipWhitespace()
	r, err := s.read()

	if err == io.EOF {
		return EOFToken
	}

	switch r {
	case '\'':
		return s.scanQuoted('\'')
	case '"':
		return s.scanQuoted('"')
	case '{':
		return BraceOpenToken
	case '}':
		return BraceCloseToken
	case ';':
		return SemicolonToken
	case '#':
		return s.scanComment()
	}

	s.unread()

	return s.scanWord()
}

// ErrSyntax means that a syntax error occurred.
// nolint:gochecknoglobals
var ErrSyntax = errors.New("syntax error")

func (s *Scanner) scanQuoted(quote rune) Token {
	var buf bytes.Buffer

	quoted := false

ForLoop:
	for {
		r, err := s.read()
		if err == io.EOF {
			panic(errors.Wrapf(ErrSyntax, "missing terminating %v character at line %d", quote, s.line))
		}
		if quoted {
			switch r {
			case 'n':
				buf.WriteRune('\n')
			case 'r':
				buf.WriteRune('\r')
			case 't':
				buf.WriteRune('\t')
			case '"':
				buf.WriteRune('"')
			case '\'':
				buf.WriteRune('\'')
			case '\\':
				buf.WriteRune('\\')
			default:
				panic(errors.Wrapf(ErrSyntax, "invalid quoted character: '\\%c'", r))
			}
			quoted = false
			continue
		}
		switch r {
		case '\n':
			panic(errors.Wrapf(ErrSyntax, "missing terminating %v character at line %d", quote, s.line))
		case '\\':
			quoted = true
		case quote:
			break ForLoop
		default:
			buf.WriteRune(r)
		}
	}

	return Token{Typ: Word, Lit: buf.String()}
}

func (s *Scanner) skipWhitespace() {
	for r, err := s.read(); err != io.EOF; r, err = s.read() {
		if !unicode.IsSpace(r) {
			s.unread()
			break
		}
	}
}

func (s *Scanner) scanComment() Token {
	var buf bytes.Buffer

	for {
		r, err := s.read()
		if err == io.EOF || r == '\n' {
			break
		}

		buf.WriteRune(r)
	}

	return Token{Typ: Comment, Lit: buf.String()}
}

func (s *Scanner) scanWord() Token {
	var buf bytes.Buffer

	for {
		r, err := s.read()
		if err == io.EOF {
			break
		}

		if unicode.IsSpace(r) {
			break
		}

		if r == ';' {
			s.unread()
			break
		}

		buf.WriteRune(r)
	}

	return Token{Typ: Word, Lit: buf.String()}
}
