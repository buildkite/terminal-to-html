package terminal

import "testing"

// There's a bit of trickery involved in parsing 24-bit color
// sequences that caused them to be silently ignored in the
// past. These tests assert that they are actually applied.
func TestColor24BitIsApplied(t *testing.T) {
	const (
		r, g, b = 100, 150, 200
		wantRGB = uint32(r<<16 | g<<8 | b) // 0x6496C8
	)

	t.Run("foreground", func(t *testing.T) {
		s := style(0).color([]string{"38", "2", "100", "150", "200"})
		if got := s.fgColorType(); got != color24Bit {
			t.Errorf("fgColorType() = %d, want %d (color24Bit)", got, color24Bit)
		}
		if got := s.fgColor(); got != wantRGB {
			t.Errorf("fgColor() = %#06x, want %#06x", got, wantRGB)
		}
	})

	t.Run("background", func(t *testing.T) {
		s := style(0).color([]string{"48", "2", "100", "150", "200"})
		if got := s.bgColorType(); got != color24Bit {
			t.Errorf("bgColorType() = %d, want %d (color24Bit)", got, color24Bit)
		}
		if got := s.bgColor(); got != wantRGB {
			t.Errorf("bgColor() = %#06x, want %#06x", got, wantRGB)
		}
	})
}
