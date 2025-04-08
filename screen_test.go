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
		s, err := NewScreen(WithMaxSize(0, test.maxlines))
		got := []string{}
		s.ScrollOutFunc = func(line string) { got = append(got, line) }
		_ = s.currentLineForWriting()
		s.Write([]byte(test.input))
		if err != nil {
			t.Errorf("Failure for '%s':\nNewScreen returned an error: %s", test.name, err)
		}
		_ = s.currentLineForWriting()

		if diff := cmp.Diff(got, test.want); diff != "" {
			t.Errorf(
				"Failure for '%s':\nscrolledOutFunc sequence of parameters diff (-got +want):\n%s",
				test.name,
				diff,
			)
		}
	}
}
