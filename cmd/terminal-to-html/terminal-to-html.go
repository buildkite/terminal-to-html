package main

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"sync"
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

var wsClientMutex = new(sync.Mutex)
var wsClients = make([]wsClient, 0)

type wsClient struct {
	ws     *websocket.Conn
	writer chan terminal.DirtyLine
}

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
	data := streamer.Flush(true)

	for _, line := range data {
		err := websocket.JSON.Send(ws, line)
		if err != nil {
			fmt.Printf("Warning: Could not send initial status to websocket: %s", err)
			return
		}
	}

	wsClient := wsClient{ws, make(chan terminal.DirtyLine, 100)}
	wsClientMutex.Lock()
	wsClients = append(wsClients, wsClient)
	wsClientMutex.Unlock()

	defer func() {
		wsClientMutex.Lock()
		for i, c := range wsClients {
			if c == wsClient {
				wsClients = append(wsClients[:i], wsClients[i+1:]...)
				break
			}
		}
		wsClientMutex.Unlock()
	}()

	for line := range wsClient.writer {
		err := websocket.JSON.Send(ws, &line)
		if err != nil {
			fmt.Printf("Warning: Could not write to websocket %s", err)
			return
		}
	}
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
	output := streamer.Flush(false)

	wsClientMutex.Lock()

	if Debug {
		for _, line := range output {
			fmt.Printf("%d: %s\n", line.Y, line.HTML)
		}
	}

	for _, client := range wsClients {
		for _, line := range output {
			client.writer <- line
		}
	}
	wsClientMutex.Unlock()

}

func stream(filename string) {
	reader := os.Stdin
	if len(filename) > 0 {
		file, err := os.Open(filename)
		check(fmt.Sprintf("could not read %s", filename), err)
		reader = file
	}

	poller := time.NewTicker(time.Millisecond * time.Duration(Interval))

	go func() {
		for _ = range poller.C {
			streamDirty()
		}
	}()

	buf := make([]byte, 100)
	bytesRead := 0.0
	for {
		n, err := reader.Read(buf)
		if err == io.EOF {
			break
		}
		bytesRead += float64(n)
		if bytesRead > float64(RateLimit)/10.0 && RateLimit > 0 {
			time.Sleep(time.Millisecond * 100)
			bytesRead = 0.0
		}
		check("could not read stdin/reader", err)
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
			Value: ":6060",
			Usage: "HTTP port number (eg --http :6060)",
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
			Value: 100,
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

		go stream(c.Args().First())

		webservice(c.String("http"))
		<-readDone

		fmt.Println(string(streamer.Render()))

	}
	app.Run(os.Args)
}
