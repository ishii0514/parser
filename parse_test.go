package parse

import (
	"fmt"
	"testing"
)

func TestParse(t *testing.T) {
	input := " 12 + 1 "
	expected := "parse.BinOpExpr{left:parse.NumExpr{literal:\"12\"}, operator:43, right:parse.NumExpr{literal:\"1\"}}"
	l := parse(input)
	if fmt.Sprintf("%#v", l.result) != expected {
		t.Errorf("acutual=[%#v],expected=[%s]", l.result, expected)
	}
}

func parse(input string) *LexerWrapper {
	l := LexerWrapper{l: lex(input)}
	yyParse(&l)
	return &l
}
