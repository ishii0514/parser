package parse

import (
	"testing"
)

func TestLex(t *testing.T) {
	l := lex("abc  def")
	actual := l.nextItem()
	expected := item{itemText, 0, "abc"}
	compare(t, actual, expected)

	actual = l.nextItem()
	expected = item{itemWhitespace, 3, "  "}
	compare(t, actual, expected)

	actual = l.nextItem()
	expected = item{itemText, 5, "def"}
	compare(t, actual, expected)
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

}
