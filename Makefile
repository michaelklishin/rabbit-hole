default: test

build:
	go build

test-install:
	go test -i

test: build test-install
	go test -test.v
