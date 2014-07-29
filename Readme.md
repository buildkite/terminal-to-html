![logo](http://buildbox.github.io/terminal/images/logo.svg)

[![Gem Version](https://badge.fury.io/rb/terminal.png)](https://rubygems.org/gems/terminal)

Terminal is a Ruby library for converting arbitrary shell output (with ANSI) into beautifully rendered HTML. See http://www.termsys.demon.co.uk/vtansi.htm for more information about ANSI Terminal Control Escape Sequences.

## Installation

Add this line to your application's Gemfile:

```ruby
gem 'terminal'
```

And then execute:

```bash
$ bundle
```

Or install it yourself as:

```bash
gem install terminal
```

## Usage

```ruby
Terminal.render("...")
```

### Rails Integration

You can use Terminal directly within your Ruby on Rails application. First require the gem
in your Gemfile:

```ruby
gem "terminal"
```

Then in your `app/assets/application.css` file, add `terminal.css`

```css
/* require "terminal" */
```

Now in your views:

```html
<div class="code"><%= Terminal.render(output)</div>
```

### Command Line

Terminal ships with a command line utility. For example, you can pipe `rspec` output to it:

```bash
rspec --color --tty | terminal
```

Or use output saved earlier:

```bash
rspec --tty --color > output.txt
terminal output.txt
```

With `rspec`, you'll need to use the `--tty` and `--color` options to force it to output colors.

We also provide a utility to preview the rendered version in a web browser. Simply append `--preview` to the command,
and when the render has finished, it will open in your web browser with a before/after show.

```bash
rspec --color --tty | terminal --preview
```

![logo](http://buildbox.github.io/terminal/images/preview.png)

### With the Buildbox API

First install [jq](http://stedolan.github.io/jq/), if you have [Homebrew](http://brew.sh/) installed, you can just `brew install jq`.

Then, you can:

```bash
export JOB_LOG_URL="https://api.buildbox.io/v1/accounts/[account]/projects/[project]/builds/[build]/jobs/[job]/log?api_key=[api-key]"
curl $JOB_LOG_URL -s | jq '.content' -r | terminal
```

For more information on the Buildbox Builds API, see: https://buildbox.io/docs/api/builds

## Generating Fixtures

To generate a fixture, first create a test case inside the `examples` folder. See the `curl.sh`
file as an example. You can then generate a `.raw` and `.rendered` file by running:

```bash
./generate curl.sh
```

You should then move the `raw` and `rendered` files to the `fixtures` folder.

```bash
mv examples/*{raw,rendered} spec/fixtures
```

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
