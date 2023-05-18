package assets

import (
	"bytes"
	"embed"
	"fmt"
	"io"
)

//go:embed terminal.css
var fs embed.FS

func TerminalCSS() ([]byte, error) {
	f, err := fs.Open("terminal.css")
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	buf := bytes.NewBuffer([]byte{})
	if _, err = io.Copy(buf, f); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return buf.Bytes(), nil
}
