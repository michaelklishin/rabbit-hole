package client

import (
	"net/http"
	"encoding/json"
)

type Client struct {
	Endpoint, Username, Password string
}

func NewClient(uri string, username string, password string) (*Client) {
	me := &Client{
		Endpoint: uri,
		Username: username,
	        Password: password}

	return me
}

func BuildPath(uri string, path string) string {
	return uri + "/api/" + path
}

type Version string
type Rate    float64
type IntRate uint64

type NodeName     string
type ProtocolName string

// TODO: custom deserializer
type IpAddress    string
type Port         uint64


type ExchangeType struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

type RateDetails struct {
	Rate        Rate   `json:"rate"`
}

type MessageStats struct {
	Publish        IntRate     `json:"publish"`
	PublishDetails RateDetails `json:"publish_details"`
}

type QueueTotals struct {
	Messages        uint64      `json:"messages"`
	MessagesDetails RateDetails `json:"messages_details"`

	MessagesReady        uint64      `json:"messages_ready"`
	MessagesReadyDetails RateDetails `json:"messages_ready_details"`

	MessagesUnacknowledged        uint64      `json:"messages_unacknowledged"`
	MessagesUnacknowledgedDetails RateDetails `json:"messages_unacknowledged_details"`
}

type ObjectTotals struct {
        Consumers   uint64   `json:"consumers"`
	Queues      uint64   `json:"queues"`
	Exchanges   uint64   `json:"exchanges"`
	Connections uint64   `json:"connections"`
	Channels    uint64   `json:"channels"`
}

type Listener struct {
	Node              NodeName       `json:"node"`
	Protocol          ProtocolName   `json:"protocol"`
	IpAddress         IpAddress      `json:"ip_address"`
	Port              Port           `json:"port"`
}

type BrokerContext struct {
	Node              NodeName       `json:"node"`
	Description       string         `json:"description"`
	Path              string         `json:"path"`
	Port              Port           `json:"port"`
	Ignore            bool           `json:"ignore_in_use"`
}



//
// GET /api/overview
//

type OverviewResponse struct {
	ManagementVersion Version         `json:"management_version"`
	StatisticsLevel   string          `json:"statistics_level"`
	RabbitMQVersion   Version         `json:"rabbitmq_version"`
	ErlangVersion     Version         `json:"erlang_version"`
	FullErlangVersion Version         `json:"erlang_full_version"`
	ExchangeTypes     []ExchangeType  `json:"exchange_types"`
	MessageStats      MessageStats    `json:"message_stats"`
	QueueTotals       QueueTotals     `json:"queue_totals"`
	ObjectTotals      ObjectTotals    `json:"object_totals"`
	Node              NodeName        `json:"node"`
	StatisticsDBNode  NodeName        `json:"statistics_db_node"`
	Listeners         []Listener      `json:"listeners"`
	Contexts          []BrokerContext `json:"contexts"`
}

func (c *Client) Overview () (OverviewResponse, error) {
	url := BuildPath(c.Endpoint, "overview")

	req, _ := http.NewRequest("GET", url, nil)
	req.SetBasicAuth(c.Username, c.Password)

	var err error

	httpc := &http.Client{}
	res, err := httpc.Do(req)
	if err != nil {
		return OverviewResponse{}, err
	}

	var rec OverviewResponse
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}