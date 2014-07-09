# Terminal

TODO: Write a gem description

## Installation

Add this line to your application's Gemfile:

    gem 'terminal'

And then execute:

    $ bundle

Or install it yourself as:

    $ gem install terminal

## Usage

```ruby
Terminal.render("...")
```

### Using with the Buildbox API and the command line

First install [jq](http://stedolan.github.io/jq/), if you have [Homebrew](http://brew.sh/) installed, you can just `brew install jq`.

Then, you can:

```bash
echo $(curl "https://api.buildbox.io/v1/accounts/[account]/projects/[project]/builds/[build]/jobs/[job]/log?api_key=[api-key]" -s | jq '.content') | terminal
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
