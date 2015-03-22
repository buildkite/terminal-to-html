package terminal

import (
	"encoding/base64"
	"fmt"
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
	return fmt.Sprintf(`<img alt=%q src="data:%s;base64,%s">`, i.alt, i.content_type, i.content)
}

func parseItermImageSequence(sequence string) (*itermImage, error) {
	// Expect 1337;File=name=1.gif;inline=1:BASE64

	imageInline := false

	prefixLen := len("1337;File=")
	if !strings.HasPrefix(sequence, "1337;File=") {
		if len(sequence) > prefixLen {
			sequence = sequence[:prefixLen] // Don't blow out our error output
		}
		return nil, fmt.Errorf("Expected sequence to start with 1337;File=, got %q instead", sequence)
	}
	sequence = sequence[prefixLen:]

	parts := strings.Split(sequence, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("Expected sequence to have one arguments part and one content part, got %d parts", len(parts))
	}
	arguments := parts[0]
	content := parts[1]

	_, err := base64.StdEncoding.DecodeString(content)
	if err != nil {
		return nil, fmt.Errorf("Expected content part to be valid Base64")
	}

	img := &itermImage{content: content}
	argsSplit := strings.Split(arguments, ";")
	for _, arg := range argsSplit {
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
		}
	}

	if img.alt == "" {
		return nil, fmt.Errorf("name= argument not supplied, required to determine content type")
	}

	if !imageInline {
		// in iTerm2, if you don't specify inline=1, the image is merely downloaded
		// and not displayed.
		img = nil
	}
	return img, nil
}

func contentTypeForFile(filename string) string {
	return "image/gif"
}
