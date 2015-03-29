all: test terminal-to-html

bench:
	godep go test -bench . -benchmem

test:
	godep go test

cmd/terminal-to-html/_bindata.go: assets/terminal.css
	go-bindata -o cmd/terminal-to-html/_bindata.go -nomemcopy assets

terminal-to-html: cmd/terminal-to-html/_bindata.go
	godep go build -o terminal-to-html cmd/terminal-to-html/*


