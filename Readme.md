![logo](http://buildkite.github.io/terminal/images/logo.svg)

Terminal is a Go library for converting arbitrary shell output (with ANSI) into beautifully rendered HTML. See http://en.wikipedia.org/wiki/ANSI_escape_code for more information about ANSI Terminal Control Escape Sequences.

It provides a single command, `terminal-to-html`, that can be used to convert terminal output via STDIN, as well as via a simple web server.

## Usage

Piping in terminal output via the command line:

``` bash
cat fixtures/pickachu.sh.raw | terminal-to-html > out.html
```

Posting terminal content via HTTP:

```bash
terminal-to-html -http=:6060 &
curl --data-binary "@fixtures/pikachu.sh.raw" http://localhost:6060/terminal > out.html
```

For coloring you can use the sample [terminal.css](/assets/terminal.css) stylesheet and wrap the output in an element with class `term-container` (e.g. `<div class="term-container"><!-- terminal output --></div>`).

## Installation

TODO

## Manual Installation

Assuming a `$GOPATH/bin` that's globally accessible, run:

```bash
go install github.com/buildkite/terminal/cmd/terminal-to-html
```

## Developing

TODO

## Benchmarking

Run `go test -bench .` to see raw Go performance. The `npm` test is the focus: this best represents the kind of use cases the original code was developed against.

## Contributing

1. Fork it ( https://github.com/[my-github-username]/terminal/fork )
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request

## Licence

> Copyright (c) 2015 Keith Pitt, Tim Lucas, Michael Pearson
>
> MIT License
>
> Permission is hereby granted, free of charge, to any person obtaining
> a copy of this software and associated documentation files (the
> "Software"), to deal in the Software without restriction, including
> without limitation the rights to use, copy, modify, merge, publish,
> distribute, sublicense, and/or sell copies of the Software, and to
> permit persons to whom the Software is furnished to do so, subject to
> the following conditions:
>
> The above copyright notice and this permission notice shall be
> included in all copies or substantial portions of the Software.
>
> THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
> EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
> MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
> NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
> LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
> OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
> WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
