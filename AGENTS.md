# Instructions for AI Agents

## What is This Codebase?

This is `rabbit-hole`, a Go client for the [RabbitMQ HTTP API](https://www.rabbitmq.com/docs/management#http-api).
It provides a single `Client` type with methods for managing and monitoring RabbitMQ clusters.
See [CONTRIBUTING.md](./CONTRIBUTING.md) for development setup details.

## Build

```shell
go build ./...
go vet ./...
go fmt ./...
```

## Run Tests

This library uses [Ginkgo](https://onsi.github.io/ginkgo/) for tests.

```shell
# run all tests (requires a running RabbitMQ node)
go test -v
```

Always run `go build ./...` and `go vet ./...` before making changes to verify the codebase compiles
and passes static analysis. If either fails, investigate and fix errors before proceeding with any modifications.

At the end of each task, run `go fmt ./...`.

### Test Node Setup

Tests are integration tests that require a locally running RabbitMQ node with the management
plugin enabled and specific virtual hosts, users, and plugins configured.

#### Use a Container

The following Make target starts a fully configured node in a container:

```shell
gmake docker.rabbitmq

# In a separate shell
go test -v
```

`gmake docker.rabbitmq` creates the `rabbit/hole` virtual host, the `policymaker` user, and enables
the `rabbitmq_federation`, `rabbitmq_federation_management`, `rabbitmq_shovel`, and
`rabbitmq_shovel_management` plugins.

#### BYON (Bring Your Own Node)

Alternatively, start a local RabbitMQ node any way you like, then run `bin/ci/before_build.sh` that will
set it up:

```shell
./bin/ci/before_build.sh
go test -v
```

This script enables the required plugins, reduces the stats emission interval, and creates the
virtual hosts and users the test suite expects. It requires `rabbitmqctl` to be in `PATH`, or the
`RABBITHOLE_RABBITMQCTL` environment variable to point to it.

## Repository Layout

All library code lives at the top level of the repository (single package, no subdirectories):

 * `client.go`: `Client` type, constructor functions (`NewClient`, `NewTLSClient`), request helpers
 * `error.go`: `ErrorResponse` type and error handling
 * `common.go`: shared data types and custom JSON unmarshalers
 * `doc.go`: package-level documentation with usage examples
 * Domain-specific files by resource type:
   * `vhosts.go`, `vhost_limits.go`, `vhosts_channels.go`, `vhosts_connections.go`
   * `users.go`, `permissions.go`, `topic_permissions.go`, `user_limits.go`
   * `exchanges.go`, `queues.go`, `bindings.go`, `consumers.go`
   * `policies.go`, `operator_policies.go`
   * `nodes.go`, `cluster.go`, `health_checks.go`
   * `federation.go`, `federation_links.go`, `shovels.go`, `streams.go`
   * `runtime_parameters.go`, `global_parameters.go`, `feature_flags.go`
   * `definitions.go`, `plugins.go`, `deprecated_features.go`, `misc.go`
 * `*_test.go`: integration and unit tests using [Ginkgo v2](https://onsi.github.io/ginkgo/) and [Gomega](https://onsi.github.io/gomega/)
 * `go.mod`: module definition (`github.com/michaelklishin/rabbit-hole/v3`)
 * `Makefile`: common development targets
 * `bin/ci/before_build.sh`: configures an existing RabbitMQ node for the test suite (plugins, vhosts, users, stats interval)

## Code Conventions

### Naming

 * Types: `PascalCase` — `VhostInfo`, `QueueSettings`, `ExchangeInfo`
 * `*Info` suffix for read-only API response structs (e.g. `QueueInfo`, `NodeInfo`)
 * `*Settings` suffix for mutable request structs used to create or update resources (e.g. `QueueSettings`, `VhostSettings`)
 * Client methods: verb-first `PascalCase` — `ListQueues`, `GetExchange`, `DeclareQueue`, `DeleteBinding`
 * JSON struct tags: `snake_case` (matching RabbitMQ HTTP API conventions)

### Operation Structure

Each resource domain follows a consistent pattern:

```go
// List/Get — HTTP GET, returns parsed response
func (c *Client) ListQueues() (rec []QueueInfo, err error)
func (c *Client) GetQueue(vhost, queue string) (rec *QueueInfo, err error)

// Create/Update — HTTP PUT with JSON body, returns raw *http.Response
func (c *Client) DeclareQueue(vhost string, queue string, info QueueSettings) (res *http.Response, err error)

// Delete — HTTP DELETE, returns raw *http.Response; 404 is treated as success (idempotent)
func (c *Client) DeleteQueue(vhost, queue string) (res *http.Response, err error)
```

### Custom JSON Unmarshaling

The library handles RabbitMQ API quirks with custom unmarshalers for types that can appear
as multiple JSON types: `Port` (int, string, or `"undefined"`), `URISet` (string or array),
`AutoDelete` (bool or `"undefined"`), `UserTags` (comma-separated string or array).
Follow the same pattern in `common.go` when adding new fields with ambiguous types.

## Test Suite Layout

 * `rabbithole_suite_test.go`: Ginkgo suite bootstrap (`TestRabbitHole`) and shared setup
 * `rabbithole_test.go`: main integration test suite (requires a running RabbitMQ node)
 * `unit_test.go`: Ginkgo specs for custom JSON unmarshaling; can be isolated with `--ginkgo.focus="Unit tests"`
 * `health_checks_test.go`: health check endpoint tests
 * `policies_test.go`: policy and operator policy tests
 * `deprecated_features_test.go`: Ginkgo specs for `DeprecationPhase` JSON unmarshaling

### Running Specific Tests

All specs run under the single `TestRabbitHole` suite function. Use Ginkgo's `--ginkgo.focus` flag
to filter by description substring:

```shell
# Run all specs
go test -v -run TestRabbitHole

# Run specs whose description contains "ListQueues"
go test -v -run TestRabbitHole --ginkgo.focus="ListQueues"

# Run only unit specs
go test -v -run TestRabbitHole --ginkgo.focus="Unit tests"
```

The Ginkgo CLI can also be used. Install it first (match the version in `go.mod`):

```shell
go install github.com/onsi/ginkgo/v2/ginkgo@v2.28.2
ginkgo --focus="ListQueues" -v
```

## Source of Domain Knowledge

 * [RabbitMQ HTTP API Reference](https://www.rabbitmq.com/docs/http-api-reference)
 * [RabbitMQ Documentation](https://www.rabbitmq.com/docs/)
 * Rust sibling client: `../rabbitmq-http-api-client-rs.git` — reference for API coverage and design

Treat the RabbitMQ documentation as the ultimate first party source of truth.

## Change Log

If asked to perform change log updates, consult and modify `ChangeLog.md` and stick to its
existing writing style (`## Changes Between X.Y.Z and A.B.C (date)` headings).

## Releases

### How to Roll (Produce) a New Release

Suppose `ChangeLog.md` has a `## Changes Between 3.N.0 and 3.(N+1).0 (in development)` section at the top.

To produce a new release:

 1. Update the changelog: replace `(in development)` with today's date, e.g. `(Apr 27, 2026)`. Make sure all notable changes since the previous release are listed
 2. Commit with the message `3.(N+1).0` (just the version number, nothing else)
 3. Tag the commit: `git tag v3.(N+1).0`
 4. Add a new `## Changes Between 3.(N+1).0 and 3.(N+2).0 (in development)` section to `ChangeLog.md` with `No changes yet.` underneath
 5. Commit with the message `Bump dev version`
 6. Push: `git push && git push --tags`

Go module consumers resolve the version from the git tag. No additional publish step is needed.

### GitHub Actions

For verifying YAML file syntax, use `yq`, Ruby or Python YAML modules (whichever is available).

## Comments

 * Only add comments that express non-obvious intent — hidden constraints, subtle invariants, or workarounds for specific bugs
 * Keep comments concise

## Git Commits

 * Do not commit changes automatically without explicit permission to do so
 * Never add yourself as a git commit coauthor
 * Never mention yourself in commit messages in any way (no "Generated by", no AI tool links, etc)

## Style Guide

 * Never add full stops to Markdown list items
 * Use backticks for type names, method names, field names, and file names
 * "virtual host" in documentation, even though `vhost` is used in code

## After Completing a Task

After completing a task, perform up to twenty iterative reviews of your changes.
In every iteration, look for meaningful improvements that were missed, for gaps in test coverage,
and for deviations from the instructions in this file.

In particular, check that:

 * New client methods follow the established naming and return-type conventions (`*Info` for reads, `*http.Response` for writes)
 * Custom JSON unmarshaling is added for any field that the RabbitMQ API returns as multiple possible types
 * Notable user-visible changes are listed in `ChangeLog.md` under the current `(in development)` section
 * New behavior is covered by integration tests in `rabbithole_test.go` or a dedicated `*_test.go` file
 * No new external runtime dependencies were introduced beyond what is already in `go.mod`

If no meaningful improvements are found for three iterations in a row,
report it and stop iterating.
