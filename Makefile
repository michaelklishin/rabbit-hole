default: test

build:
	go build src/rabbitmq/http/client/*

test-install:
	go test -i src/rabbitmq/http/client/*

test: build test-install
	go test -test.v src/rabbitmq/http/client/*
