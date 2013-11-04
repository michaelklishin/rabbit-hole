/*
Rabbit Hole is a Go client for the RabbitMQ HTTP API.

All HTTP API operations are accessible via `rabbithole.Client`, which
should be instantiated with `rabbithole.NewClient`.

        // URI, username, password
        rmqc, _ = NewClient("http://127.0.0.1:15672", "guest", "guest")

Getting Overview

        res, err := rmqc.Overview()

Node and Cluster Status

        var err error

        // => []NodeInfo, err
        xs, err := rmqc.ListNodes()
*/
package rabbithole