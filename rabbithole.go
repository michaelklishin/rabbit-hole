package rabbithole

import (
	"net/http"
	"encoding/json"
)

type Client struct {
	Endpoint, Username, Password string
}

type Version string
type Rate    float64
type IntRate int32

// TODO: this probably should be fixed in RabbitMQ management plugin
type OsPid   string

type NodeName       string
type ProtocolName   string
type ConnectionName string

type Username       string
type VhostName      string

type Properties map[string]interface{}

// TODO: custom deserializer
type IpAddress    string
type Port         int


type NameDescriptionEnabled struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

type NameDescriptionVersion struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Version     Version `json:"version"`
}

type ExchangeType NameDescriptionEnabled

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
	Channels    uint32   `json:"channels"`
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

type Overview struct {
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

func (c *Client) Overview () (Overview, error) {
	var err error
	req, err := NewHTTPRequest(c, "GET", "overview")

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
// GET /api/nodes
//

type AuthMechanism NameDescriptionEnabled
type ErlangApp     NameDescriptionVersion

type NodeInfo struct {
	Name      NodeName `json:"name"`
	NodeType  string   `json:"type"`
	IsRunning bool     `json:"running"`
	OsPid     OsPid    `json:"os_pid"`

	FdUsed       uint32  `json:"fd_used"`
	FdTotal      uint32  `json:"fd_total"`
	SocketsUsed  uint32  `json:"sockets_used"`
	SocketsTotal uint32  `json:"sockets_total"`
	MemUsed      uint64  `json:"mem_used"`
	MemLimit     uint64  `json:"mem_limit"`

	MemAlarm      bool   `json:"mem_alarm"`
	DiskFreeAlarm bool   `json:"disk_free_alarm"`

	ExchangeTypes  []ExchangeType  `json:"exchange_types"`
	AuthMechanisms []AuthMechanism `json:"auth_mechanisms"`
	ErlangApps     []ErlangApp     `json:"applications"`
	Contexts       []BrokerContext `json:"contexts"`
}


func (c *Client) ListNodes() ([]NodeInfo, error) {
	var err error
	req, err := NewHTTPRequest(c, "GET", "nodes")
	if err != nil {
		return []NodeInfo{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return []NodeInfo{}, err
	}

	var rec []NodeInfo
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}

//
// GET /api/connections
//

type ConnectionInfo struct {
        Name           ConnectionName      `json:"name"`
	Node           NodeName            `json:"node"`
	Channels       uint32              `json:"channels"`
	State          string              `json:"state"`
	Type           string              `json:"type"`

	Port           Port                `json:"port"`
	PeerPort       Port                `json:"peer_port"`

	Host           string              `json:"host"`
	PeerHost       string              `json:"peer_host"`

	LastBlockedBy  string              `json:"last_blocked_by"`
	LastBlockedAge string              `json:"last_blocked_age"`

	UsesTLS          bool              `json:"ssl"`
	PeerCertSubject  string            `json:"peer_cert_subject"`
	PeerCertValidity string            `json:"peer_cert_validity"`
	PeerCertIssuer   string            `json:"peer_cert_issuer"`

	SSLProtocol      string            `json:"ssl_protocol"`
	SSLKeyExchange   string            `json:"ssl_key_exchange"`
	SSLCipher        string            `json:"ssl_cipher"`
	SSLHash          string            `json:"ssl_hash"`

	Protocol         ProtocolName      `json:"protocol"`
	User             Username          `json:"user"`
	Vhost            VhostName         `json:"vhost"`

	Timeout           uint32           `json:"timeout"`
	FrameMax          uint32           `json:"frame_max"`

	ClientProperties  Properties       `json:"client_properties"`


	RecvOct      uint64          `json:"recv_oct"`
	SendOct      uint64          `json:"send_oct"`
	RecvCount    uint64          `json:"recv_cnt"`
	SendCount    uint64          `json:"send_cnt"`
	SendPendi    uint64          `json:"send_pend"`
	RecvOctDetails RateDetails   `json:"recv_oct_details"`
	SendOctDetails RateDetails   `json:"send_oct_details"`
}


func (c *Client) ListConnections() ([]ConnectionInfo, error) {
var err error
	req, err := NewHTTPRequest(c, "GET", "connections")
	if err != nil {
		return []ConnectionInfo{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return []ConnectionInfo{}, err
	}

	var rec []ConnectionInfo
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}


//
// Implementation
//

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

func NewHTTPRequest(client *Client, method string, path string) (*http.Request, error) {
	url := BuildPath(client.Endpoint, path)

	req, err := http.NewRequest(method, url, nil)
	req.SetBasicAuth(client.Username, client.Password)

	return req, err
}

func ExecuteHTTPRequest(client *Client, req *http.Request) (*http.Response, error) {
	httpc := &http.Client{}

	return httpc.Do(req)
}