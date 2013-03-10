package ini

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
	"unicode"
)

type stateFunc func(p *Parser, r rune) (stateFunc, error)

type Parser struct {
	r    *bufio.Reader
	line int
	pos  int
	eof  bool

	buf bytes.Buffer
	t   Token

	Comments string

	SectionStart rune
	SectionEnd   rune

	Equals rune
}

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

		return nil, nil
	}

	p.buf.WriteRune(r)

	return (*Parser).right, nil
}

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

type SectionToken struct {
	start, end rune

	Name string
}

func (t SectionToken) String() string {
	var buf bytes.Buffer

	buf.WriteRune(t.start)
	buf.WriteString(t.Name)
	buf.WriteRune(t.end)

	return buf.String()
}

type SettingToken struct {
	equals rune

	Left  string
	Right string
}

func (t SettingToken) String() string {
	var buf bytes.Buffer

	buf.WriteString(t.Left)
	buf.WriteRune(t.equals)
	buf.WriteString(t.Right)

	return buf.String()
}

type CommentToken struct {
	start rune

	Comment string
}

func (t CommentToken) String() string {
	var buf bytes.Buffer

	buf.WriteRune(t.start)
	buf.WriteString(t.Comment)

	return buf.String()
}
