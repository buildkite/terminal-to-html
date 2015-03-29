package terminal

import (
	"encoding/base64"
	"fmt"
	"mime"
	"strings"
)

type itermImage struct {
	alt          string
	content_type string
	content      string
	height       string
	width        string
}

func (i *itermImage) asHTML() string {
	parts := []string{
		fmt.Sprintf(`alt="%s"`, i.alt),
		fmt.Sprintf(`src="data:%s;base64,%s"`, i.content_type, i.content),
	}
	if i.width != "" {
		parts = append(parts, fmt.Sprintf(`width="%s"`, i.width))
	}
	if i.height != "" {
		parts = append(parts, fmt.Sprintf(`height="%s"`, i.height))
	}
	return fmt.Sprintf(`<img %s>`, strings.Join(parts, " "))
}

func parseItermImageSequence(sequence string) (*itermImage, error) {
	// Expect 1337;File=name=1.gif;inline=1:BASE64

	arguments, content, err := splitAndVerifyImageSequence(sequence)
	if err != nil {
		return nil, err
	}

	arguments = strings.Map(htmlStripper, arguments)
	imageInline := false

	img := &itermImage{content: content}
	for _, arg := range strings.Split(arguments, ";") {
		argParts := strings.SplitN(arg, "=", 2)
		if len(argParts) != 2 {
			continue
		}
		key := argParts[0]
		val := argParts[1]
		switch strings.ToLower(key) {
		case "name":
			img.alt = val
			img.content_type = contentTypeForFile(val)
		case "inline":
			imageInline = val == "1"
		case "width":
			img.width = parseImageDimension(val)
		case "height":
			img.height = parseImageDimension(val)
		}
	}

	if img.alt == "" {
		return nil, fmt.Errorf("name= argument not supplied, required to determine content type")
	}
	if img.content_type == "" {
		return nil, fmt.Errorf("can't determine content type for %q", img.alt)
	}

	if !imageInline {
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
	prefixLen := len("1337;File=")
	if !strings.HasPrefix(s, "1337;File=") {
		if len(s) > prefixLen {
			s = s[:prefixLen] // Don't blow out our error output
		}
		return "", "", fmt.Errorf("expected sequence to start with 1337;File=, got %q instead", s)
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
