## Changes Between 2.4.0 and 2.5.0 (in development)

### Shovels: Support for Numerical Delete-After Values

The `delete-after` Shovel parameter now can be deserialised to
a numerical TTL value as well as special string values such as `"never"`.

Contributed by Michal @mkuratczyk Kuratczyk.

GitHub issue: [#164](https://github.com/michaelklishin/rabbit-hole/pull/164)


## Changes Between 2.3.0 and 2.4.0 (Aug 4th, 2020)

### More Thorough Error Checking of HTTP[S] Requests

Suggested by @mammothbane.

GitHub issue: [#158](https://github.com/michaelklishin/rabbit-hole/issues/158)

### Salt Generation Helper Now Uses `crypto/rand` Instead of `math/rand`

Suggested by @mammothbane.

GitHub issue: [#160](https://github.com/michaelklishin/rabbit-hole/issues/160)

## More Standardized Response Errors

Error responses (`40x` with the exception of `404` in response to `DELETE` operations,
`50x`) HTTP API response errors are now always wrapped into`ErrorResponse`,
even if they do not carry a JSON payload.


## Changes Between 2.2.0 and 2.3.0 (July 11th, 2020)

### New Endpoints for Listing Federation Links

Contributed by @niclic.

GitHub issue: [#155](https://github.com/michaelklishin/rabbit-hole/pull/155)

### Support for More Shovel Parameters (e.g. for AMQP 1.0 Sources and Destinations)

Contributed by @akurz.

GitHub issue: [#155](https://github.com/michaelklishin/rabbit-hole/pull/157)

### Conditional Exclusion of Expiration Field

Contributed by @niclic.

GitHub issue: [#154](https://github.com/michaelklishin/rabbit-hole/pull/154)


## Changes Between 2.1.0 and 2.2.0 (May 21st, 2020)

### [Runtime Parameter](https://www.rabbitmq.com/parameters.html) and [Federation Upstream](https://www.rabbitmq.com/federation.html) Management

Contributed by @niclic.

GitHub issue: [#150](https://github.com/michaelklishin/rabbit-hole/pull/150)

### Improved Error Reporting

Contributed by @niclic.

GitHub issue: [michaelklishin/rabbit-hole#152](https://github.com/michaelklishin/rabbit-hole/pull/152)

### Fixed a null Pointer in HTTP Response Handling

Contributed by @justabaka.

GitHub issue: [#148](https://github.com/michaelklishin/rabbit-hole/pull/148)


## Changes Between 2.0.0 and 2.1.0 (Feb 1st, 2020)

### Corrects Package Version

See [Semantic Go Module Import Versioning](https://github.com/golang/go/wiki/Modules#semantic-import-versioning) for details

GitHub issue: [#146](https://github.com/michaelklishin/rabbit-hole/issues/146)

### New Endpoint, `DELETE /topic-permissions/{vhost}/{user}/{exchange}`

Contributed by Barnaby Shearer.

GitHub issues: [#147](https://github.com/michaelklishin/rabbit-hole/pull/147)

### Exposed Client Connection Time Field

Available in RabbitMQ 3.7 and later versions.

Contributed by @kgrieco.

GitHub issue: [#144](https://github.com/michaelklishin/rabbit-hole/pull/144)

### Authentication Failures Now Return a Reasonable Error

Contributed by @mazamats.

GitHub issues: [#145](https://github.com/michaelklishin/rabbit-hole/pull/145), [#112](https://github.com/michaelklishin/rabbit-hole/issues/112)


## Changes Between 1.5.0 and 2.0.0 (October 8th, 2019)

### Go 1.9 through 1.11 Support Dropped

This library now only supports Go 1.12 and 1.13 (two most recent minor GA releases).

### Unroutable Message Metric Support

The `drop_unroutable` metric is specific to RabbitMQ 3.8.

Contributed by David Ansari and Feroz Jilla.

### Support for Exchange Ingress and Egress Rates

Contributed by Rajendra N Acharya.

### Eager Synchronization of Classic Queue

It is now possible to initiate an eager sync of a classic mirrored queue and cancel it.

Contributed by Jaroslaw Bochniak.

GitHub issue: [#143](https://github.com/michaelklishin/rabbit-hole/pull/143)

### Queue Status JSON Serialization Fixed

Contributed by Andrew Wang.

### GET /api/consumers Support

Contributed by Thomas Hudry.

GitHub issue: [#140](https://github.com/michaelklishin/rabbit-hole/pull/140)

### http.Transport Replaced by http.RoundTripper

HTTP client configuration now uses `http.RoundTripper`.

GitHub issue: [#123](https://github.com/michaelklishin/rabbit-hole/pull/123).

Contributed by Radek Simko.

### Go Modules Support

GitHub issues: [#124](https://github.com/michaelklishin/rabbit-hole/pull/124), [#128](https://github.com/michaelklishin/rabbit-hole/pull/128).

Contributed by Radek Simko and Gerhard Lazu.


## Changes Between 1.4.0 and 1.5.0 (February 13th, 2019)

### More Binding Management Functions

`ListExchangeBindings`, `ListExchangeBindingsWithSource`, `ListExchangeBindingsWithDestination`,
and `ListExchangeBindingsBetween` are new functions that list bindings,
in particular between exchanges.

GitHub issue: [#109](https://github.com/michaelklishin/rabbit-hole/pull/109).

### Password Hash Generation Helpers

It is now possible to specify a `password_hash` when creating a user.
Helper functions `GenerateSalt` and `SaltedPasswordHashSHA256` make this more
straightforward compared to implementing [the algorithm](http://www.rabbitmq.com/passwords.html#computing-password-hash)
directly.

GitHub issue: [#119](https://github.com/michaelklishin/rabbit-hole/pull/119)

### Paginated Queue Listing

A new function, `PagedListQueuesWithParameters`, can list queues with pagination support.

GitHub issue: [#118](https://github.com/michaelklishin/rabbit-hole/pull/118)

### More `NodeInfo` and `QueueInfo` Attributes

GitHub issue: [#115](https://github.com/michaelklishin/rabbit-hole/issues/115)

### URL.Opaque Left to Its Own Devices

The client no longer messes with `URL.Opaque` as it doesn't seem to
be necessary any more for correct %-encoding of URL path.

GitHub issue: [#121](https://github.com/michaelklishin/rabbit-hole/issues/121)


## Changes Between 1.0.0 and 1.1.0 (Dec 1st, 2015)

### More Complete Message Stats Information

Message stats now include fields such as `deliver_get` and `redeliver`.

GH issue: [#73](https://github.com/michaelklishin/rabbit-hole/pull/73).

Contributed by Edward Wilde.


## 1.0 (first tagged release, Dec 25th, 2015)

### TLS Support

`rabbithole.NewTLSClient` is a new function which works
much like `NewClient` but additionally accepts a transport.

Contributed by @[GrimTheReaper](https://github.com/GrimTheReaper).

### Federation Support

It is now possible to create federation links
over HTTP API.

Contributed by [Ryan Grenz](https://github.com/grenzr-bskyb).

### Core Operations Support

Most common HTTP API operations (listing and management of
vhosts, users, permissions, queues, exchanges, and bindings)
are supported by the client.
