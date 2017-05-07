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
	"golang.org/x/net/websocket"
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
var Debug = false
var RateLimit = 0
var Interval = 0

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

var streamer = new(terminal.Streamer)
var readDone = make(chan bool)

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

func terminalWebsocket(ws *websocket.Conn) {
}

func webservice(listen string) {
	http.HandleFunc("/terminal", func(w http.ResponseWriter, r *http.Request) {
		input, err := ioutil.ReadAll(r.Body)
		check("could not read from HTTP stream", err)
		w.Write(wrapPreview(terminal.Render(input)))
	})

	http.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, r.URL.Path[1:])
	})

	http.Handle("/ws", websocket.Handler(terminalWebsocket))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "assets/index.html")
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

func streamDirty() {
	output, err := streamer.Dirty()
	check("Error streaming output", err)
	for _, line := range output {
		if Debug {
			fmt.Printf("%s\n", line)
		}
	}
}

func stream() {
	reader := os.Stdin
	if len(flag.Arg(0)) > 0 {
		file, err := os.Open(flag.Arg(0))
		check(fmt.Sprintf("could not read %s", flag.Arg(0)), err)
		reader = file
	}

	poller := time.NewTicker(time.Millisecond * time.Duration(Interval))

	go func() {
		for _ = range poller.C {
			streamDirty()
		}
	}()

	buf := make([]byte, 100)
	bytesRead := 0
	for {
		n, err := reader.Read(buf)
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
	streamDirty()
	readDone <- true
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
			Name:  "debug",
			Usage: "Print updates from the streamer to stdout",
		},
		cli.IntFlag{
			Name:  "interval",
			Usage: "Send updates to clients every N milliseconds (default 100)",
		},
		cli.IntFlag{
			Name:  "rateLimit",
			Usage: "Rate limit the STDIN / file reader to N bytes per second",
		},
	}
	app.Action = func(c *cli.Context) {
		PreviewMode = c.Bool("preview")
		Debug = c.Bool("debug")
		RateLimit = c.Int("rateLimit")
		Interval = c.Int("interval")
		if Interval == 0 {
			Interval = 100
		}

		go stream()
		if c.String("http") != "" {
			webservice(c.String("http"))
		}
		<-readDone
	}
	app.Run(os.Args)
}
