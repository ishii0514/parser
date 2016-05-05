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
	//lexTextから始まって、l.state(l)で現在の状況のl.stateにする。
	for l.state = lexText; l.state != nil; {
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

func lexText(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], doubleQuote) {
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexString
		} else if strings.HasPrefix(l.input[l.pos:], lineComment) {
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexLineComment
		} else if strings.HasPrefix(l.input[l.pos:], leftComment) {
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexBlockComment
		}

		r := l.next()
		if unicode.IsSpace(r) {
			l.backup()
			if l.pos > l.start {
				l.emit(itemText)
			}
			return lexWhitespace
		} else if r == eof {
			break
		}
	}

	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF)
	return nil
}

func lexString(l *lexer) stateFn {
	l.next()
	for {
		switch r := l.next(); {
		case r == '"':
			l.emit(itemString)
			return lexText
		case r == '\\':
			if l.accept(`"\/bfnrt`) {
				break
			} else if r := l.next(); r == 'u' {
				for i := 0; i < 4; i++ {
					if !l.accept("0123456789ABCDEFabcdef") {
						return l.errorf("expected 4 hexadecimal digits")
					}
				}
			} else {
				return l.errorf("unsupported escape character")
			}
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
	if sign := l.peek(); sign == '+' || sign == '-' {
		if !l.scanNumber() || l.input[l.pos-1] != 'i' {
			return l.errorf("bad number syntax: %q", l.input[l.start:l.pos])
		}
		//l.emit(itemComplex)
	} else {
		l.emit(itemNumber)
	}
	return lexText
}

func (l *lexer) scanNumber() bool {
	l.accept("+-")
	digits := "0123456789"
	l.acceptRun(digits)
	if l.accept(".") {
		l.acceptRun(digits)
	}
	if l.accept("eE") {
		l.accept("+-")
		l.acceptRun(digits)
	}
	l.accept("i")
	if isAlphaNumeric(l.peek()) {
		l.next()
		return false
	}
	return true
}

func lexWhitespace(l *lexer) stateFn {
	for unicode.IsSpace(l.next()) {
	}
	l.backup()
	l.emit(itemWhitespace)
	return lexText
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
			return lexText
		}
	}
}

func lexBlockComment(l *lexer) stateFn {
	for {
		if strings.HasPrefix(l.input[l.pos:], rightComment) {
			l.pos += len(rightComment)
			if l.pos > l.start {
				l.emit(itemBlockComment)
			}
			return lexText
		}
		if l.next() == eof {
			break
		}
	}
	if l.pos > l.start {
		l.emit(itemText)
	}
	l.emit(itemEOF)
	return nil
}

//reports whether r is an alphabetic, digit, or underscore.
func isAlphaNumeric(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}
