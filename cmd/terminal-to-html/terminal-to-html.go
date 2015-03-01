package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/buildkite/terminal"
	"github.com/codegangsta/cli"
)

func check(m string, e error) {
	if e != nil {
		log.Fatalf("%s: %v", m, e)
	}
}

func webservice(listen string) {
	http.HandleFunc("/terminal", func(w http.ResponseWriter, r *http.Request) {
		input, err := ioutil.ReadAll(r.Body)
		check("could not read from HTTP stream", err)
		w.Write(terminal.Render(input))
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
	fmt.Printf("%s", terminal.Render(input))
}

func main() {
	app := cli.NewApp()
	app.Name = "terminal-to-html"
	app.Usage = "input ANSI on STDIN, output HTML to STDOUT"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "http",
			Value: "",
			Usage: "HTTP service mode (eg --http :6060), endpoint is /terminal",
		},
	}
	app.Action = func(c *cli.Context) {
		if c.String("http") != "" {
			webservice(c.String("http"))
		} else {
			stdin()
		}
	}
	app.Run(os.Args)

}
