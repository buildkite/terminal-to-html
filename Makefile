.PHONY: clean bench test

DEPS=*.go cmd/terminal-to-html/*.go

all: test terminal-to-html

bench:
	godep go test -bench . -benchmem

test:
	godep go test

clean:
	rm -f terminal-to-html
	rm -rf pkg

cmd/terminal-to-html/_bindata.go: assets/terminal.css
	go-bindata -o cmd/terminal-to-html/bindata.go -nomemcopy assets

terminal-to-html: $(DEPS)
	godep go build -o terminal-to-html cmd/terminal-to-html/*


