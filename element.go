package terminal

import (
	"encoding/base64"
	"errors"
	"fmt"
	"html"
	"mime"
	"strings"
)

const (
	elementITermImage = iota
	elementITermLink
	elementImage
	elementLink
)

type element struct {
	url         string
	alt         string
	contentType string
	content     string
	height      string
	width       string
	elementType int
}

var errUnsupportedElementSequence = errors.New("Unsupported element sequence")

func (i *element) asHTML() string {
	h := html.EscapeString

	if i.elementType == elementLink {
		content := i.content
		if content == "" {
			content = i.url
		}
		return fmt.Sprintf(`<a href="%s">%s</a>`, h(sanitizeURL(i.url)), h(content))
	}

	alt := i.alt
	if alt == "" {
		alt = i.url
	}

	parts := []string{fmt.Sprintf(`alt="%s"`, h(alt))}

	switch i.elementType {
	case elementITermImage:
		src := fmt.Sprintf(`src="data:%s;base64,%s"`, h(i.contentType), h(i.content))
		parts = append(parts, src)

	case elementImage:
		url := sanitizeURL(i.url)
		if url == "" || url == unsafeURLSubstitution {
			// don't emit an <img> at all if the URL is empty or didn't sanitize
			return ""
		}
		src := fmt.Sprintf(`src="%s"`, h(url))
		parts = append(parts, src)

	default:
		// unreachable, but…
		return ""
	}

	if i.width != "" {
		parts = append(parts, fmt.Sprintf(`width="%s"`, h(i.width)))
	}
	if i.height != "" {
		parts = append(parts, fmt.Sprintf(`height="%s"`, h(i.height)))
	}

	return fmt.Sprintf(`<img %s>`, strings.Join(parts, " "))
}

func parseElementSequence(sequence string) (*element, error) {
	// Expect:
	// - iTerm style hyperlink:    8;id=1234;http://example.com/
	// - iTerm style inline image: 1337;File=name=1.gif;inline=1:BASE64
	// - Buildkite external image: 1338;url=…;alt=…;width=…;height=…
	// - Buildkite hyperlink:      1339;url=…;content=…

	args, elementType, content, err := splitAndVerifyElementSequence(sequence)
	if err != nil {
		if err == errUnsupportedElementSequence {
			err = nil
		}
		return nil, err
	}

	elem := &element{content: content, elementType: elementType}

	if elementType == elementITermLink {
		// For "iTerm" links (OSC 8), tokens[0] is params and tokens[1] is the URL.
		// Aside from not quoting the URL, we ignore params.
		// The link "content" comes after the element and is stored
		// as regular text in the screen line, because they are designed to gracefully
		// degrade to plain text if the sequence isn't supported.
		tokens := strings.Split(args, ";")
		if len(tokens) != 2 {
			// Probably malformed
			return nil, nil
		}
		elem.url = tokens[1]
		return elem, nil
	}

	tokens, err := tokenizeString(args, ';', '\\')
	if err != nil {
		return nil, err
	}

	imageInline := false

	for _, token := range tokens {
		parts := strings.SplitN(token, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := parts[0]
		val := parts[1]
		switch strings.ToLower(key) {
		case "name":
			nameBytes, err := base64.StdEncoding.DecodeString(val)
			if err != nil {
				return nil, fmt.Errorf("name= value of %q is not valid base64", val)
			}
			elem.url = string(nameBytes)
			elem.contentType = contentTypeForFile(elem.url)
		case "url":
			elem.url = val
		case "content":
			elem.content = val
		case "inline":
			imageInline = val == "1"
		case "width":
			elem.width = parseImageDimension(val)
		case "height":
			elem.height = parseImageDimension(val)
		case "alt":
			elem.alt = val
		}
	}

	if elem.elementType == elementITermImage {
		if elem.url == "" {
			return nil, fmt.Errorf("name= argument not supplied, required to determine content type")
		}
		if elem.contentType == "" {
			return nil, fmt.Errorf("can't determine content type for %q", elem.url)
		}
	} else {
		if elem.url == "" {
			return nil, fmt.Errorf("url= argument not supplied")
		}
	}

	if elem.elementType == elementITermImage && !imageInline {
		// in iTerm2, if you don't specify inline=1, the image is merely downloaded
		// and not displayed.
		elem = nil
	}
	return elem, nil
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

func splitAndVerifyElementSequence(s string) (arguments string, elementType int, content string, err error) {
	if rem, has := strings.CutPrefix(s, "8;"); has {
		return rem, elementITermLink, "", nil
	}
	if rem, has := strings.CutPrefix(s, "1338;"); has {
		return rem, elementImage, "", nil
	}
	if rem, has := strings.CutPrefix(s, "1339;"); has {
		return rem, elementLink, "", nil
	}

	rem, has := strings.CutPrefix(s, "1337;File=")
	if !has {
		return "", 0, "", errUnsupportedElementSequence
	}
	s = rem

	parts := strings.Split(s, ":")
	if len(parts) != 2 {
		return "", 0, "", fmt.Errorf("expected sequence to have one arguments part and one content part, got %d part(s)", len(parts))
	}

	elementType = elementITermImage
	arguments = parts[0]
	content = parts[1]
	if len(content) == 0 {
		return "", 0, "", fmt.Errorf("image content missing")
	}

	_, err = base64.StdEncoding.DecodeString(content)
	if err != nil {
		return "", 0, "", fmt.Errorf("expected content part to be valid Base64")
	}

	return
}

func tokenizeString(input string, sep, escape rune) (tokens []string, err error) {
	var runes []rune
	inEscape := false
	inSingleQuotes := false
	inDoubleQuotes := false
	for _, rune := range input {
		switch {
		case inEscape:
			inEscape = false
			fallthrough
		default:
			runes = append(runes, rune)
		case rune == '\'' && !inDoubleQuotes:
			inSingleQuotes = !inSingleQuotes
		case rune == '"' && !inSingleQuotes:
			inDoubleQuotes = !inDoubleQuotes
		case rune == escape:
			inEscape = true
		case rune == sep && !inSingleQuotes && !inDoubleQuotes:
			// end of token: append to tokens and start a new token
			tokens = append(tokens, string(runes))
			runes = runes[:0]
		}
	}
	// end of all tokens; append final token to tokens
	tokens = append(tokens, string(runes))
	if inEscape {
		err = errors.New("invalid terminal escape")
	}
	if inSingleQuotes || inDoubleQuotes {
		err = errors.New("invalid syntax: unclosed quotation marks")
	}
	return tokens, err
}
