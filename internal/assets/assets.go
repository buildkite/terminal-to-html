package assets

import (
	"bytes"
	"compress/gzip"
	"embed"
	"fmt"
	"io"
)

//go:generate gzip --keep --force terminal.css

//go:embed terminal.css.gz
var fs embed.FS

func TerminalCSS() ([]byte, error) {
	gzf, err := fs.Open("terminal.css.gz")
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	r, err := gzip.NewReader(gzf)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	defer r.Close()

	buf := bytes.NewBuffer([]byte{})
	if _, err = io.Copy(buf, r); err != nil {
		return nil, fmt.Errorf("%w", err)
	}

	return buf.Bytes(), nil
}
