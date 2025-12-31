package rabbithole

// Stream protocol management. Requires the rabbitmq_stream_management plugin.

import (
	"net/http"
	"net/url"
)

// StreamConnectionInfo represents a stream protocol connection.
type StreamConnectionInfo struct {
	Name        string `json:"name"`
	Node        string `json:"node"`
	Vhost       string `json:"vhost"`
	User        string `json:"user"`
	State       string `json:"state"`
	Host        string `json:"host"`
	Port        Port   `json:"port"`
	PeerHost    string `json:"peer_host"`
	PeerPort    Port   `json:"peer_port"`
	ConnectedAt int64  `json:"connected_at"`
	FrameMax    int    `json:"frame_max"`
	Heartbeat   int    `json:"heartbeat"`
	SendOct     uint64 `json:"send_oct"`
	RecvOct     uint64 `json:"recv_oct"`
	Publishers  int    `json:"publishers"`
	Consumers   int    `json:"consumers"`
}

// StreamPublisherInfo represents a stream publisher.
type StreamPublisherInfo struct {
	ConnectionName    string `json:"connection_name"`
	ConnectionPid     string `json:"connection_pid"`
	Node              string `json:"node"`
	Vhost             string `json:"vhost"`
	Stream            string `json:"stream"`
	PublisherId       int    `json:"publisher_id"`
	Reference         string `json:"reference"`
	MessagesPublished uint64 `json:"messages_published"`
	MessagesConfirmed uint64 `json:"messages_confirmed"`
	MessagesErrored   uint64 `json:"messages_errored"`
}

// StreamConsumerInfo represents a stream consumer.
type StreamConsumerInfo struct {
	ConnectionName   string `json:"connection_name"`
	ConnectionPid    string `json:"connection_pid"`
	Node             string `json:"node"`
	Vhost            string `json:"vhost"`
	Stream           string `json:"stream"`
	SubscriptionId   int    `json:"subscription_id"`
	Credits          int    `json:"credits"`
	MessagesConsumed uint64 `json:"messages_consumed"`
	Offset           uint64 `json:"offset"`
	OffsetLag        uint64 `json:"offset_lag"`
	Active           bool   `json:"active"`
}

//
// GET /api/stream/connections
//

// ListStreamConnections returns all stream protocol connections.
func (c *Client) ListStreamConnections() (rec []StreamConnectionInfo, err error) {
	req, err := newGETRequest(c, "stream/connections")
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

//
// GET /api/stream/connections/{vhost}
//

// ListStreamConnectionsIn returns stream connections in a virtual host.
func (c *Client) ListStreamConnectionsIn(vhost string) (rec []StreamConnectionInfo, err error) {
	req, err := newGETRequest(c, "stream/connections/"+url.PathEscape(vhost))
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

//
// GET /api/stream/connections/{vhost}/{name}
//

// GetStreamConnection returns a stream connection by name.
func (c *Client) GetStreamConnection(vhost, name string) (rec *StreamConnectionInfo, err error) {
	req, err := newGETRequest(c, "stream/connections/"+url.PathEscape(vhost)+"/"+url.PathEscape(name))
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

//
// DELETE /api/stream/connections/{vhost}/{name}
//

// CloseStreamConnection closes a stream connection.
func (c *Client) CloseStreamConnection(vhost, name string) (res *http.Response, err error) {
	req, err := newRequestWithBody(c, "DELETE", "stream/connections/"+url.PathEscape(vhost)+"/"+url.PathEscape(name), nil)
	if err != nil {
		return nil, err
	}

	if res, err = executeRequest(c, req); err != nil {
		return nil, err
	}

	return res, nil
}

//
// GET /api/stream/publishers
//

// ListStreamPublishers returns all stream publishers.
func (c *Client) ListStreamPublishers() (rec []StreamPublisherInfo, err error) {
	req, err := newGETRequest(c, "stream/publishers")
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

//
// GET /api/stream/publishers/{vhost}
//

// ListStreamPublishersIn returns stream publishers in a virtual host.
func (c *Client) ListStreamPublishersIn(vhost string) (rec []StreamPublisherInfo, err error) {
	req, err := newGETRequest(c, "stream/publishers/"+url.PathEscape(vhost))
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

//
// GET /api/stream/publishers/{vhost}/{stream}
//

// ListStreamPublishersToStream returns publishers to a stream.
func (c *Client) ListStreamPublishersToStream(vhost, stream string) (rec []StreamPublisherInfo, err error) {
	req, err := newGETRequest(c, "stream/publishers/"+url.PathEscape(vhost)+"/"+url.PathEscape(stream))
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

//
// GET /api/stream/consumers
//

// ListStreamConsumers returns all stream consumers.
func (c *Client) ListStreamConsumers() (rec []StreamConsumerInfo, err error) {
	req, err := newGETRequest(c, "stream/consumers")
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

//
// GET /api/stream/consumers/{vhost}
//

// ListStreamConsumersIn returns stream consumers in a virtual host.
func (c *Client) ListStreamConsumersIn(vhost string) (rec []StreamConsumerInfo, err error) {
	req, err := newGETRequest(c, "stream/consumers/"+url.PathEscape(vhost))
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}
