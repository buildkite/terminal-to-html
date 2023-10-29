package terminal

import (
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseBuildkiteAPC(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		parser     *parser
		sequence   string
		wantData   map[string]string
		wantLastTS int64
	}{
		{
			name:       "t field",
			parser:     &parser{},
			sequence:   "bk;t=12345",
			wantData:   map[string]string{"t": "12345"},
			wantLastTS: 12345,
		},
		{
			name:       "dt becomes t",
			parser:     &parser{lastTimestamp: 12345},
			sequence:   "bk;dt=123",
			wantData:   map[string]string{"t": "12468"},
			wantLastTS: 12468,
		},
		{
			name:       "other field parsed",
			parser:     &parser{},
			sequence:   "bk;t=12345;p=np",
			wantData:   map[string]string{"t": "12345", "p": "np"},
			wantLastTS: 12345,
		},
	}

	for _, test := range tests {
		test := test
		t.Run(test.sequence, func(t *testing.T) {
			t.Parallel()

			got, err := test.parser.parseBuildkiteAPC(test.sequence)
			if err != nil {
				t.Fatalf("parseBuildkiteAPC(%q) error = %v", test.sequence, err)
			}
			if diff := cmp.Diff(got, test.wantData); diff != "" {
				t.Errorf("parsed buildkite APC data diff (-got +want):\n%s", diff)
			}
			if got, want := test.parser.lastTimestamp, test.wantLastTS; got != want {
				t.Errorf("parser.lastTimestamp = %d, want %d", got, want)
			}
		})

	}
}
