package parse

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

func (i item) String() string {
	switch {
	case i.typ == itemEOF:
		return "EOF"
	case i.typ == itemError:
		return i.val
	case len(i.val) > 10:
		return fmt.Sprintf("%.10q...", i.val)
	}
	return fmt.Sprintf("%q", i.val)
}

const (
	itemError itemType = iota
	itemString
	itemNumber
	itemIdentifier
	itemPeriod
	itemArithmeticOperator
	itemTerminate
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
	//channelからitemを取り出す。
	item := <-l.items
	l.lastPos = item.pos
	return item
}

func lex(input string) *lexer {
	//lexerを生成して返す。
	//l.run()をgoroutineで呼び出してスキャンを実施する。
	l := &lexer{
		input: input,
		items: make(chan item),
	}
	go l.run()
	return l
}

func (l *lexer) run() {
	//lの状態を管理する。
	//lexBaseから始まって、l.state(l)で現在の状況のl.stateにする。
	for l.state = lexBase; l.state != nil; {
		l.state = l.state(l)
	}
}

func (l *lexer) next() rune {
	//一文字取り出す
	if int(l.pos) >= len(l.input) {
		l.width = 0
		return eof
	}
	r, w := utf8.DecodeRuneInString(l.input[l.pos:])
	l.width = w
	l.pos += w
	return r
}

func (l *lexer) peek() rune {
	//一文字取り出すが、l.posは変えない。
	r := l.next()
	l.backup()
	return r
}

func (l *lexer) backup() {
	//posを一文字分戻す。
	l.pos -= l.width
}

func (l *lexer) ignore() {
	l.start = l.pos
}

func (l *lexer) emit(t itemType) {
	//指定されたitemTypeのitemを生成してitems channelに送る。
	//pos,valueは、lのコンテキストから作成
	l.items <- item{t, l.start, l.input[l.start:l.pos]}
	l.start = l.pos
}

func (l *lexer) accept(valid string) bool {
	//validに次の文字が含まれていたらtrue
	//含まれていなかったら、戻してfalse
	if strings.IndexRune(valid, l.next()) >= 0 {
		return true
	}
	l.backup()
	return false
}

func (l *lexer) acceptRun(valid string) {
	//validの文字が登場しなくなるまで進める。
	for strings.IndexRune(valid, l.next()) >= 0 {
	}
	l.backup()
}

func (l *lexer) errorf(format string, args ...interface{}) stateFn {
	//error itemをchannelに送る。stateFnとしてnilを返す。
	l.items <- item{itemError, l.start, fmt.Sprintf(format, args...)}
	return nil
}

const (
	doubleQuote  = `"`
	singleQuote  = `'`
	lineComment  = "--"
	leftComment  = "/*"
	rightComment = "*/"
)

func lexBase(l *lexer) stateFn {
	if strings.HasPrefix(l.input[l.pos:], lineComment) {
		return lexLineComment
	} else if strings.HasPrefix(l.input[l.pos:], leftComment) {
		return lexBlockComment
	}
	switch r := l.next(); {
	case r == eof:
		if l.pos > l.start {
			return l.errorf("illegal item")
		}
		l.emit(itemEOF)
		return nil
	case isSpace(r):
		l.backup()
		return lexWhitespace
	case r == '\'':
		return lexString
	case r == '.':
		if !unicode.IsDigit(l.peek()) {
			l.emit(itemPeriod)
			return lexBase
		}
		fallthrough
	case ('0' <= r && r <= '9'):
		l.backup()
		return lexNumber
	case r == '-' || r == '+':
		//+-はここでは演算子として扱う
		//TODO 数値の符号かどうかは構文解析で考える
		l.emit(itemArithmeticOperator)
		//l.emit(itemType(r))
	case r == '*' || r == '/':
		l.emit(itemArithmeticOperator)
	case r == ';':
		l.emit(itemTerminate)
	case isIdentifierBegin(r):
		l.backup()
		return lexIdentifiler
	}
	return lexBase
}

func lexString(l *lexer) stateFn {
	for {
		switch r := l.next(); {
		case r == '\'':
			if l.accept(singleQuote) {
				continue
			}
			l.emit(itemString)
			return lexBase
		case unicode.IsControl(r):
			return l.errorf("cannot contain control characters in strings")
		case r == eof:
			return l.errorf("unclosed string")
		}
	}
}

func lexNumber(l *lexer) stateFn {
	if !l.scanNumber() {
		return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
	}
	//imaginary is not support.
	/*if sign := l.peek(); sign == '+' || sign == '-' {
		if !l.scanNumber() || l.input[l.pos-1] != 'i' {
			return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
		}
		l.emit(itemComplex)
	} else {
		l.emit(itemNumber)
	}*/
	l.emit(itemNumber)
	return lexBase
}

func (l *lexer) scanNumber() bool {
	//l.accept("+-")
	digits := "0123456789"
	//数値の存否
	exists_int := l.accept(digits)
	exists_decimal := false

	l.acceptRun(digits)
	if l.accept(".") {
		exists_decimal = l.accept(digits)
		l.acceptRun(digits)
	}
	if !exists_int && !exists_decimal {
		return false
	}
	if l.accept("eE") {
		l.accept("+-")
		if !l.accept(digits) {
			return false
		}
		l.acceptRun(digits)
	}
	//imaginary is not support.
	//l.accept("i")
	if isAlphaNumeric(l.peek()) {
		l.next()
		return false
	}
	return true
}

func lexWhitespace(l *lexer) stateFn {
	for isSpace(l.next()) {
	}
	l.backup()
	//l.emit(itemWhitespace)
	l.ignore()
	return lexBase
}

func lexIdentifiler(l *lexer) stateFn {
	for isAlphaNumeric(l.next()) {
	}
	l.backup()
	l.emit(itemIdentifier)
	return lexBase
}

func lexLineComment(l *lexer) stateFn {
	for {
		r := l.next()
		if r == '\n' || r == eof {
			if l.pos > l.start {
				l.emit(itemLineComment)
			}
			if r == eof {
				l.emit(itemEOF)
				return nil
			}
			return lexBase
		}
	}
}

func lexBlockComment(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], rightComment) {
			l.pos += len(rightComment)
			l.emit(itemBlockComment)
			return lexBase
		}
		if l.next() == eof {
			break
		}
	}
	if l.pos > l.start {
		l.errorf("unclosed brock comment")
	}
	l.emit(itemEOF)
	return nil
}

func isSpace(r rune) bool {
	return r == ' ' || r == '\t'
}

func isEndOfLine(r rune) bool {
	return r == '\r' || r == '\n'
}

//reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return isIdentifierBegin(r) || unicode.IsDigit(r)
}

//識別子の戦闘かどうか
func isIdentifierBegin(r rune) bool {
	return 'a' <= r && r <= 'z' || 'A' <= r && r <= 'Z' || r == '_' || r >= 0x80 && unicode.IsLetter(r)
}
