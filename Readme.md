![logo](http://buildboxhq.github.io/terminal/images/logo.png)

[![Gem Version](https://badge.fury.io/rb/terminal.png)](https://rubygems.org/gems/terminal)

Terminal takes any arbitrary crazy shell output (ASCII), and turns it into beautifully rendered HTML.

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

### With the Buildbox API

First install [jq](http://stedolan.github.io/jq/), if you have [Homebrew](http://brew.sh/) installed, you can just `brew install jq`.

Then, you can:

```bash
$JOB_LOG_URL="https://api.buildbox.io/v1/accounts/[account]/projects/[project]/builds/[build]/jobs/[job]/log?api_key=[api-key]"
echo $(curl $JOB_LOG_URL -s | jq '.content') | terminal
```

For more information on the Buildbox Builds API, see: https://buildbox.io/docs/api/builds

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
