package terminal

import (
	"fmt"
	"testing"
)

func TestSanitizeURL(t *testing.T) {
	testCases := []struct {
		input string
		want  string
	}{
		// allowed schemes
		{input: "https://example.org/", want: "https://example.org/"},
		{input: "http://example.org/path?a=b&c=d#frag", want: "http://example.org/path?a=b&c=d#frag"},
		{input: "artifact://hello.txt", want: "artifact://hello.txt"},

		// host-relative URLs (no scheme, no host)
		{input: "/hello.txt", want: "/hello.txt"},
		{input: "hello.txt", want: "hello.txt"},
		{input: "hello.txt?a=b#frag", want: "hello.txt?a=b#frag"},

		// known-dangerous schemes
		{input: "javascript:alert(1)", want: "#"},

		// not-specifically-allow-listed schemes
		{input: "ftp://example.org/", want: "#"},
		{input: "tel:0123456789", want: "#"},
		{input: "entirelymadeup://default-deny/this-is-the-way", want: "#"},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%q -> %q", tc.input, tc.want), func(t *testing.T) {
			got := sanitizeURL(tc.input)
			if got != tc.want {
				t.Errorf("wanted %q -> %q, got %q", tc.input, tc.want, got)
			}
		})
	}
}
