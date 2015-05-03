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
		`sequence does not begin with 1337;File= or 1338;`,
		"foobar",
		`expected sequence to start with 1337;File= or 1338;, got "foobar" instead`,
	}, {
		`sequence beginning error is cropped`,
		"123456789012345678901234567890123456789012345678901234567890",
		`expected sequence to start with 1337;File= or 1338;, got "1234567890" instead`,
	}, {
		`1337: sequence does not have a content part and a arguments part`,
		"1337;File=foobar",
		`expected sequence to have one arguments part and one content part, got 1 part(s)`,
	}, {
		`1337: sequence has too many parts`,
		"1337;File=foobar:baz:foo",
		`expected sequence to have one arguments part and one content part, got 3 part(s)`,
	}, {
		`1337: content part is not valid Base64`,
		"1337;File=foobar:!!!!!",
		`expected content part to be valid Base64`,
	}, {
		`1337: image name is missing`,
		"1337;File=foobar:AA==",
		`name= argument not supplied, required to determine content type`,
	}, {
		`1337: invalid base64 encoding`,
		"1337;File=name=foo.baz:AA==",
		`name= value of "foo.baz" is not valid base64`,
	}, {
		`1337: can't determine content type`,
		"1337;File=name=" + base64Encode("foo.baz") + ":AA==",
		`can't determine content type for "foo.baz"`,
	}, {
		`1337: no image content`,
		"1337;File=name=foo.jpg:",
		`image content missing`,
	}, {
		`1338: url missing`,
		"1338;",
		`url= argument not supplied`,
	},
}

var validCases = []struct {
	name     string
	input    string
	expected *image
}{
	{
		`1337: image with name, content & inline`,
		`1337;File=name=` + base64Encode("foo.gif") + `;inline=1:AA==`,
		&image{filename: "foo.gif", content: "AA==", content_type: "image/gif", iTerm: true},
	}, {
		`1337: image without inline=1 does not render`,
		`1337;File=name=` + base64Encode("foo.gif") + `:AA==`,
		nil,
	}, {
		`1337: adapts content type based on image name`,
		`1337;File=name=` + base64Encode("foo.jpg") + `;inline=1:AA==`,
		&image{filename: "foo.jpg", content: "AA==", content_type: "image/jpeg", iTerm: true},
	}, {
		`1337: handles width & height`,
		`1337;File=name=` + base64Encode("foo.jpg") + `;width=100%;height=50px;inline=1:AA==`,
		&image{filename: "foo.jpg", content: "AA==", content_type: "image/jpeg", width: "100%", height: "50px", iTerm: true},
	}, {
		`1337: protects against XSS in image name, width & height by stripping brackets & quotes`,
		`1337;File=name=` + base64Encode("foo.gif") + `;width="100%;height='50px>;inline=1:AA==`,
		&image{filename: "foo.gif", content: "AA==", content_type: "image/gif", width: "100%", height: "50px", iTerm: true},
	}, {
		`1337: converts width & height without percent or px to em`,
		`1337;File=name=` + base64Encode("foo.jpg") + `;width=1;height=5;inline=1:AA==`,
		&image{filename: "foo.jpg", content: "AA==", content_type: "image/jpeg", width: "1em", height: "5em", iTerm: true},
	}, {
		`1337: malfored arguments are silently ignored`,
		`1337;File=name=` + base64Encode("foo.gif") + `;inline=1;sdfsdfs;====ddd;herp=derps:AA==`,
		&image{filename: "foo.gif", content: "AA==", content_type: "image/gif", iTerm: true},
	}, {
		`1338: image with filename`,
		"1338;url=tmp/foo.gif",
		&image{filename: "tmp/foo.gif"},
	}, {
		`1338: image with filename containing an escaped ;`,
		"1338;url=tmp/foo\\;bar.gif",
		&image{filename: "tmp/foo;bar.gif"},
	}, {
		`1338: image with filename, width, height & alt tag`,
		"1338;url=foo.gif;width=50px;height=50px;alt=foo gif",
		&image{filename: "foo.gif", width: "50px", height: "50px", alt: "foo gif"},
	},
}

func TestErrorCases(t *testing.T) {
	for _, c := range errorCases {
		img, err := parseImageSequence(c.input)
		if img != nil {
			t.Errorf("%s\ninput\t\t%q\nexpected no image, received %+v", c.name, c.input, img)
		} else if err.Error() != c.expected {
			t.Errorf("%s\ninput\t\t%q\nreceived\t%q\nexpected\t%q", c.name, c.input, err.Error(), c.expected)
		}
	}
}

func TestImageCases(t *testing.T) {
	for _, c := range validCases {
		img, err := parseImageSequence(c.input)
		if err != nil {
			t.Errorf("%s\ninput\t\t%q\nexpected no error, received %s", c.name, c.input, err.Error())
		} else if !reflect.DeepEqual(img, c.expected) {
			t.Errorf("%s\ninput\t\t%q\nreceived\t%+v\nexpected\t%+v", c.name, c.input, img, c.expected)
		}
	}
}
