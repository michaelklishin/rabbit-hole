module github.com/michaelklishin/rabbit-hole/v3

require (
	github.com/onsi/ginkgo/v2 v2.27.2
	github.com/onsi/gomega v1.38.3
	github.com/rabbitmq/amqp091-go v1.10.0
)

require (
	github.com/Masterminds/semver/v3 v3.4.0 // indirect
	github.com/go-logr/logr v1.4.3 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20250820193118-f64d9cf942d6 // indirect
	go.yaml.in/yaml/v3 v3.0.4 // indirect
	golang.org/x/mod v0.30.0 // indirect
	golang.org/x/net v0.47.0 // indirect
	golang.org/x/sync v0.18.0 // indirect
	golang.org/x/sys v0.39.0 // indirect
	golang.org/x/text v0.31.0 // indirect
	golang.org/x/tools v0.39.0 // indirect
	gopkg.in/check.v1 v1.0.0-20201130134442-10cb98267c6c // indirect
)

go 1.24.0

toolchain go1.24.1

retract (
	v3.1.1 // Used to retract v3.0.0
	v3.0.0 // Incorrect module name, do not use
)
