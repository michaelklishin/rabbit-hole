module github.com/michaelklishin/rabbit-hole/v3

require (
	github.com/onsi/ginkgo/v2 v2.32.0
	github.com/onsi/gomega v1.42.1
	github.com/rabbitmq/amqp091-go v1.13.0
)

require (
	github.com/Masterminds/semver/v3 v3.5.0 // indirect
	github.com/go-logr/logr v1.4.4 // indirect
	github.com/go-task/slim-sprig/v3 v3.0.0 // indirect
	github.com/google/go-cmp v0.7.0 // indirect
	github.com/google/pprof v0.0.0-20260402051712-545e8a4df936 // indirect
	go.yaml.in/yaml/v3 v3.0.5 // indirect
	golang.org/x/mod v0.39.0 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sync v0.22.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	golang.org/x/tools v0.48.0 // indirect
)

go 1.25.0

retract (
	v3.1.1 // Used to retract v3.0.0
	v3.0.0 // Incorrect module name, do not use
)
