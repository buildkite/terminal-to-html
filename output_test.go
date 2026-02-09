package terminal

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestScreenLineAsHTML_Interleaving(t *testing.T) {
	// ANSI escapes can come in any order, but it is invalid to interleave HTML
	// tags.

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "a span /a /span",
			input: "five \x1b]8;;http://example.com\x1b\\six \x1b[35mseven \x1b]8;;\x1b\\eight\x1b[0m",
			want:  `five <a href="http://example.com">six <span class="term-fg35">seven </span></a><span class="term-fg35">eight</span>` + "\n",
		},
		{
			name:  "span a /span /a",
			input: "five \x1b[35msix \x1b]8;;http://example.com\x1b\\seven \x1b[0meight\x1b]8;;\x1b\\",
			want:  `five <span class="term-fg35">six <a href="http://example.com">seven </a></span><a href="http://example.com">eight</a>` + "\n",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			s, err := NewScreen()
			if err != nil {
				t.Fatalf("NewScreen() = %v", err)
			}
			s.Write([]byte(test.input))
			if len(s.screen) != 1 {
				t.Fatalf("len(s.screen) = %d, want 1", len(s.screen))
			}

			got := lineToHTML(s.screen[:1], true)
			if diff := cmp.Diff(got, test.want); diff != "" {
				t.Errorf("lineToHTML(s.screen[:1], true) diff (-got +want):\n%s", diff)
			}
		})
	}
}
