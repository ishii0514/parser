package parser

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type item struct {
	typ itemType
	pos int
	val string
}

type itemType int

func (i item) String() {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	}
	return fmt.Sprintf("%q", i.val)
}

const (
	itemError itemTuype = iota
	itemString
	itemText
	itemBlockComment
	itemLineComment
	itemWhitespace
	itemEOF
)

const eof = -1

type stateFn func(*lexer) stateFn

type lexer struct {
	input   string
	state   stateFn
	pos     int
	start   int
	width   int
	lastPos int
	items   chan item
}

func (l *lexer) nextItem() item {
	item := <-l.items
	l.lastPos = item.pos
	return item
}
