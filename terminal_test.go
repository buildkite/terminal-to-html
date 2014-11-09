package terminal

import (
	"fmt"
	"io/ioutil"
	"testing"
)

var TestFiles = []string{
	"control.sh",
	"curl.sh",
	"homer.sh",
	"pikachu.sh",
	"npm.sh",
}

func loadFixture(base string, ext string) []byte {
	filename := fmt.Sprintf("spec/fixtures/%s.%s", base, ext)
	data, err := ioutil.ReadFile(filename)
	check(err)
	return data
}

func TestRenderer(t *testing.T) {
	for _, base := range TestFiles {
		raw := loadFixture(base, "raw")
		expected := string(loadFixture(base, "rendered"))
		output := string(Render(raw))
		if output != expected {
			t.Fatalf("%s did not match\n\nGOT (len %d)\n%v\n\nEXPECTED (len %d)\n%v\n", base, len(output), output, len(expected), expected)
		}
	}
}

func BenchmarkRendererControl(b *testing.B) {
	benchmark("control.sh", b)
}

func BenchmarkRendererCurl(b *testing.B) {
	benchmark("curl.sh", b)
}

func BenchmarkRendererHomer(b *testing.B) {
	benchmark("homer.sh", b)
}

func BenchmarkRendererPikachu(b *testing.B) {
	benchmark("pikachu.sh", b)
}

func BenchmarkRendererNpm(b *testing.B) {
	benchmark("npm.sh", b)
}

func benchmark(filename string, b *testing.B) {
	raw := loadFixture(filename, "raw")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Render(raw)
	}
}
