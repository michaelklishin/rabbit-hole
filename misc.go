package rabbithole

import "encoding/json"

//
// GET /api/overview
//

type QueueTotals struct {
	Messages        int         `json:"messages"`
	MessagesDetails RateDetails `json:"messages_details"`

	MessagesReady        int         `json:"messages_ready"`
	MessagesReadyDetails RateDetails `json:"messages_ready_details"`

	MessagesUnacknowledged        int         `json:"messages_unacknowledged"`
	MessagesUnacknowledgedDetails RateDetails `json:"messages_unacknowledged_details"`
}

type ObjectTotals struct {
	Consumers   int `json:"consumers"`
	Queues      int `json:"queues"`
	Exchanges   int `json:"exchanges"`
	Connections int `json:"connections"`
	Channels    int `json:"channels"`
}

type Listener struct {
	Node      string `json:"node"`
	Protocol  string `json:"protocol"`
	IpAddress string `json:"ip_address"`
	Port      Port   `json:"port"`
}

type Overview struct {
	ManagementVersion string          `json:"management_version"`
	StatisticsLevel   string          `json:"statistics_level"`
	RabbitMQVersion   string          `json:"rabbitmq_version"`
	ErlangVersion     string          `json:"erlang_version"`
	FullErlangVersion string          `json:"erlang_full_version"`
	ExchangeTypes     []ExchangeType  `json:"exchange_types"`
	MessageStats      MessageStats    `json:"message_stats"`
	QueueTotals       QueueTotals     `json:"queue_totals"`
	ObjectTotals      ObjectTotals    `json:"object_totals"`
	Node              string          `json:"node"`
	StatisticsDBNode  string          `json:"statistics_db_node"`
	Listeners         []Listener      `json:"listeners"`
	Contexts          []BrokerContext `json:"contexts"`
}

func (c *Client) Overview() (Overview, error) {
	var err error
	req, err := NewGETRequest(c, "overview")

	if err != nil {
		return Overview{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return Overview{}, err
	}

	var rec Overview
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}

//
// GET /api/whoami
//

type WhoamiInfo struct {
	Name        string `json:"name"`
	Tags        string `json:"tags"`
	AuthBackend string `json:"auth_backend"`
}

func (c *Client) Whoami() (rec *WhoamiInfo, err error) {
	req, err := NewGETRequest(c, "whoami")
	if err != nil {
		return nil, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return nil, err
	}

	decoder := json.NewDecoder(res.Body)
	defer res.Body.Close()
	err = decoder.Decode(&rec)
	if err != nil {
		return nil, err
	}

	return rec, nil
}
