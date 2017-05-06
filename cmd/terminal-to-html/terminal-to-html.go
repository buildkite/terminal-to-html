package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"io"

	"github.com/buildkite/terminal"
	"github.com/codegangsta/cli"
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
var DebugPoller = 0
var RateLimit = 0

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

func wrapPreview(s []byte) []byte {
	if PreviewMode {
		s = bytes.Replace([]byte(PreviewTemplate), []byte("CONTENT"), s, 1)
		s = bytes.Replace(s, []byte("STYLESHEET"), MustAsset("assets/terminal.css"), 1)
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

func streamStdin() {
	streamer := new(terminal.Streamer)

	if len(flag.Arg(0)) > 0 {
		log.Fatalf("Can't specify debugPoller and an input file, use stdin instead")
	}
	poller := time.NewTicker(time.Millisecond * time.Duration(DebugPoller))

	go func() {
		for _ = range poller.C {
			output, err := streamer.Dirty()
			check("Error streaming output", err)
			for _, line := range output {
				fmt.Printf("%s\n", line)
			}
		}
	}()

	buf := make([]byte, 100)
	bytesRead := 0
	for {
		n, err := os.Stdin.Read(buf)
		if err == io.EOF {
			break
		}
		bytesRead += n
		if bytesRead > RateLimit/10 && RateLimit > 0 {
			time.Sleep(time.Millisecond * 100)
		}
		check("could not read stdin", err)
		streamer.Write(buf[0:n])
	}
	poller.Stop()
	output, err := streamer.Dirty()
	check("Error streaming output", err)
	for _, line := range output {
		fmt.Printf("%s\n", line)
	}
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
		cli.IntFlag{
			Name:  "debugPoller",
			Usage: "Print streaming updates every N milliseconds",
		},
		cli.IntFlag{
			Name:  "rateLimit",
			Usage: "Rate limit the STDIN / file reader to N bytes per second",
		},
	}
	app.Action = func(c *cli.Context) {
		PreviewMode = c.Bool("preview")
		DebugPoller = c.Int("debugPoller")
		RateLimit = c.Int("rateLimit")

		if c.String("http") != "" {
			webservice(c.String("http"))
		} else if DebugPoller > 0 {
			streamStdin()
		} else {
			stdin()
		}
	}
	app.Run(os.Args)
}
