package ini

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

// stateFunc represents a state for the parser's state machine.
type stateFunc func(p *Parser, r rune) (stateFunc, error)

// A Parser parses an INI file from an io.Reader.
type Parser struct {
	// Comments contains all runes which can start a comment.
	//
	// Default: #;
	Comments string

	// SectionStart is the rune which starts a section token.
	//
	// Default: [
	SectionStart rune

	// SectionEnd is the rune which ends a section token.
	//
	// Default: ]
	SectionEnd rune

	// Equals is the rune that separates the left side of a setting
	// token from the right.
	//
	// Default: =
	Equals rune

	r    *bufio.Reader
	line int
	pos  int
	eof  bool

	buf bytes.Buffer
	t   Token
}

// NewParser initializes a new Parser for the given io.Reader.
func NewParser(r io.Reader) *Parser {
	var rr *bufio.Reader
	if br, ok := r.(*bufio.Reader); ok {
		rr = br
	} else {
		rr = bufio.NewReader(r)
	}

	return &Parser{
		r: rr,

		Comments: "#;",

		SectionStart: '[',
		SectionEnd:   ']',

		Equals: '=',
	}
}

func (p *Parser) start(r rune) (stateFunc, error) {
	p.buf.Reset()

	if unicode.IsSpace(r) {
		return (*Parser).whitespace, nil
	}

	switch r {
	case p.SectionStart:
		return (*Parser).section, nil
	case '\n':
		return (*Parser).start, nil
	}

	if strings.ContainsRune(p.Comments, r) {
		p.r.UnreadRune()

		return (*Parser).comment, nil
	}

	p.r.UnreadRune()

	return (*Parser).left, nil
}

func (p *Parser) whitespace(r rune) (stateFunc, error) {
	if strings.ContainsRune(p.Comments, r) {
		p.r.UnreadRune()

		return (*Parser).comment, nil
	}

	if unicode.IsSpace(r) {
		return (*Parser).whitespace, nil
	}

	p.r.UnreadRune()

	return (*Parser).start, nil
}

func (p *Parser) section(r rune) (stateFunc, error) {
	switch r {
	case p.SectionStart:
		return nil, p.parseError("Unexpected rune: " + string(r))
	case p.SectionEnd:
		p.t = &SectionToken{
			start: p.SectionStart,
			end:   p.SectionEnd,

			Name: p.buf.String(),
		}

		return nil, nil
	}

	if strings.ContainsRune(p.Comments, r) {
		return nil, p.parseError("Unexpected rune: " + string(r))
	}

	p.buf.WriteRune(r)

	return (*Parser).section, nil
}

func (p *Parser) comment(r rune) (stateFunc, error) {
	if p.t == nil {
		p.t = &CommentToken{
			start: r,
		}

		return (*Parser).comment, nil
	}

	if r == '\n' {
		p.t.(*CommentToken).Comment = p.buf.String()

		return nil, nil
	}

	p.buf.WriteRune(r)

	return (*Parser).comment, nil
}

func (p *Parser) left(r rune) (stateFunc, error) {
	switch r {
	case '\n':
		return nil, p.parseError("Newline in left-hand side")
	case p.Equals:
		p.t = &SettingToken{
			equals: r,

			Left: p.buf.String(),
		}

		p.buf.Reset()

		return (*Parser).right, nil
	}

	if strings.ContainsRune(p.Comments, r) {
		return nil, p.parseError("Unexpected rune: " + string(r))
	}

	p.buf.WriteRune(r)

	return (*Parser).left, nil
}

func (p *Parser) right(r rune) (stateFunc, error) {
	switch r {
	case '\n':
		p.t.(*SettingToken).Right = p.buf.String()

		return nil, nil
	}

	if strings.ContainsRune(p.Comments, r) {
		p.r.UnreadRune()

		p.t.(*SettingToken).Right = p.buf.String()

		return nil, nil
	}

	p.buf.WriteRune(r)

	return (*Parser).right, nil
}

// Next reads the next token from the underlying io.Reader. It returns
// an io.EOF when there are no more tokens available.
func (p *Parser) Next() (Token, error) {
	if p.eof {
		return nil, io.EOF
	}

	p.t = nil

	state := (*Parser).start

	for {
		r, _, err := p.r.ReadRune()
		if err != nil {
			if err == io.EOF {
				p.eof = true
				r = '\n'
			} else {
				return nil, err
			}
		}

		if r == '\n' {
			p.line++
			p.pos = 0
		} else {
			p.pos++
		}

		state, err = state(p, r)
		if err != nil {
			return nil, err
		}

		if (state == nil) || p.eof {
			break
		}
	}

	if p.eof && (p.t == nil) {
		return nil, io.EOF
	}

	return p.t, nil
}

// ParseError is returned by (*Parser).Next() if it encounters an error.
type ParseError struct {
	Line int
	Pos  int
	Err  string
}

func (p *Parser) parseError(msg string) error {
	return &ParseError{
		Line: p.line,
		Err:  msg,
	}
}

func (err *ParseError) Error() string {
	return fmt.Sprintf("%v:%v: %v", err.Line, err.Pos, err.Err)
}

type Token interface{}

// A SectionToken represents a section header. For example,
//
//	[Name]
type SectionToken struct {
	// Name is the name of the section.
	Name string

	start, end rune
}

// String recreates the original section token in the INI file.
func (t SectionToken) String() string {
	var buf bytes.Buffer

	buf.WriteRune(t.start)
	buf.WriteString(t.Name)
	buf.WriteRune(t.end)

	return buf.String()
}

// A SettingToken represents a setting. For example,
//
//	left=right
type SettingToken struct {
	// Left is the left-hand side of the setting assignment.
	Left string

	// Right is the right-hand side of the setting assignment.
	Right string

	equals rune
}

// String recreates the original setting token.
func (t SettingToken) String() string {
	var buf bytes.Buffer

	buf.WriteString(t.Left)
	buf.WriteRune(t.equals)
	buf.WriteString(t.Right)

	return buf.String()
}

// A CommentToken represents a comment.
type CommentToken struct {
	// Comment is the text of the comment, including any leading
	// whitespace.
	Comment string

	start rune
}

// String recreates the original comment token.
func (t CommentToken) String() string {
	var buf bytes.Buffer

	buf.WriteRune(t.start)
	buf.WriteString(t.Comment)

	return buf.String()
}
