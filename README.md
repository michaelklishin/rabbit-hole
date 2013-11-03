# Rabbit Hole, a RabbitMQ HTTP API Client for Go

This library is a [RabbitMQ HTTP API](http://hg.rabbitmq.com/rabbitmq-management/raw-file/450b7ea22cfd/priv/www/api/index.html) client for the Go language.

## Supported Go Versions

Rabbit Hole requires Go 1.1+.


## Supported RabbitMQ Versions

 * RabbitMQ 3.x
 * RabbitMQ 2.x

All versions require [RabbitMQ Management UI plugin](http://www.rabbitmq.com/management.html) to be installed and enabled.


## Project Maturity

Rabbit Hole is a **very immature** project and breaking changes are quite likely.
It also completely lacks documentation and at the moment likely isn't useful to
anyone but the author.


## Installation

```
go get github.com/michaelklishin/rabbit-hole
```


## Documentation

### Overview

To import the package:

``` go
import (
       "github.com/michaelklishin/rabbit-hole"
)
```

All HTTP API operations are accessible via `rabbithole.Client`, which
should be instantiated with `rabbithole.NewClient`:

``` go
// URI, username, password
rmqc, _ = NewClient("http://127.0.0.1:15672", "guest", "guest")
```

### Getting Overview

``` go
res, err := rmqc.Overview()
```

### Node and Cluster Status

``` go
var err error

// => []NodeInfo, err
xs, err := rmqc.ListNodes()
```


### Operations on Connections

TBD


### Operations on Channels

TBD


### Operations on Exchanges

TBD


### Operations on Queues

TBD


### Operations on Bindings

TBD


### Operations on Vhosts

TBD


### Managing Users

TBD


### Managing Permissions

TBD




## Contributing

1. Fork it
2. Create your feature branch (`git checkout -b my-new-feature`)
3. Commit your changes (`git commit -am 'Add some feature'`)
4. Push to the branch (`git push -u origin my-new-feature`)
5. Create new Pull Request


## License & Copyright

2-clause BSD license.

(c) Michael S. Klishin, 2013.
