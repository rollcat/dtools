.PHONY: all
all: dlaunch

dlaunch: cmd/dlaunch/*.go
	go build -o dlaunch cmd/dlaunch/*.go

.PHONY: clean
clean:
	rm -f dlaunch
