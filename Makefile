default: all

ALL=go-mud-linux-amd64 go-mud-linux-arm64 go-mud-macOS-amd64 go-mud-windows-amd64.exe
all: $(ALL)

SRCS=main.go

go-mud-linux-amd64: $(SRC)
	GOOS=linux GOARCH=amd64 go build -o $@ main.go

go-mud-linux-arm64: $(SRC)
	GOOS=linux GOARCH=arm64 go build -o $@ main.go

go-mud-macOS-amd64: $(SRC)
	GOOS=darwin GOARCH=amd64 go build -o $@ main.go

go-mud-windows-amd64.exe: $(SRC)
	GOOS=windows GOARCH=amd64 go build -o $@ main.go

zip: all
	zip go-mud.zip go-mud-{linux,macOS,windows}-* *.lua

clean:
	rm -f $(ALL)
