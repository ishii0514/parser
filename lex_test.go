package parse

import (
	"testing"
)

func TestLexIdentifier(t *testing.T) {
	input := "abc  def"
	expects := []item{
		item{itemIdentifier, 0, "abc"},
		item{itemIdentifier, 5, "def"},
		item{itemEOF, len(input), ""},
	}
	check(t, input, expects)

	input = "あいう えお"
	expects = []item{
		item{itemIdentifier, 0, "あいう"},
		item{itemIdentifier, 10, "えお"},
		item{itemEOF, len(input), ""},
	}
	check(t, input, expects)
}

func TestLexNumber(t *testing.T) {
	input := "123 456"
	expects := []item{
		item{NUMBER, 0, "123"},
		item{NUMBER, 4, "456"},
		item{itemEOF, len(input), ""},
	}
	check(t, input, expects)

	input = "12.3 .456"
	expects = []item{
		item{NUMBER, 0, "12.3"},
		item{NUMBER, 5, ".456"},
		item{itemEOF, len(input), ""},
	}
	check(t, input, expects)
	input = "1.2.3 456."
	expects = []item{
		item{NUMBER, 0, "1.2"},
		item{NUMBER, 3, ".3"},
		item{NUMBER, 6, "456."},
		item{itemEOF, len(input), ""},
	}
	check(t, input, expects)

	input = "012 00.12e10 .12E+0 12.e-10"
	expects = []item{
		item{NUMBER, 0, "012"},
		item{NUMBER, 4, "00.12e10"},
		item{NUMBER, 13, ".12E+0"},
		item{NUMBER, 20, "12.e-10"},
		item{itemEOF, len(input), ""},
	}
	check(t, input, expects)

	input = "123 455t"
	expects = []item{
		item{NUMBER, 0, "123"},
		item{itemError, 4, "bad number syntax: \"455t\""},
	}
	check(t, input, expects)

	input = "1.2+456"
	expects = []item{
		item{NUMBER, 0, "1.2"},
		item{itemType('+'), 3, "+"},
		item{NUMBER, 4, "456"},
		item{itemEOF, len(input), ""},
	}
	check(t, input, expects)

	input = "12+4"
	expects = []item{
		item{NUMBER, 0, "12"},
		item{itemType('+'), 2, "+"},
		item{NUMBER, 3, "4"},
		item{itemEOF, len(input), ""},
	}
	check(t, input, expects)

	input = "  12  +  4  "
	expects = []item{
		item{NUMBER, 2, "12"},
		item{itemType('+'), 6, "+"},
		item{NUMBER, 9, "4"},
		item{itemEOF, len(input), ""},
	}
	check(t, input, expects)

	input = "4 SELECT txt"
	expects = []item{
		item{NUMBER, 0, "4"},
		item{itemType(SELECT), 2, "SELECT"},
		item{itemIdentifier, 9, "txt"},
		item{itemEOF, len(input), ""},
	}
	check(t, input, expects)

}

func check(t *testing.T, input string, expects []item) {
	l := lex(input)
	for _, expected := range expects {
		compare(t, l.nextItem(), expected)
	}
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
