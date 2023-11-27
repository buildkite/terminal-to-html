package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/buildkite/terminal-to-html/v3"
)

func check(m string, e error) {
	if e != nil {
		log.Fatalf("%s: %v", m, e)
	}
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
	output := terminal.Render(input)
	fmt.Printf("%s\n", output)
}

func main() {
	stdin()
}
