package main

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/buildkite/terminal-to-html/v3"
	"github.com/buildkite/terminal-to-html/v3/internal/assets"
	"github.com/urfave/cli/v2"
)

var AppHelpTemplate = `{{.Name}} - {{.Usage}}

STDIN/STDOUT USAGE:
  cat input.raw | {{.Name}} [arguments...] > out.html

WEBSERVICE USAGE:
  {{.Name}} --http :6060 &
  curl --data-binary "@input.raw" http://localhost:6060/terminal > out.html

OPTIONS:
  {{range .Flags}}{{.}}
  {{end}}
`

var PreviewMode = false

var PreviewTemplate = `
	<!DOCTYPE html>
	<html>
		<head>
			<meta charset="UTF-8">
			<title>terminal-to-html Preview</title>
			<style>STYLESHEET</style>
		</head>
		<body>
			<div class="term-container">CONTENT</div>
		</body>
	</html>
`

func check(m string, e error) {
	if e != nil {
		log.Fatalf("%s: %v", m, e)
	}
}

func wrapPreview(s []byte) []byte {
	if PreviewMode {
		s = bytes.Replace([]byte(PreviewTemplate), []byte("CONTENT"), s, 1)
		styleSheet, err := assets.TerminalCSS()
		check("could not retrive stylesheet", err)
		s = bytes.Replace(s, []byte("STYLESHEET"), styleSheet, 1)
	}
	return s
}

func webservice(listen string) {
	http.HandleFunc("/terminal", func(w http.ResponseWriter, r *http.Request) {
		input, err := io.ReadAll(r.Body)
		check("could not read from HTTP stream", err)
		w.Write(wrapPreview(terminal.Render(input)))
	})

	log.Printf("Listening on %s", listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}

func stdin() {
	var input []byte
	var err error
	if len(flag.Arg(0)) > 0 {
		input, err = os.ReadFile(flag.Arg(0))
		check(fmt.Sprintf("could not read %s", flag.Arg(0)), err)
	} else {
		input, err = io.ReadAll(os.Stdin)
		check("could not read stdin", err)
	}
	fmt.Printf("%s", wrapPreview(terminal.Render(input)))
}

func main() {
	cli.AppHelpTemplate = AppHelpTemplate

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
	}
	app.Action = func(c *cli.Context) error {
		PreviewMode = c.Bool("preview")
		if c.String("http") != "" {
			webservice(c.String("http"))
		} else {
			stdin()
		}
		return nil
	}
	app.Run(os.Args)
}
