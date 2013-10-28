export GOPATH := $(CURDIR)

all: test

.PHONY: test

test: install-dependencies
	go test -v

install-dependencies:
	go get github.com/onsi/ginkgo
	go get github.com/onsi/gomega
	go get github.com/streadway/amqp
