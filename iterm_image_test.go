package terminal

import (
	"reflect"
	"testing"
)

var errorCases = []struct {
	name     string
	input    string
	expected string
}{
	{
		`sequence does not begin with 1337;File=`,
		"foobar",
		`expected sequence to start with 1337;File=, got "foobar" instead`,
	}, {
		`sequence beginning error is cropped`,
		"123456789012345678901234567890123456789012345678901234567890",
		`expected sequence to start with 1337;File=, got "1234567890" instead`,
	}, {
		`sequence does not have a content part and a arguments part`,
		"1337;File=foobar",
		`expected sequence to have one arguments part and one content part, got 1 part(s)`,
	}, {
		`sequence has too many parts`,
		"1337;File=foobar:baz:foo",
		`expected sequence to have one arguments part and one content part, got 3 part(s)`,
	}, {
		`content part is not valid Base64`,
		"1337;File=foobar:!!!!!",
		`expected content part to be valid Base64`,
	}, {
		`image name is missing`,
		"1337;File=foobar:AA==",
		`name= argument not supplied, required to determine content type`,
	}, {
		`can't determine content type`,
		"1337;File=name=foo.baz:AA==",
		`can't determine content type for "foo.baz"`,
	}, {
		`no image content`,
		"1337;File=name=foo.jpg:",
		`image content missing`,
	},
}

var validCases = []struct {
	name     string
	input    string
	expected *itermImage
}{
	{
		`image with name, content & inline`,
		"1337;File=name=foo.gif;inline=1:AA==",
		&itermImage{alt: "foo.gif", content: "AA==", content_type: "image/gif"},
	}, {
		`image without inline=1 does not render`,
		"1337;File=name=foo.gif:AA==",
		nil,
	}, {
		`adapts content type based on image name`,
		"1337;File=name=foo.jpg;inline=1:AA==",
		&itermImage{alt: "foo.jpg", content: "AA==", content_type: "image/jpeg"},
	}, {
		`handles width & height`,
		"1337;File=name=foo.jpg;width=100%;height=50px;inline=1:AA==",
		&itermImage{alt: "foo.jpg", content: "AA==", content_type: "image/jpeg", width: "100%", height: "50px"},
	}, {
		`protects against XSS in image name, width & height by stripping brackets & quotes`,
		`1337;File=name=foo<.gif;width="100%;height='50px>;inline=1:AA==`,
		&itermImage{alt: "foo.gif", content: "AA==", content_type: "image/gif", width: "100%", height: "50px"},
	},
}

func TestErrorCases(t *testing.T) {
	for _, c := range errorCases {
		img, err := parseItermImageSequence(c.input)
		if img != nil {
			t.Errorf("%s\ninput\t\t%q\nexpected no image, received %+v", c.name, c.input, img)
		} else if err.Error() != c.expected {
			t.Errorf("%s\ninput\t\t%q\nreceived\t%q\nexpected\t%q", c.name, c.input, err.Error(), c.expected)
		}
	}
}

func TestImageCases(t *testing.T) {
	for _, c := range validCases {
		img, err := parseItermImageSequence(c.input)
		if err != nil {
			t.Errorf("%s\ninput\t\t%q\nexpected no error, received %s", c.name, c.input, err.Error())
		} else if !reflect.DeepEqual(img, c.expected) {
			t.Errorf("%s\ninput\t\t%q\nreceived\t%+v\nexpected\t%+v", c.name, c.input, img, c.expected)
		}
	}
}
