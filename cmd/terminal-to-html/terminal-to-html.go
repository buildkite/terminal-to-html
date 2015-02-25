package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/buildkite/terminal"
)

func check(m string, e error) {
	if e != nil {
		log.Fatalf("%s: %v", m, e)
	}
}

var serve = flag.String("http", "", "HTTP service address (e.g., ':6060')")

func main() {
	flag.ErrHelp = errors.New("flag: help requested")

	flag.Parse()

	if *serve != "" {
		http.HandleFunc("/terminal", func(w http.ResponseWriter, r *http.Request) {
			input, err := ioutil.ReadAll(r.Body)

			check("could not read from HTTP stream", err)
			output := terminal.Render(input)
			w.Write(output)
		})

		log.Printf("Listening on %s", *serve)
		log.Fatal(http.ListenAndServe(*serve, nil))
	} else {
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
}
