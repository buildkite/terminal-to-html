package terminal

import (
	"fmt"
	"testing"
)

func TestParseSimpleXY(t *testing.T) {
	s := parsedScreen("hello")
	if err := assertTextXY(t, s, "hello", 5, 0); err != nil {
		t.Error(err)
	}
}

func TestParseAfterCursorMovement(t *testing.T) {
	s := parsedScreen("hello\x1b[4D!")
	if err := assertTextXY(t, s, "h!llo", 2, 0); err != nil {
		t.Error(err)
	}
}

func TestParseAfterOverwriteAndClearToEndOfLine(t *testing.T) {
	s := parsedScreen("hello\x1b[4Di!\x1b[0K")
	if err := assertTextXY(t, s, "hi!", 3, 0); err != nil {
		t.Error(err)
	}
}

// Application Program Command should be zero-width
func TestParseZeroWidthAPC(t *testing.T) {
	s := parsedScreen("\x1b_bk;t=0\x07")
	if err := assertTextXY(t, s, "", 0, 0); err != nil {
		t.Error(err)
	}
}

// Application Program Command can be followed by normal text
func TestParseAPCPrefix(t *testing.T) {
	s := parsedScreen("\x1b_bk;t=0\x07hello")
	if err := assertTextXY(t, s, "hello", 5, 0); err != nil {
		t.Error(err)
	}
}

// Application Program Command should be zero-width for cursor movement
func TestParseXYAfterCursorMovementThroughBuildkiteTimestampAPC(t *testing.T) {
	s := parsedScreen("hel\x1b_bk;t=0\x07lo\x1b[4D3")
	if err := assertTextXY(t, s, "h3llo", 2, 0); err != nil {
		t.Error(err)
	}
}

// ----------------------------------------

func parsedScreen(data string) *screen {
	s := &screen{}
	parseANSIToScreen(s, []byte(data))
	return s
}

func assertXY(t *testing.T, s *screen, x, y int) error {
	if s.x != x {
		return fmt.Errorf("expected screen.x == %d, got %d", x, s.x)
	}
	if s.y != y {
		return fmt.Errorf("expected screen.y == %d, got %d", y, s.y)
	}
	return nil
}

func assertText(t *testing.T, s *screen, expected string) error {
	if actual := s.asPlainText(); actual != expected {
		return fmt.Errorf("expected text %q, got %q", expected, actual)
	}
	return nil
}

func assertTextXY(t *testing.T, s *screen, expected string, x, y int) error {
	if err := assertXY(t, s, x, y); err != nil {
		return err
	}
	if err := assertText(t, s, expected); err != nil {
		return err
	}
	return nil
}
