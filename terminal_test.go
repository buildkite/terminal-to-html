package terminal

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/go-cmp/cmp"
)

var TestFiles = []string{
	"control.sh",
	"curl.sh",
	"cursor-save-restore.sh",
	"docker-compose-pull.sh",
	"docker-pull.sh",
	"homer.sh",
	"itermlinks.sh",
	"npm.sh",
	"pikachu.sh",
	"playwright.sh",
	"pwsh.sh",
	"rustfmt.sh",
	"weather.sh",
}

func loadFixture(t testing.TB, base, ext string) []byte {
	filename := filepath.Join("fixtures", base+"."+ext)
	data, err := os.ReadFile(filename)
	if err != nil {
		t.Fatalf("could not load fixture %s: %v", filename, err)
	}
	return data
}

func base64Encode(stringToEncode string) string {
	return base64.StdEncoding.EncodeToString([]byte(stringToEncode))
}

var rendererTestCases = []struct {
	name  string
	input string
	want  string
}{
	{
		name:  "input that ends in a newline will not include that newline",
		input: "hello\n",
		want:  "hello",
	},
	{
		name:  "closes colors that get opened",
		input: "he\033[32mllo",
		want:  "he<span class=\"term-fg32\">llo</span>",
	},
	{
		name:  "treats multi-byte unicode characters as individual runes",
		input: "€€€€€€\b\b\baaa",
		want:  "€€€aaa",
	},
	{
		name:  "skips over colors when backspacing",
		input: "he\x1b[32m\x1b[33m\bllo",
		want:  "h<span class=\"term-fg33\">llo</span>",
	},
	{
		name:  "handles \x1b[m (no parameter) as a reset",
		input: "\x1b[36mthis has a color\x1b[mthis is normal now\r\n",
		want:  "<span class=\"term-fg36\">this has a color</span>this is normal now",
	},
	{
		name:  "treats \x1b[39m as a reset",
		input: "\x1b[36mthis has a color\x1b[39mthis is normal now\r\n",
		want:  "<span class=\"term-fg36\">this has a color</span>this is normal now",
	},
	{
		name:  "starts overwriting characters when you carriage-return midway through something",
		input: "hello\rb",
		want:  "bello",
	},
	{
		name:  "colors across multiple lines",
		input: "\x1b[32mhello\n\nfriend\x1b[0m",
		want:  "<span class=\"term-fg32\">hello</span>\n&nbsp;\n<span class=\"term-fg32\">friend</span>",
	},
	{
		name:  "allows you to control the cursor forwards",
		input: "this is\x1b[4Cpoop and stuff",
		want:  "this is    poop and stuff",
	},
	{
		name:  "allows you to jump down further than the bottom of the buffer",
		input: "this is great \x1b[1Bhello",
		want:  "this is great\n              hello",
	},
	{
		name:  "allows you to control the cursor backwards",
		input: "this is good\x1b[4Dpoop and stuff",
		want:  "this is poop and stuff",
	},
	{
		name:  "allows large cursor movements backwards",
		input: strings.Repeat("w", 300) + "\x1b[300Dhahaha",
		want:  "hahaha" + strings.Repeat("w", 294),
	},
	{
		name:  "allows you to control the cursor upwards",
		input: "1234\n56\x1b[1A78\x1b[B",
		want:  "1278\n56",
	},
	{
		name: "allows you to control the cursor downwards",
		// creates a grid of:
		// aaaa
		// bbbb
		// cccc
		// Then goes up 2 rows, down 1 row, jumps to the begining
		// of the line, rewrites it to 1234, then jumps back down
		// to the end of the grid.
		input: "aaaa\nbbbb\ncccc\x1b[2A\x1b[1B\r1234\x1b[1B",
		want:  "aaaa\n1234\ncccc",
	},
	{
		name:  "doesn't blow up if you go back too many characters",
		input: "this is good\x1b[100Dpoop and stuff",
		want:  "poop and stuff",
	},
	{
		name:  "doesn't blow up if you backspace too many characters",
		input: "hi\b\b\b\b\b\b\b\bbye",
		want:  "bye",
	},
	{
		name:  "ESC [1K clears everything before it",
		input: "hello\x1b[1Kfriend!",
		want:  "     friend!",
	},
	{
		name:  "clears everything after ESC [0K",
		input: "hello\nfriend!\x1b[A\r\x1b[0K",
		want:  "&nbsp;\nfriend!",
	},
	{
		name:  "handles ESC [0G",
		input: "hello friend\x1b[Ggoodbye buddy!",
		want:  "goodbye buddy!",
	},
	{
		name:  "preserves characters already written in a certain color",
		input: "  \x1b[90m․\x1b[0m\x1b[90m․\x1b[0m\x1b[0G\x1b[90m․\x1b[0m\x1b[90m․\x1b[0m",
		want:  "<span class=\"term-fgi90\">․․․․</span>",
	},
	{
		name:  "replaces empty lines with non-breaking spaces",
		input: "hello\n\nfriend",
		want:  "hello\n&nbsp;\nfriend",
	},
	{
		name:  "preserves opening colors when using ESC [0G",
		input: "\x1b[33mhello\x1b[0m\x1b[33m\x1b[44m\x1b[0Ggoodbye",
		want:  "<span class=\"term-fg33 term-bg44\">goodbye</span>",
	},
	{
		name:  "allows cursor movement with ESC [...H",
		input: "line 1\nline 2\nline 3\n\x1b[2;3Hm",
		// This should be:
		//   want:  "line 1\nlime 2\nline 3",
		// but because we can't implement it properly yet:
		want: "line 1\nline 2\nline 3\n  m",
	},
	{
		name:  "allows clearing lines below the current line",
		input: "foo\nbar\x1b[A\x1b[Jbaz",
		want:  "foobaz",
	},
	{
		name:  "doesn't freak out about clearing lines below when there aren't any",
		input: "foobar\x1b[0J",
		want:  "foobar",
	},
	{
		name:  "allows clearing lines above the current line",
		input: "foo\nbar\nbaz\x1b[A\x1b[1Jqux",
		want:  "&nbsp;\n   qux\nbaz",
	},
	{
		name:  "doesn't freak out about clearing lines above when there aren't any",
		input: "\x1b[1Jfoobar",
		want:  "foobar",
	},
	{
		name:  "allows clearing the entire scrollback buffer with ESC [2J",
		input: "this is a big long bit of terminal output\nplease pay it no mind, we will clear it soon\nokay, get ready for a disappearing act...\nand...and...\n\n\x1b[2Jhey presto",
		want:  "hey presto",
	},
	{
		name:  "allows clearing the entire scrollback buffer with ESC [3J also",
		input: "this is a big long bit of terminal output\nplease pay it no mind, we will clear it soon\nokay, get ready for a disappearing act...\nand...and...\n\n\x1b[2Jhey presto",
		want:  "hey presto",
	},
	{
		name:  "allows erasing the current line up to a point",
		input: "hello friend\x1b[1K!",
		want:  "            !",
	},
	{
		name:  "allows clearing of the current line",
		input: "hello friend\x1b[2K!",
		want:  "            !",
	},
	{
		name:  "doesn't close spans if no colors have been opened",
		input: "hello \x1b[0mfriend",
		want:  "hello friend",
	},
	{
		name:  "ESC [K correctly clears all previous parts of the string",
		input: "remote: Compressing objects:   0% (1/3342)\x1b[K\rremote: Compressing objects:   1% (34/3342)",
		want:  "remote: Compressing objects:   1% (34&#47;3342)",
	},
	{
		name:  "handles reverse linefeed",
		input: "meow\npurr\nnyan\x1bMrawr",
		want:  "meow\npurrrawr\nnyan",
	},
	{
		name:  "collapses many spans of the same color into 1",
		input: "\x1b[90m․\x1b[90m․\x1b[90m․\x1b[90m․\n\x1b[90m․\x1b[90m․\x1b[90m․\x1b[90m․",
		want:  "<span class=\"term-fgi90\">․․․․</span>\n<span class=\"term-fgi90\">․․․․</span>",
	},
	{
		name:  "escapes HTML",
		input: "hello <strong>friend</strong>",
		want:  "hello &lt;strong&gt;friend&lt;&#47;strong&gt;",
	},
	{
		name:  "escapes HTML in color codes",
		input: "hello \x1b[\"hellomfriend",
		want:  "hello [&quot;hellomfriend",
	},
	{
		name:  "handles background colors",
		input: "\x1b[30;42m\x1b[2KOK (244 tests, 558 assertions)",
		want:  "<span class=\"term-fg30 term-bg42\">OK (244 tests, 558 assertions)</span>",
	},
	{
		name:  "does not attempt to incorrectly nest CSS in HTML (https://github.com/buildkite/terminal-to-html/issues/36)",
		input: "Some plain text\x1b[0;30;42m yay a green background \x1b[0m\x1b[0;33;49mnow this has no background but is yellow \x1b[0m",
		want:  "Some plain text<span class=\"term-fg30 term-bg42\"> yay a green background </span><span class=\"term-fg33\">now this has no background but is yellow </span>",
	},
	{
		name:  "handles xterm colors",
		input: "\x1b[38;5;169;48;5;50mhello\x1b[0m \x1b[38;5;179mgoodbye",
		want:  "<span class=\"term-fgx169 term-bgx50\">hello</span> <span class=\"term-fgx179\">goodbye</span>",
	},
	{
		name:  "handles non-xterm codes on the same line as xterm colors",
		input: "\x1b[38;5;228;5;1mblinking and bold\x1b",
		want:  `<span class="term-fgx228 term-fg1 term-fg5">blinking and bold</span>`,
	},
	{
		name:  "ignores broken escape characters, stripping the escape rune itself",
		input: "hi amazing \x1b[12 nom nom nom friends",
		want:  "hi amazing [12 nom nom nom friends",
	},
	{
		name:  "handles colors with 3 attributes",
		input: "\x1b[0;10;4m\x1b[1m\x1b[34mgood news\x1b[0;10m\n\neveryone",
		want:  "<span class=\"term-fg34 term-fg1 term-fg4\">good news</span>\n&nbsp;\neveryone",
	},
	{
		name:  "ends underlining with ESC [24m",
		input: "\x1b[4mbegin\x1b[24m\r\nend",
		want:  "<span class=\"term-fg4\">begin</span>\nend",
	},
	{
		name:  "ends bold with ESC [21m",
		input: "\x1b[1mbegin\x1b[21m\r\nend",
		want:  "<span class=\"term-fg1\">begin</span>\nend",
	},
	{
		name:  "ends bold with ESC [22m",
		input: "\x1b[1mbegin\x1b[22m\r\nend",
		want:  "<span class=\"term-fg1\">begin</span>\nend",
	},
	{
		name:  "ends crossed out with ESC [29m",
		input: "\x1b[9mbegin\x1b[29m\r\nend",
		want:  "<span class=\"term-fg9\">begin</span>\nend",
	},
	{
		name:  "ends italic out with \x1b[23m",
		input: "\x1b[3mbegin\x1b[23m\r\nend",
		want:  "<span class=\"term-fg3\">begin</span>\nend",
	},
	{
		name:  "ends decreased intensity with \x1b[22m",
		input: "\x1b[2mbegin\x1b[22m\r\nend",
		want:  "<span class=\"term-fg2\">begin</span>\nend",
	},
	{
		name:  "ignores cursor show/hide",
		input: "\x1b[?25ldoing a thing without a cursor\x1b[?25h",
		want:  "doing a thing without a cursor",
	},
	{
		name:  "renders simple images on their own line", // http://iterm2.com/images.html
		input: "hi\x1b]1337;File=name=MS5naWY=;inline=1:AA==\ahello",
		want:  "hi\n" + `<img alt="1.gif" src="data:image/gif;base64,AA==">` + "\nhello",
	},
	{
		name:  "does not start a new line for iterm images if we're already at the start of a line",
		input: "\x1b]1337;File=name=MS5naWY=;inline=1:AA==\a",
		want:  `<img alt="1.gif" src="data:image/gif;base64,AA==">`,
	},
	{
		name:  "silently ignores unsupported ANSI escape sequences",
		input: "abc\x1b]9999\aghi",
		want:  "abcghi",
	},
	{
		name:  "correctly handles images that we decide not to render",
		input: "hi\x1b]1337;File=name=MS5naWY=;inline=0:AA==\ahello",
		want:  "hihello",
	},
	{
		name:  "renders external images",
		input: "\x1b]1338;url=http://foo.com/foobar.gif;alt=foo bar\a",
		want:  `<img alt="foo bar" src="http://foo.com/foobar.gif">`,
	},
	{
		name:  "disallows non-allow-listed schemes for images",
		input: "before\x1b]1338;url=javascript:alert(1);alt=hello\x07after",
		want:  "before\n&nbsp;\nafter", // don't really care about the middle, as long as it's white-spacey
	},
	{
		name:  "renders links, and renders them inline on other content",
		input: "a link to \x1b]1339;url=http://google.com;content=google\a.",
		want:  `a link to <a href="http://google.com">google</a>.`,
	},
	{
		name:  "renders OSC 8 links",
		input: "a link to \x1b]8;;http://google.com\x1b\\google\x1b]8;;\x1b\\.",
		want:  `a link to <a href="http://google.com">google</a>.`,
	},
	{
		name:  "uses URL as link content if missing",
		input: "\x1b]1339;url=http://google.com\a",
		want:  `<a href="http://google.com">http://google.com</a>`,
	},
	{
		name:  "protects inline images against XSS by escaping HTML during rendering",
		input: "hi\x1b]1337;File=name=" + base64Encode("<script>.pdf") + ";inline=1:AA==\ahello",
		want:  "hi\n" + `<img alt="&lt;script&gt;.pdf" src="data:application/pdf;base64,AA==">` + "\nhello",
	},
	{
		name:  "protects external images against XSS by escaping HTML during rendering",
		input: "\x1b]1338;url=\"https://example.com/a.gif&a=<b>&c='d'\";alt=foo&bar;width=\"<wat>\";height=2px\a",
		want:  `<img alt="foo&amp;bar" src="https://example.com/a.gif&amp;a=%3Cb%3E&amp;c=%27d%27" width="&lt;wat&gt;em" height="2px">`,
	},
	{
		name:  "protects links against XSS by escaping HTML during rendering",
		input: "\x1b]1339;url=\"https://example.com/a.gif&a=<b>&c='d'\";content=<h1>hello</h1>\a",
		want:  `<a href="https://example.com/a.gif&amp;a=%3Cb%3E&amp;c=%27d%27">&lt;h1&gt;hello&lt;/h1&gt;</a>`,
	},
	{
		name:  "protects OSC 8 links against XSS by escaping HTML during rendering",
		input: "a link to \x1b]8;;https://example.com/a.gif&a=<b>&c='d'\x1b\\<h1>hello</h1>\x1b]8;;\x1b\\.",
		want:  `a link to <a href="https://example.com/a.gif&amp;a=%3Cb%3E&amp;c=%27d%27">&lt;h1&gt;hello&lt;&#47;h1&gt;</a>.`,
	},
	{
		name:  "disallows javascript: scheme URLs",
		input: "\x1b]1339;url=javascript:alert(1);content=hello\x07",
		want:  `<a href="#">hello</a>`,
	},
	{
		name:  "disallows javascript: scheme URLs in OSC 8 links",
		input: "\x1b]8;;javascript:alert(1)\x07XSS!\x1b]8;;\x1b\\",
		want:  `<a href="#">XSS!</a>`,
	},
	{
		name:  "allows artifact: scheme URLs",
		input: "\x1b]1339;url=artifact://hello.txt\x07\n",
		want:  `<a href="artifact://hello.txt">artifact://hello.txt</a>`,
	},
	{
		name:  "allows artifact: scheme URLs in OSC 8 links",
		input: "\x1b]8;;artifact://hello.txt\x07the hello.txt artifact\x1b]8;;\x07\n",
		want:  `<a href="artifact://hello.txt">the hello.txt artifact</a>`,
	},
	{
		name:  "renders bk APC escapes followed by text",
		input: "\x1b_bk;t=123\x07hello",
		want:  `<time datetime="1970-01-01T00:00:00.123Z">1970-01-01T00:00:00.123Z</time>hello`,
	},
	{
		name:  "handles bk APC escapes surrounded by text",
		input: "hello \x1b_bk;t=123\x07world",
		want:  `<time datetime="1970-01-01T00:00:00.123Z">1970-01-01T00:00:00.123Z</time>hello world`,
	},
	{
		name:  "prefixes lines with the last timestamp seen",
		input: "hello\x1b_bk;t=123\x07 world\x1b_bk;t=456\x07!",
		want:  `<time datetime="1970-01-01T00:00:00.456Z">1970-01-01T00:00:00.456Z</time>hello world!`,
	},
	{
		name: "handles timestamps across multiple lines",
		input: strings.Join([]string{
			"hello\x1b_bk;t=123\x07 world\x1b_bk;t=234\x07!",
			"another\x1b_bk;t=345\x07 line\x1b_bk;t=456\x07!",
		}, "\n"),
		want: strings.Join([]string{
			`<time datetime="1970-01-01T00:00:00.234Z">1970-01-01T00:00:00.234Z</time>hello world!`,
			`<time datetime="1970-01-01T00:00:00.456Z">1970-01-01T00:00:00.456Z</time>another line!`,
		}, "\n"),
	},
	{
		name: "handles timestamps and delta timestamps",
		input: strings.Join([]string{
			"hello\x1b_bk;t=123\x07 world\x1b_bk;dt=111\x07!",
			"another\x1b_bk;dt=111\x07 line\x1b_bk;dt=111\x07!",
		}, "\n"),
		want: strings.Join([]string{
			`<time datetime="1970-01-01T00:00:00.234Z">1970-01-01T00:00:00.234Z</time>hello world!`,
			`<time datetime="1970-01-01T00:00:00.456Z">1970-01-01T00:00:00.456Z</time>another line!`,
		}, "\n"),
	},
}

func TestRendererAgainstCases(t *testing.T) {
	for _, c := range rendererTestCases {
		t.Run(c.name, func(t *testing.T) {
			got := Render([]byte(c.input))
			want := c.want

			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("Render(%q) diff (-got +want):\n%s", c.input, diff)
			}
		})
	}
}

func TestRendererAgainstFixtures(t *testing.T) {
	for _, base := range TestFiles {
		t.Run(fmt.Sprintf("for fixture %q", base), func(t *testing.T) {
			raw := loadFixture(t, base, "raw")
			want := string(loadFixture(t, base, "rendered"))

			got := Render(raw)

			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("Render diff (-got +want):\n%s", diff)
			}
		})
	}
}

func streamingRender(t testing.TB, raw []byte) string {
	var buf strings.Builder
	s, err := NewScreen(WithMaxSize(-1, 300))
	if err != nil {
		t.Fatalf("NewScreen error: %v", err)
	}
	s.ScrollOutFunc = func(line string) { fmt.Fprintln(&buf, line) }
	s.Write(raw)
	buf.WriteString(s.AsHTML())
	return buf.String()
}

func TestStreamingRendererAgainstCases(t *testing.T) {
	for _, c := range rendererTestCases {
		t.Run(c.name, func(t *testing.T) {
			got := streamingRender(t, []byte(c.input))
			want := c.want

			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("streamingRender(%q) diff (-got +want):\n%s", c.input, diff)
			}
		})
	}
}

func TestStreamingRendererAgainstFixtures(t *testing.T) {
	for _, base := range TestFiles {
		t.Run(fmt.Sprintf("for fixture %q", base), func(t *testing.T) {
			raw := loadFixture(t, base, "raw")
			want := string(loadFixture(t, base, "rendered"))

			got := streamingRender(t, raw)

			if diff := cmp.Diff(got, want); diff != "" {
				t.Errorf("streamingRender diff (-got +want):\n%s", diff)
			}
		})
	}
}

func TestScreenWriteToXY(t *testing.T) {
	s, err := NewScreen()
	if err != nil {
		t.Fatalf("NewScreen() error = %v", err)
	}
	s.write('a')

	s.x = 1
	s.y = 1
	s.write('b')

	s.x = 2
	s.y = 2
	s.write('c')

	output := s.AsHTML()
	expected := "a\n b\n  c"
	if output != expected {
		t.Errorf("got %q, wanted %q", output, expected)
	}
}

func BenchmarkRendererControl(b *testing.B)    { benchmarkRender("control.sh", b) }
func BenchmarkRendererCurl(b *testing.B)       { benchmarkRender("curl.sh", b) }
func BenchmarkRendererHomer(b *testing.B)      { benchmarkRender("homer.sh", b) }
func BenchmarkRendererITermLinks(b *testing.B) { benchmarkRender("itermlinks.sh", b) }
func BenchmarkRendererDockerPull(b *testing.B) { benchmarkRender("docker-pull.sh", b) }
func BenchmarkRendererPikachu(b *testing.B)    { benchmarkRender("pikachu.sh", b) }
func BenchmarkRendererPlaywright(b *testing.B) { benchmarkRender("playwright.sh", b) }
func BenchmarkRendererPowershell(b *testing.B) { benchmarkRender("pwsh.sh", b) }
func BenchmarkRendererRustFmt(b *testing.B)    { benchmarkRender("rustfmt.sh", b) }
func BenchmarkRendererWeather(b *testing.B)    { benchmarkRender("weather.sh", b) }
func BenchmarkRendererNpm(b *testing.B)        { benchmarkRender("npm.sh", b) }

func BenchmarkRendererDockerComposePull(b *testing.B) {
	benchmarkRender("docker-compose-pull.sh", b)
}

func benchmarkRender(filename string, b *testing.B) {
	raw := loadFixture(b, filename, "raw")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Render(raw)
	}
}

func BenchmarkStreamingControl(b *testing.B)    { benchmarkStreaming("control.sh", b) }
func BenchmarkStreamingCurl(b *testing.B)       { benchmarkStreaming("curl.sh", b) }
func BenchmarkStreamingHomer(b *testing.B)      { benchmarkStreaming("homer.sh", b) }
func BenchmarkStreamingITermLinks(b *testing.B) { benchmarkStreaming("itermlinks.sh", b) }
func BenchmarkStreamingDockerPull(b *testing.B) { benchmarkStreaming("docker-pull.sh", b) }
func BenchmarkStreamingPikachu(b *testing.B)    { benchmarkStreaming("pikachu.sh", b) }
func BenchmarkStreamingPlaywright(b *testing.B) { benchmarkStreaming("playwright.sh", b) }
func BenchmarkStreamingPowershell(b *testing.B) { benchmarkStreaming("pwsh.sh", b) }
func BenchmarkStreamingRustFmt(b *testing.B)    { benchmarkStreaming("rustfmt.sh", b) }
func BenchmarkStreamingWeather(b *testing.B)    { benchmarkStreaming("weather.sh", b) }
func BenchmarkStreamingNpm(b *testing.B)        { benchmarkStreaming("npm.sh", b) }

func BenchmarkStreamingDockerComposePull(b *testing.B) {
	benchmarkStreaming("docker-compose-pull.sh", b)
}

func benchmarkStreaming(filename string, b *testing.B) {
	raw := loadFixture(b, filename, "raw")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s, err := NewScreen(WithMaxSize(-1, 300))
		if err != nil {
			b.Fatalf("NewScreen(WithMaxSize(-1, 300)) error = %v", err)
		}
		// Set a non-nil scroll out func to exercise the codepath.
		s.ScrollOutFunc = func(line string) {}
		s.Write(raw)
		_ = s.AsHTML()
	}
}
