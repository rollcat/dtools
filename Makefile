.PHONY: all
all: dlaunch dxkbmap

dlaunch: cmd/dlaunch/*.go
	go build -o dlaunch cmd/dlaunch/*.go

dxkbmap: cmd/dxkbmap/*.go
	go build -o dxkbmap cmd/dxkbmap/*.go

.PHONY: clean
clean:
	rm -f dlaunch dxkbmap
