default: all

ALL=go-mud-macOS-amd64      \
	go-mud-linux-amd64      \
	go-mud-linux-arm8       \
	go-mud-linux-arm7       \
	go-mud-windows-x86.exe  \
	go-mud-windows-amd64.exe

all: $(patsubst %,dist/%,$(ALL))

GOOPTS=-trimpath

SRCS=main.go

dist/go-mud-linux-amd64: $(SRC)
	GOOS=linux GOARCH=amd64 go build $(GOOPTS) -o $@ main.go

dist/go-mud-linux-arm7: $(SRC)
	GOOS=linux GOARM=7 GOARCH=arm go build $(GOOPTS) -o $@ main.go

dist/go-mud-linux-arm8: $(SRC)
	GOOS=linux GOARCH=arm64 go build $(GOOPTS) -o $@ main.go

dist/go-mud-macOS-amd64: $(SRC)
	GOOS=darwin GOARCH=amd64 go build $(GOOPTS) -o $@ main.go

dist/go-mud-windows-amd64.exe: $(SRC)
	GOOS=windows GOARCH=amd64 go build $(GOOPTS) -o $@ main.go

dist/go-mud-windows-x86.exe: $(SRC)
	GOOS=windows GOARCH=386 go build $(GOOPTS) -o $@ main.go

clean:
	rm -rf dist/
