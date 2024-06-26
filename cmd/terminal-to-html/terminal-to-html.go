package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"time"

	"github.com/buildkite/terminal-to-html/v3"
	"github.com/buildkite/terminal-to-html/v3/internal/assets"
	"github.com/buildkite/terminal-to-html/v3/internal/rusage"
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
		if _, _, _, err := process(b, r.Body, preview, maxLines); err != nil {
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

func logStats(start time.Time, in, out int, s *terminal.Screen) {
	var fullStats struct {
		// Wall-clock time
		Rtime time.Duration

		// OS-reported statistics
		*rusage.Resources

		// Total input and output bytes processed
		InputBytes, OutputBytes int

		// Screen processing statistics (see terminal.Screen)
		LinesScrolledOut int
		CursorUpOOB      int
		CursorBackOOB    int

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

	fullStats.LinesScrolledOut = s.LinesScrolledOut
	fullStats.CursorUpOOB = s.CursorUpOOB
	fullStats.CursorBackOOB = s.CursorBackOOB

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

type writeCounter struct {
	out     io.Writer
	counter int
}

func (wc *writeCounter) Write(b []byte) (int, error) {
	n, err := wc.out.Write(b)
	wc.counter += n
	return n, err
}

// process streams the src through a terminal renderer to the dst. If preview is
// true, the preview wrapper is added.
func process(dst io.Writer, src io.Reader, preview bool, maxLines int) (in, out int, screen *terminal.Screen, err error) {
	// Wrap dst in writeCounter to count bytes written
	wc := &writeCounter{out: dst}

	if preview {
		if err := writePreviewStart(wc); err != nil {
			return 0, wc.counter, nil, fmt.Errorf("write start of preview: %w", err)
		}
	}

	screen = &terminal.Screen{
		MaxLines:      maxLines,
		ScrollOutFunc: func(line string) { fmt.Fprintln(wc, line) },
	}
	inBytes, err := io.Copy(screen, src)
	if err != nil {
		return int(inBytes), wc.counter, screen, fmt.Errorf("read input into screen buffer: %w", err)
	}

	// Write what remains in the screen buffer (everything that didn't scroll
	// out of the top).
	fmt.Fprint(wc, screen.AsHTML())

	if preview {
		if err := writePreviewEnd(wc); err != nil {
			return int(inBytes), wc.counter, screen, fmt.Errorf("write end of preview: %w", err)
		}
	}
	return int(inBytes), wc.counter, screen, nil
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
		&cli.BoolFlag{
			Name:  "log-stats-to-stderr",
			Usage: "Logs a JSON object to stderr containing resource and processing statistics after successfully processing",
		},
		&cli.IntFlag{
			Name:  "buffer-max-lines",
			Value: 300,
			Usage: "Sets a limit on the number of lines to hold in the screen buffer, allowing the renderer to operate in a streaming fashion and enabling the processing of large inputs. Setting to 0 disables the limit, causing the renderer to buffer the entire screen before producing any output",
		},
	}
	app.Action = func(c *cli.Context) error {
		// Run a web server?
		if addr := c.String("http"); addr != "" {
			webservice(addr, c.Bool("preview"), c.Int("buffer-max-lines"))
			return nil
		}

		start := time.Now()

		// Read input from either stdin or a file.
		input := os.Stdin
		if args := c.Args(); args.Len() > 0 {
			fpath := args.Get(0)
			f, err := os.Open(fpath)
			if err != nil {
				return fmt.Errorf("read %s: %w", fpath, err)
			}
			input = f
		}

		in, out, screen, err := process(os.Stdout, input, c.Bool("preview"), c.Int("buffer-max-lines"))
		if err != nil {
			return err
		}
		if c.Bool("log-stats-to-stderr") {
			logStats(start, in, out, screen)
		}
		return nil
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Couldn't %v", err)
	}
}
