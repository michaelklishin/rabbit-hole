default: test

build:
	go build src/rabbitmq/http/client/*

test: build
	go test -test.v src/rabbitmq/http/client/*

