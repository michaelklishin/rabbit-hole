package rabbithole

import (
	"encoding/json"
	"net/url"
)

//
// GET /api/exchanges
//

type IngressEgressStats struct {
	PublishIn        int         `json:"publish_in"`
	PublishInDetails RateDetails `json:"publish_in_details"`

	PublishOut        int         `json:"publish_out"`
	PublishOutDetails RateDetails `json:"publish_out_details"`
}

type ExchangeInfo struct {
	Name       string                 `json:"name"`
	Vhost      string                 `json:"vhost"`
	Type       string                 `json:"type"`
	Durable    bool                   `json:"durable"`
	AutoDelete bool                   `json:"auto_delete"`
	Internal   bool                   `json:"internal"`
	Arguments  map[string]interface{} `json:"arguments"`

	MessageStats IngressEgressStats `json:"message_stats"`
}

func (c *Client) ListExchanges() ([]ExchangeInfo, error) {
	var err error
	req, err := NewGETRequest(c, "exchanges")
	if err != nil {
		return []ExchangeInfo{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return []ExchangeInfo{}, err
	}

	var rec []ExchangeInfo
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}

//
// GET /api/exchanges/{vhost}
//

func (c *Client) ListExchangesIn(vhost string) ([]ExchangeInfo, error) {
	var err error
	req, err := NewGETRequest(c, "exchanges/"+url.QueryEscape(vhost))
	if err != nil {
		return []ExchangeInfo{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return []ExchangeInfo{}, err
	}

	var rec []ExchangeInfo
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}

//
// GET /api/exchanges/{vhost}/{name}
//

// Example response:
//
// {
//   "incoming": [
//     {
//       "stats": {
//         "publish": 2760,
//         "publish_details": {
//           "rate": 20
//         }
//       },
//       "channel_details": {
//         "name": "127.0.0.1:46928 -> 127.0.0.1:5672 (2)",
//         "number": 2,
//         "connection_name": "127.0.0.1:46928 -> 127.0.0.1:5672",
//         "peer_port": 46928,
//         "peer_host": "127.0.0.1"
//       }
//     }
//   ],
//   "outgoing": [
//     {
//       "stats": {
//         "publish": 1280,
//         "publish_details": {
//           "rate": 20
//         }
//       },
//       "queue": {
//         "name": "amq.gen-7NhO_yRr4lDdp-8hdnvfuw",
//         "vhost": "rabbit\/hole"
//       }
//     }
//   ],
//   "message_stats": {
//     "publish_in": 2760,
//     "publish_in_details": {
//       "rate": 20
//     },
//     "publish_out": 1280,
//     "publish_out_details": {
//       "rate": 20
//     }
//   },
//   "name": "amq.fanout",
//   "vhost": "rabbit\/hole",
//   "type": "fanout",
//   "durable": true,
//   "auto_delete": false,
//   "internal": false,
//   "arguments": {
//   }
// }

type ExchangeIngressDetails struct {
	Stats          MessageStats      `json:"stats"`
	ChannelDetails PublishingChannel `json:"channel_details"`
}

type PublishingChannel struct {
	Number         int    `json:"number"`
	Name           string `json:"name"`
	ConnectionName string `json:"connection_name"`
	PeerPort       Port   `json:"peer_port"`
	PeerHost       string `json:"peer_host"`
}

type NameAndVhost struct {
	Name  string `json:"name"`
	Vhost string `json:"vhost"`
}

type ExchangeEgressDetails struct {
	Stats MessageStats `json:"stats"`
	Queue NameAndVhost `json:"queue"`
}

type DetailedExchangeInfo struct {
	Name       string                 `json:"name"`
	Vhost      string                 `json:"vhost"`
	Type       string                 `json:"type"`
	Durable    bool                   `json:"durable"`
	AutoDelete bool                   `json:"auto_delete"`
	Internal   bool                   `json:"internal"`
	Arguments  map[string]interface{} `json:"arguments"`

	Incoming ExchangeIngressDetails `json:"incoming"`
	Outgoing ExchangeEgressDetails  `json:"outgoing"`
}

func (c *Client) GetExchange(vhost, exchange string) (DetailedExchangeInfo, error) {
	var err error
	req, err := NewGETRequest(c, "exchanges/"+url.QueryEscape(vhost)+"/"+exchange)
	if err != nil {
		return DetailedExchangeInfo{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return DetailedExchangeInfo{}, err
	}

	var rec DetailedExchangeInfo
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}
