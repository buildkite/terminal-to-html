package terminal

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

var currentLineForWritingTestCases = []struct {
	name     string
	input    string
	want     []string
	maxlines int
}{
	{
		name: "Test no index out of range panic",
		input: "\n",
		want: []string{"&nbsp;\n"},
		maxlines: 1,
	},
	{
		name: "Test scroll out first line",
		input: "a\n",
		want: []string{"a\n"},
		maxlines: 1,
	},
	{
		name: "Test scroll out several lines",
		input: "a\nb\nc\nd",
		want: []string{"a\n", "b\n"},
		maxlines: 2,
	},
}

func TestCurrentLineForWriting(t *testing.T) {
	for _, test := range currentLineForWritingTestCases {
		t.Run(test.name, func(t *testing.T) {
			s, err := NewScreen(WithMaxSize(0, test.maxlines))
			if err != nil {
				t.Fatalf("NewScreen(WithMaxSize(0, %d)) error: %s", test.maxlines, err)
			}
			got := []string{}
			s.ScrollOutFunc = func(line string) { got = append(got, line) }
			_ = s.currentLineForWriting()
			s.Write([]byte(test.input))
			_ = s.currentLineForWriting()

			if diff := cmp.Diff(got, test.want); diff != "" {
				t.Errorf("scrolledOutFunc sequence of parameters diff (-got +want):\n%s", diff)
			}
		})
	}
}
