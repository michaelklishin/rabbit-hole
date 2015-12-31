## Changes Between 1.0.0 and 1.1.0 (unreleased)

No changes yet.


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
