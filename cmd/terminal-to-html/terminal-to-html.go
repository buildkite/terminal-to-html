package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/buildkite/terminal-to-html/v3"
	"github.com/buildkite/terminal-to-html/v3/internal/assets"
	"github.com/urfave/cli/v2"
)

const appHelpTemplate = `{{.Name}} - {{.Usage}}

STDIN/STDOUT USAGE:
  cat input.raw | {{.Name}} [arguments...] > out.html

WEBSERVICE USAGE:
  {{.Name}} --http :6060 &
  curl --data-binary "@input.raw" http://localhost:6060/terminal > out.html

OPTIONS:
  {{range .Flags}}{{.}}
  {{end}}
`

const (
	// Preview = prologue + stylesheet + interlogue + content + epilogue

	previewPrologue = `
<!DOCTYPE html>
<html>
	<head>
		<meta charset="UTF-8">
		<title>terminal-to-html Preview</title>
		<style>`

	previewInterlogue = `</style>
	</head>
	<body>
		<div class="term-container">`

	previewEpilogue = `</div>
	</body>
</html>
`
)

func writePreviewStart(w io.Writer) error {
	styleSheet, err := assets.TerminalCSS()
	if err != nil {
		return err
	}
	if _, err := w.Write([]byte(previewPrologue)); err != nil {
		return err
	}
	if _, err := w.Write(styleSheet); err != nil {
		return err
	}
	if _, err := w.Write([]byte(previewInterlogue)); err != nil {
		return err
	}
	return nil
}

func writePreviewEnd(w io.Writer) error {
	_, err := w.Write([]byte(previewEpilogue))
	return err
}

func webservice(listen string, preview bool, maxLines int) {
	http.HandleFunc("/terminal", func(w http.ResponseWriter, r *http.Request) {
		// Process the request body, but write to a buffer before serving it.
		// Consuming the body before any writes is necessary because of HTTP
		// limitations (see http.ResponseWriter):
		// > Depending on the HTTP protocol version and the client, calling
		// > Write or WriteHeader may prevent future reads on the
		// > Request.Body.
		// However, it lets us provide Content-Length in all cases.
		b := bytes.NewBuffer(nil)
		if err := process(b, r.Body, preview, maxLines); err != nil {
			log.Printf("error starting preview: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Error creating preview.")
		}

		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Content-Length", strconv.Itoa(b.Len()))
		if _, err := w.Write(b.Bytes()); err != nil {
			log.Printf("error writing response: %v", err)
		}
	})

	log.Printf("Listening on %s", listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}

// process streams the src through a terminal renderer to the dst. If preview is
// true, the preview wrapper is added.
func process(dst io.Writer, src io.Reader, preview bool, maxLines int) error {
	if preview {
		if err := writePreviewStart(dst); err != nil {
			return fmt.Errorf("write start of preview: %w", err)
		}
	}

	s := &terminal.Screen{
		MaxLines:      maxLines,
		ScrollOutFunc: func(line string) { fmt.Fprintln(dst, line) },
	}
	if _, err := io.Copy(s, src); err != nil {
		return fmt.Errorf("read input into screen buffer: %w", err)
	}

	// Write what remains in the screen buffer (everything that didn't scroll
	// out of the top).
	fmt.Fprintln(dst, s.AsHTML())

	if preview {
		return writePreviewEnd(dst)
	}
	return nil
}

func main() {
	cli.AppHelpTemplate = appHelpTemplate

	app := cli.NewApp()

	app.Name = "terminal-to-html"
	app.Version = terminal.Version()
	app.Usage = "turn ANSI in to HTML"
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:  "http",
			Value: "",
			Usage: "HTTP service mode (eg --http :6060), endpoint is /terminal",
		},
		&cli.BoolFlag{
			Name:  "preview",
			Usage: "wrap output in HTML & CSS so it can be easily viewed directly in a browser",
		},
		&cli.IntFlag{
			Name:  "buffer-max-lines",
			Usage: "Sets a limit on the number of lines to hold in the screen buffer, allowing the renderer to operate in a streaming fashion and enabling the processing of large inputs",
		},
	}
	app.Action = func(c *cli.Context) error {
		// Run a web server?
		if addr := c.String("http"); addr != "" {
			webservice(addr, c.Bool("preview"), c.Int("buffer-max-lines"))
			return nil
		}

		// Read input from either stdin or a file.
		input := os.Stdin
		if args := c.Args(); args.Len() > 0 {
			fpath := args.Get(0)
			f, err := os.Open(args.Get(0))
			if err != nil {
				return fmt.Errorf("read %s: %w", fpath, err)
			}
			input = f
		}
		return process(os.Stdout, input, c.Bool("preview"), c.Int("buffer-max-lines"))
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Couldn't %v", err)
	}
}
