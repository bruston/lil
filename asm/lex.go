package asm

import (
	"bufio"
	"bytes"
	"io"
	"unicode"
	"unicode/utf8"
)

const (
	ItemIdentifier = iota
	ItemStringLit
	ItemNumLit
	ItemComma
	ItemLabel
	ItemVar
)

type Item struct {
	Type  int
	Value string
	Line  int
	Pos   int
}

type Lexer struct {
	src     *bufio.Reader
	buf     *bytes.Buffer
	current Item
	last    rune
	line    int
	pos     int
	err     error
}

func NewLexer(r io.Reader) *Lexer {
	return &Lexer{
		src:  bufio.NewReader(r),
		buf:  bytes.NewBuffer(make([]byte, 0, 1024)),
		line: 1,
	}
}

func (l *Lexer) Scanning() bool {
	l.scan()
	return l.err == nil
}

func (l *Lexer) read() (rune, error) {
	ch, size, err := l.src.ReadRune()
	if err != nil {
		return 0, err
	}
	l.pos += size
	l.last = ch
	if ch == '\n' {
		l.line++
	}
	return ch, nil
}

func (l *Lexer) unread() {
	l.src.UnreadRune()
	if l.last == '\n' {
		l.line--
	}
	l.pos -= utf8.RuneLen(l.last)
	l.last = 0 // shouldn't matter, we only ever back up once
}

func (l *Lexer) peek() (rune, error) {
	ch, _, err := l.src.ReadRune()
	l.src.UnreadRune()
	return ch, err
}

func (l *Lexer) skipSpace() {
	for {
		ch, _ := l.read()
		if !unicode.IsSpace(ch) {
			l.unread()
			break
		}
	}
}

func (l *Lexer) scanIdent() (Item, error) {
	defer l.buf.Reset()
	line, pos := l.line, l.pos
	for {
		ch, err := l.read()
		if err != nil {
			if l.buf.Len() == 0 {
				return Item{}, err
			}
			l.unread()
			break
		}
		if unicode.IsSpace(ch) {
			l.unread()
			break
		}
		l.buf.WriteRune(ch)
	}
	return Item{Type: ItemIdentifier, Value: l.buf.String(), Line: line, Pos: pos}, nil
}

func (l *Lexer) scanNumber() (Item, error) {
	defer l.buf.Reset()
	line, pos := l.line, l.pos
	for {
		ch, err := l.read()
		if err != nil {
			if l.buf.Len() == 0 {
				return Item{}, err
			}
			l.unread()
			break
		}
		if !unicode.IsDigit(ch) {
			l.unread()
			break
		}
		l.buf.WriteRune(ch)
	}
	return Item{Type: ItemNumLit, Value: l.buf.String(), Line: line, Pos: pos}, nil
}

func (l *Lexer) scanStringLit() (Item, error) {
	return Item{}, nil
}

func (l *Lexer) scan() {
	l.skipSpace()
	ch, err := l.peek()
	if err != nil {
		l.err = err
		return
	}
	if ch == ':' {
		l.current, l.err = l.scanIdent()
		l.current.Type = ItemLabel
		return
	}
	if unicode.IsLetter(ch) {
		l.current, l.err = l.scanIdent()
		if l.current.Value == "var" {
			l.current.Type = ItemVar
		}
		return
	}
	if unicode.IsDigit(ch) || ch == '-' {
		l.current, l.err = l.scanNumber()
		return
	}
}

func (l *Lexer) Item() Item { return l.current }

func (l *Lexer) Err() error {
	if l.err == nil || l.err == io.EOF {
		return nil
	}
	return l.err
}
