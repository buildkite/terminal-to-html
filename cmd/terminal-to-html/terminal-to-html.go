package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/buildkite/terminal-to-html/v3"
	"github.com/buildkite/terminal-to-html/v3/internal/assets"
	"github.com/buildkite/terminal-to-html/v3/internal/rusage"
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

func wrapPreview(s []byte) ([]byte, error) {
	if PreviewMode {
		s = bytes.Replace([]byte(PreviewTemplate), []byte("CONTENT"), s, 1)
		styleSheet, err := assets.TerminalCSS()
		if err != nil {
			return nil, err
		}
		s = bytes.Replace(s, []byte("STYLESHEET"), styleSheet, 1)
	}
	return s, nil
}

func webservice(listen string) {
	http.HandleFunc("/terminal", func(w http.ResponseWriter, r *http.Request) {
		input, err := io.ReadAll(r.Body)
		if err != nil {
			log.Printf("could not read from HTTP stream: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Error reading request.")
			return
		}

		respBody, err := wrapPreview(terminal.Render(input))
		if err != nil {
			log.Printf("error wrapping preview: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, "Error creating preview.")
			return
		}

		_, err = w.Write(respBody)
		if err != nil {
			log.Printf("error writing response: %v", err)
		}
	})

	log.Printf("Listening on %s", listen)
	log.Fatal(http.ListenAndServe(listen, nil))
}

func stdin() (in, out int) {
	var input []byte
	var err error
	if len(flag.Arg(0)) > 0 {
		input, err = os.ReadFile(flag.Arg(0))
		check(fmt.Sprintf("could not read %s", flag.Arg(0)), err)
	} else {
		input, err = io.ReadAll(os.Stdin)
		check("could not read stdin", err)
	}
	output, err := wrapPreview(terminal.Render(input))
	check("could not wrap preview", err)
	fmt.Printf("%s", output)
	return len(input), len(output)
}

func logResourceStats(start time.Time, in, out int) {
	var fullStats struct {
		// Wall-clock time
		Rtime time.Duration

		// OS-reported statistics
		*rusage.Resources

		// Total input and output bytes processed
		InputBytes, OutputBytes int

		// Other useful memory statistics (see runtime.MemStats)
		TotalAlloc    uint64
		HeapAlloc     uint64
		HeapInuse     uint64
		Mallocs       uint64
		Frees         uint64
		PauseTotalNs  uint64
		NumGC         uint32
		GCCPUFraction float64
	}
	fullStats.Rtime = time.Since(start)
	fullStats.InputBytes = in
	fullStats.OutputBytes = out

	ru, err := rusage.Stats()
	if err != nil {
		log.Printf("Could not read OS resource usage: %v", err)
	}
	fullStats.Resources = ru

	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)
	fullStats.TotalAlloc = memStats.TotalAlloc
	fullStats.HeapAlloc = memStats.HeapAlloc
	fullStats.HeapInuse = memStats.HeapInuse
	fullStats.Mallocs = memStats.Mallocs
	fullStats.Frees = memStats.Frees
	fullStats.PauseTotalNs = memStats.PauseTotalNs
	fullStats.NumGC = memStats.NumGC
	fullStats.GCCPUFraction = memStats.GCCPUFraction

	if err := json.NewEncoder(os.Stderr).Encode(&fullStats); err != nil {
		log.Fatalf("Could not encode resource usage: %v", err)
	}
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
		&cli.BoolFlag{
			Name:  "log-resource-stats",
			Usage: "Log resource statistics to stderr after successfully processing",
		},
	}
	app.Action = func(c *cli.Context) error {
		PreviewMode = c.Bool("preview")
		if c.String("http") != "" {
			webservice(c.String("http"))
		} else {
			start := time.Now()
			in, out := stdin()

			if c.Bool("log-resource-stats") {
				logResourceStats(start, in, out)
			}
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
