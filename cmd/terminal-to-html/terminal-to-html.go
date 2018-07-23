package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/buildkite/terminal"
	"github.com/codegangsta/cli"
	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/html"
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
var MinifyMode = false

var PreviewTemplate = `
	<!DOCTYPE html>
	<html>
		<head>
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

func MinifyHtml(s string) string {
	m := minify.New()
	m.AddFunc("text/html", html.Minify)
	m.Add("text/html", &html.Minifier{
		KeepDefaultAttrVals: true,
		KeepWhitespace:      false,
	})
	var err error
	s, err = m.String("text/html", s)
	if err != nil {
		panic(err)
	}
	return s
}

func MinifyCss(b []byte) []byte {
	m := minify.New()
	m.AddFunc("text/css", css.Minify)
	m.Add("text/css", &css.Minifier{})
	var err error
	b, err = m.Bytes("text/css", b)
	if err != nil {
		panic(err)
	}
	return b
}

func wrapPreview(s []byte) []byte {
	if PreviewMode {
		if MinifyMode {
			s = bytes.Replace([]byte(MinifyHtml(PreviewTemplate)), []byte("CONTENT"), s, 1)
			s = bytes.Replace(s, []byte("STYLESHEET"), MinifyCss(MustAsset("assets/terminal.css")), 1)
		} else {
			s = bytes.Replace([]byte(PreviewTemplate), []byte("CONTENT"), s, 1)
			s = bytes.Replace(s, []byte("STYLESHEET"), MustAsset("assets/terminal.css"), 1)
		}
	}
	return s
}

func webservice(listen string) {
	http.HandleFunc("/terminal", func(w http.ResponseWriter, r *http.Request) {
		input, err := ioutil.ReadAll(r.Body)
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
		input, err = ioutil.ReadFile(flag.Arg(0))
		check(fmt.Sprintf("could not read %s", flag.Arg(0)), err)
	} else {
		input, err = ioutil.ReadAll(os.Stdin)
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
		cli.StringFlag{
			Name:  "http",
			Value: "",
			Usage: "HTTP service mode (eg --http :6060), endpoint is /terminal",
		},
		cli.BoolFlag{
			Name:  "preview",
			Usage: "wrap output in HTML & CSS so it can be easily viewed directly in a browser",
		},
		cli.BoolFlag{
			Name:  "minify",
			Usage: "minify the html header and css",
		},
	}
	app.Action = func(c *cli.Context) {
		PreviewMode = c.Bool("preview")
		MinifyMode = c.Bool("minify")
		if c.String("http") != "" {
			webservice(c.String("http"))
		} else {
			stdin()
		}
	}
	app.Run(os.Args)
}
