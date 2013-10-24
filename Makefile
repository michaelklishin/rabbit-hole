default: test

install-dependencies:
	go get github.com/onsi/ginkgo
	go get github.com/onsi/gomega
	go get github.com/streadway/amqp

build: install-dependencies
	go build

test-install:
	go test -i

test: build test-install
	go test -test.v
