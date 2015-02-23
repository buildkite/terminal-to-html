![logo](http://buildbox.github.io/terminal/images/logo.svg)

Terminal is a Go library for converting arbitrary shell output (with ANSI) into beautifully rendered HTML. See http://en.wikipedia.org/wiki/ANSI_escape_code for more information about ANSI Terminal Control Escape Sequences.

It provides a single command, `ansi2html`, that can be used either as a simple webservice or via STDIN/STDOUT. It can also be used as a library.

## Installation

Assuming a `$GOPATH/bin` that's globally accessible, run:

```bash
go install github.com/buildbox/terminal/cmd/ansi2html
```

This will give you the `ansi2html` command. It's called `ansi2html` and not `terminal` as installing something called `terminal` globally might confuse people looking for an actual terminal.

## Usage

``` bash
# STDIN/STDOUT Usage
cat fixtures/pickachu.sh.raw | ansi2html > out.html

# Webservice Usage
ansi2html -http=:6060 &
curl --data-binary "@fixtures/pikachu.sh.raw" http://localhost:6060/terminal > out.html
```

You'll need to wrap the resulting output inside a `.term-container` HTML entity and use the stylesheet in `assets/terminal.css`

## Benchmarking

Run `go test -bench .` to see raw Go performance. The `npm` test is the focus: this best represents the kind of use cases the original code was developed against. As a guide, this test was 80ms per iteration on an 2013 Retina MBP, and was 2500 ms per iteration in the original pure Ruby implementation.

## TODO

 * UTF8 enforcement
 * Emoji
 * "Demo" functionality that wraps output in the stylesheet

## Contributing

1. Fork it ( https://github.com/[my-github-username]/terminal/fork )
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push origin my-new-feature`)
5. Create a new Pull Request

## Licence

> Copyright (c) 2014 Keith Pitt, Buildbox
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
