// termplayer outputs the contents of a file "slowly". It "plays back" raw
// Buildkite job logs as though the job was running in a local terminal.
package main

import (
	"bufio"
	"flag"
	"io"
	"log"
	"os"
	"regexp"
	"strconv"
	"time"
)

var (
	buildkiteMode = flag.Bool("bk", true, "If the file contains BK metadata, emit output at times corresponding to embedded timestamps instead of at a fixed rate")
	speed         = flag.Int("speed", 1, "Rate of lines emitted per second. In BK mode, this multiplies the output speed")
)

var buildkiteRE = regexp.MustCompile(`^_bk;t=(\d+)$`)

func main() {
	flag.Parse()

	input := os.Stdin
	if len(flag.Args()) > 0 && flag.Arg(0) != "-" {
		f, err := os.Open(flag.Arg(0))
		if err != nil {
			log.Fatalf("Couldn't open file: %v", err)
		}
		defer f.Close()
		input = f
	}

	rd := bufio.NewReader(input)
	if *buildkiteMode {
		buildkiteModeOutput(rd)
	} else {
		fixedRateOutput(rd)
	}
}

func buildkiteModeOutput(rd *bufio.Reader) {
	var lastTS int
	for {
		chunk, err := rd.ReadBytes(0x1b)
		os.Stdout.Write(chunk)
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatalf("Reading byte from input: %v\n", err)
		}
		code, err := rd.Peek(20)
		if err == bufio.ErrBufferFull || err == io.EOF {
			continue
		}
		if err != nil {
			log.Fatalf("Peeking 20 bytes from input: %v\n", err)
		}
		matches := buildkiteRE.FindSubmatch(code)
		if matches == nil {
			continue
		}
		ts, err := strconv.Atoi(string(matches[1]))
		if err != nil {
			log.Fatalf("Converting string to int: %v", err)
		}
		if lastTS == 0 {
			lastTS = ts
			continue
		}
		dt := time.Duration(ts-lastTS) * time.Millisecond
		lastTS = ts
		if dt > 0 {
			time.Sleep(dt / time.Duration(*speed))
		}
	}
}

func fixedRateOutput(rd *bufio.Reader) {
	for range time.Tick(time.Second / time.Duration(*speed)) {
		line, err := rd.ReadBytes('\n')
		os.Stdout.Write(line)
		if err == io.EOF {
			return
		}
		if err != nil {
			log.Fatalf("Reading bytes from input: %v\n", err)
		}
	}
}
