package parse

import (
	"testing"
)

func TestLex(t *testing.T) {
	input := "abc  def"
	expects := []item{
		item{itemIdentifier, 0, "abc"},
		item{itemWhitespace, 3, "  "},
		item{itemIdentifier, 5, "def"},
	}
	check(t, input, expects)

	input = "あいう えお"
	expects = []item{
		item{itemIdentifier, 0, "あいう"},
		item{itemWhitespace, 9, " "},
		item{itemIdentifier, 10, "えお"},
	}
	check(t, input, expects)
}

func check(t *testing.T, input string, expects []item) {
	l := lex(input)
	for _, expected := range expects {
		compare(t, l.nextItem(), expected)
	}
	compare(t, l.nextItem(), item{itemEOF, len(input), ""})
}

func compare(t *testing.T, actual item, expected item) {
	if actual != expected {
		t.Errorf("actual=[typ:%d,pos=%d,val=%q], expected=[typ=%d,pos=%d,val=%q]", actual.typ, actual.pos, actual.val, expected.typ, expected.pos, expected.val)
	}
}

func TestIsAlphaNumeric(t *testing.T) {
	if isAlphaNumeric('_') == false {
		t.Errorf("error %q", '_')
	}
	if isAlphaNumeric('%') {
		t.Errorf("error %q", '%')
	}

	if isAlphaNumeric('1') == false {
		t.Errorf("error %q/", '1')
	}
	if isAlphaNumeric(' ') {
		t.Errorf("error %q", ' ')
	}

	if isAlphaNumeric('あ') == false {
		t.Errorf("error %q", 'あ')
	}
}
