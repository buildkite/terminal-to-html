package terminal

import "testing"

func assert(t *testing.T, outputB []byte, expected string) {
	output := string(outputB)

	if output != expected {
		t.Errorf("got %v, expected %v", output, expected)
	}
}

var emptyScreen = screen{style: &emptyStyle}

func TestScreenWriteToXY(t *testing.T) {
	s := emptyScreen
	s.write('a')

	s.x = 1
	s.y = 1
	s.write('b')

	s.x = 2
	s.y = 2
	s.write('c')

	assert(t, s.output(), "a\n b\n  c")
}
