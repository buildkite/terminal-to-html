package terminal

import (
	"fmt"
	"strconv"
	"strings"
	"testing"
)

func TestParseSimpleXY(t *testing.T) {
	s := parsedScreen(t, "hello")
	if err := assertTextXY(s, "hello", 5, 0); err != nil {
		t.Error(err)
	}
}

func TestParseAfterCursorMovement(t *testing.T) {
	s := parsedScreen(t, "hello\x1b[4D!")
	if err := assertTextXY(s, "h!llo", 2, 0); err != nil {
		t.Error(err)
	}
}

func TestParseAfterOverwriteAndClearToEndOfLine(t *testing.T) {
	s := parsedScreen(t, "hello\x1b[4Di!\x1b[0K")
	if err := assertTextXY(s, "hi!", 3, 0); err != nil {
		t.Error(err)
	}
}

// Application Program Command should be zero-width
func TestParseZeroWidthAPC(t *testing.T) {
	s := parsedScreen(t, "\x1b_bk;t=0\x07")
	if err := assertTextXY(s, "", 0, 0); err != nil {
		t.Error(err)
	}
}

// Application Program Command can be followed by normal text
func TestParseAPCPrefix(t *testing.T) {
	s := parsedScreen(t, "\x1b_bk;t=0\x07hello")
	if err := assertTextXY(s, "hello", 5, 0); err != nil {
		t.Error(err)
	}
}

// Application Program Command can be terminated with ESC \
func TestParseAPCWithSTPrefix(t *testing.T) {
	s := parsedScreen(t, "\x1b_bk;t=0\x1b\\hello")
	if err := assertTextXY(s, "hello", 5, 0); err != nil {
		t.Error(err)
	}
}

// Application Program Command should be zero-width for cursor movement
func TestParseXYAfterCursorMovementThroughBuildkiteTimestampAPC(t *testing.T) {
	s := parsedScreen(t, "hel\x1b_bk;t=0\x07lo\x1b[4D3")
	if err := assertTextXY(s, "h3llo", 2, 0); err != nil {
		t.Error(err)
	}
}

// Operating System Command can be terminated with BEL
func TestParseOSCHyperlink(t *testing.T) {
	s := parsedScreen(t, "\x1b]8;;http://example.com/\x07hello")
	if err := assertTextXY(s, "hello", 5, 0); err != nil {
		t.Error(err)
	}
}

// Operating System Command can be terminated with ESC \
func TestParseOSCWithST(t *testing.T) {
	s := parsedScreen(t, "\x1b]8;;http://example.com/\x1b\\hello")
	if err := assertTextXY(s, "hello", 5, 0); err != nil {
		t.Error(err)
	}
}

func TestParseDECCursorSaveRestore(t *testing.T) {
	decsc := "\x1b7"
	decrc := "\x1b8"
	moveUpAndClearLine := csi(2, "A") + csi(2, "K") + csi(1, "G")

	s := parsedScreen(t, "one\ntwo\nthree\n"+decsc+moveUpAndClearLine+"overwrite\n"+decrc+"four\n")

	expected := strings.Join([]string{"one", "overwrite", "three", "four"}, "\n")
	if err := assertTextXY(s, expected, 0, 4); err != nil {
		t.Error(err)
	}
}

// ----------------------------------------

func parsedScreen(t *testing.T, data string) *Screen {
	s, err := NewScreen()
	if err != nil {
		t.Fatalf("NewScreen error: %v", err)
	}
	s.Write([]byte(data))
	return s
}

// csi is a test helper for CSI ANSI sequences.
// https://en.wikipedia.org/wiki/ANSI_escape_code#CSI_(Control_Sequence_Introducer)_sequences
func csi(n int, code string) string {
	return "\x1b[" + strconv.Itoa(n) + code
}

func assertXY(s *Screen, x, y int) error {
	if s.x != x {
		return fmt.Errorf("expected screen.x == %d, got %d", x, s.x)
	}
	if s.y != y {
		return fmt.Errorf("expected screen.y == %d, got %d", y, s.y)
	}
	return nil
}

func assertText(s *Screen, expected string) error {
	if actual := s.AsPlainText(); actual != expected {
		return fmt.Errorf("expected text %q, got %q", expected, actual)
	}
	return nil
}

func assertTextXY(s *Screen, expected string, x, y int) error {
	if err := assertXY(s, x, y); err != nil {
		return err
	}
	if err := assertText(s, expected); err != nil {
		return err
	}
	return nil
}
