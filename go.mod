module github.com/michaelklishin/rabbit-hole/v3

require (
	github.com/onsi/ginkgo/v2 v2.23.1
	github.com/onsi/gomega v1.36.2
	github.com/rabbitmq/amqp091-go v1.10.0
)

require (
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/google/pprof v0.0.0-20241210010833-40e02aabc2ad // indirect
	golang.org/x/net v0.36.0 // indirect
	golang.org/x/sys v0.30.0 // indirect
	golang.org/x/text v0.22.0 // indirect
	golang.org/x/tools v0.30.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

go 1.23.0

toolchain go1.24.1

retract (
	v3.1.1 // Used to retract v3.0.0
	v3.0.0 // Incorrect module name, do not use
)
