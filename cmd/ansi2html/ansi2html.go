package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/buildbox/terminal"
)

func check(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func main() {
	if len(os.Args) > 1 {
		if os.Args[1] == "-serve" {
			http.HandleFunc("/terminal", func(w http.ResponseWriter, r *http.Request) {
				input, err := ioutil.ReadAll(r.Body)

				check(err)
				output := terminal.Render(input)
				w.Write([]byte(output))
			})

			log.Printf("Listening on port 1337")
			log.Fatal(http.ListenAndServe(":1337", nil))
		} else {
			input, err := ioutil.ReadFile(os.Args[1])
			check(err)
			fmt.Printf("%s", terminal.Render(input))
		}
	} else {
		input, err := ioutil.ReadAll(os.Stdin)
		check(err)
		fmt.Printf("%s", terminal.Render(input))
	}
}
