package terminal

import (
	"encoding/base64"
	"fmt"
	"mime"
	"strings"
)

type image struct {
	filename     string
	alt          string
	content_type string
	content      string
	height       string
	width        string
	iTerm        bool
}

func (i *image) asHTML() string {
	alt := i.alt
	if alt == "" {
		alt = i.filename
	}

	parts := []string{fmt.Sprintf(`alt="%s"`, alt)}

	if i.iTerm {
		parts = append(parts, fmt.Sprintf(`src="data:%s;base64,%s"`, i.content_type, i.content))
	} else {
		parts = append(parts, fmt.Sprintf(`src="%s"`, i.filename))
	}

	if i.width != "" {
		parts = append(parts, fmt.Sprintf(`width="%s"`, i.width))
	}
	if i.height != "" {
		parts = append(parts, fmt.Sprintf(`height="%s"`, i.height))
	}
	return fmt.Sprintf(`<img %s>`, strings.Join(parts, " "))
}

func parseImageSequence(sequence string) (*image, error) {
	// Expect 1337;File=name=1.gif;inline=1:BASE64

	arguments, content, err := splitAndVerifyImageSequence(sequence)
	if err != nil {
		return nil, err
	}

	arguments = strings.Map(htmlStripper, arguments)
	arguments = strings.Replace(arguments, `\;`, "\x00", -1)

	imageInline := false

	img := &image{content: content, iTerm: content != ""}

	for _, arg := range strings.Split(arguments, ";") {
		arg = strings.Replace(arg, "\x00", ";", -1) // reconstitute escaped semicolons
		argParts := strings.SplitN(arg, "=", 2)
		if len(argParts) != 2 {
			continue
		}
		key := argParts[0]
		val := argParts[1]
		switch strings.ToLower(key) {
		case "name":
			nameBytes, err := base64.StdEncoding.DecodeString(val)
			if err != nil {
				return nil, fmt.Errorf("name= value of %q is not valid base64", val)
			}
			img.filename = string(nameBytes)
			img.content_type = contentTypeForFile(string(nameBytes))
		case "url":
			img.filename = val
		case "inline":
			imageInline = val == "1"
		case "width":
			img.width = parseImageDimension(val)
		case "height":
			img.height = parseImageDimension(val)
		case "alt":
			img.alt = val
		}
	}

	if img.iTerm {
		if img.filename == "" {
			return nil, fmt.Errorf("name= argument not supplied, required to determine content type")
		}
		if img.content_type == "" {
			return nil, fmt.Errorf("can't determine content type for %q", img.filename)
		}
	} else {
		if img.filename == "" {
			return nil, fmt.Errorf("url= argument not supplied")
		}
	}

	if img.iTerm && !imageInline {
		// in iTerm2, if you don't specify inline=1, the image is merely downloaded
		// and not displayed.
		img = nil
	}
	return img, nil
}

func contentTypeForFile(filename string) string {
	dot := strings.LastIndex(filename, ".")
	if dot == -1 {
		return ""
	}
	return mime.TypeByExtension(filename[dot:])
}

func parseImageDimension(s string) string {
	s = strings.ToLower(s)
	if !strings.HasSuffix(s, "px") && !strings.HasSuffix(s, "%") {
		return s + "em"
	} else {
		return s
	}
}

func htmlStripper(r rune) rune {
	switch r {
	case '<', '>', '\'', '"':
		return -1
	default:
		return r
	}
}

func splitAndVerifyImageSequence(s string) (arguments string, content string, err error) {
	if strings.HasPrefix(s, "1338;") {
		// non-iTerm image, don't need to extract content
		return s[len("1338;"):], "", nil
	}

	prefixLen := len("1337;File=")
	if !strings.HasPrefix(s, "1337;File=") {
		if len(s) > prefixLen {
			s = s[:prefixLen] // Don't blow out our error output
		}
		return "", "", fmt.Errorf("expected sequence to start with 1337;File= or 1338;, got %q instead", s)
	}
	s = s[prefixLen:]

	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("expected sequence to have one arguments part and one content part, got %d part(s)", len(parts))
	}

	arguments = parts[0]
	content = parts[1]
	if len(content) == 0 {
		return "", "", fmt.Errorf("image content missing")
	}

	_, err = base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", "", fmt.Errorf("expected content part to be valid Base64")
	}

	return
}
