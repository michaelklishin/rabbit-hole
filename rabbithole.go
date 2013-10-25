package rabbithole

import (
	"net/http"
	"encoding/json"
)

type Client struct {
	Endpoint, Username, Password string
}

// TODO: this probably should be fixed in RabbitMQ management plugin
type OsPid   string

type Properties map[string]interface{}
type Port       int


type NameDescriptionEnabled struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

type NameDescriptionVersion struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Version     string `json:"version"`
}

type ExchangeType NameDescriptionEnabled

type RateDetails struct {
	Rate        int   `json:"rate"`
}

type MessageStats struct {
	Publish        int      `json:"publish"`
	PublishDetails RateDetails `json:"publish_details"`
}


type QueueTotals struct {
	Messages        int      `json:"messages"`
	MessagesDetails RateDetails `json:"messages_details"`

	MessagesReady        int      `json:"messages_ready"`
	MessagesReadyDetails RateDetails `json:"messages_ready_details"`

	MessagesUnacknowledged        int      `json:"messages_unacknowledged"`
	MessagesUnacknowledgedDetails RateDetails `json:"messages_unacknowledged_details"`
}

type ObjectTotals struct {
        Consumers   int   `json:"consumers"`
	Queues      int   `json:"queues"`
	Exchanges   int   `json:"exchanges"`
	Connections int   `json:"connections"`
	Channels    int   `json:"channels"`
}

type Listener struct {
	Node              string       `json:"node"`
	Protocol          string       `json:"protocol"`
	IpAddress         string       `json:"ip_address"`
	Port              Port         `json:"port"`
}

type BrokerContext struct {
	Node              string       `json:"node"`
	Description       string         `json:"description"`
	Path              string         `json:"path"`
	Port              Port           `json:"port"`
	Ignore            bool           `json:"ignore_in_use"`
}



//
// GET /api/overview
//

type Overview struct {
	ManagementVersion string         `json:"management_version"`
	StatisticsLevel   string          `json:"statistics_level"`
	RabbitMQVersion   string         `json:"rabbitmq_version"`
	ErlangVersion     string         `json:"erlang_version"`
	FullErlangVersion string         `json:"erlang_full_version"`
	ExchangeTypes     []ExchangeType  `json:"exchange_types"`
	MessageStats      MessageStats    `json:"message_stats"`
	QueueTotals       QueueTotals     `json:"queue_totals"`
	ObjectTotals      ObjectTotals    `json:"object_totals"`
	Node              string          `json:"node"`
	StatisticsDBNode  string          `json:"statistics_db_node"`
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
	Name      string   `json:"name"`
	NodeType  string   `json:"type"`
	IsRunning bool     `json:"running"`
	OsPid     OsPid    `json:"os_pid"`

	FdUsed       int  `json:"fd_used"`
	FdTotal      int  `json:"fd_total"`
	SocketsUsed  int  `json:"sockets_used"`
	SocketsTotal int  `json:"sockets_total"`
	MemUsed      int  `json:"mem_used"`
	MemLimit     int  `json:"mem_limit"`

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
        Name           string              `json:"name"`
	Node           string              `json:"node"`
	Channels       int              `json:"channels"`
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

	Protocol         string            `json:"protocol"`
	User             string            `json:"user"`
	Vhost            string            `json:"vhost"`

	Timeout           int           `json:"timeout"`
	FrameMax          int           `json:"frame_max"`

	ClientProperties  Properties       `json:"client_properties"`


	RecvOct      int          `json:"recv_oct"`
	SendOct      int          `json:"send_oct"`
	RecvCount    int          `json:"recv_cnt"`
	SendCount    int          `json:"send_cnt"`
	SendPendi    int          `json:"send_pend"`
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
// GET /api/channels
//

type BriefConnectionDetails struct {
        Name           string               `json:"name"`
	PeerPort       Port                `json:"peer_port"`
	PeerHost       string              `json:"peer_host"`

}

type ChannelInfo struct {
	Number         int               `json:"number"`
	Name           string               `json:"name"`

        PrefetchCount  int               `json:"prefetch_count"`
        ConsumerCount  int               `json:"consumer_count"`

        UnacknowledgedMessageCount  int  `json:"messages_unacknowledged"`
        UnconfirmedMessageCount     int  `json:"messages_unconfirmed"`
        UncommittedMessageCount     int  `json:"messages_uncommitted"`
        UncommittedAckCount         int  `json:"acks_uncommitted"`

	// TODO: custom deserializer to date/time?
	IdleSince      string               `json:"idle_since"`

	UsesPublisherConfirms bool          `json:"confirm"`
	Transactional         bool          `json:"transactional"`
	ClientFlowBlocked     bool          `json:"client_flow_blocked"`

        User           string               `json:"user"`
	Vhost          string               `json:"vhost"`
	Node           string               `json:"node"`

	ConnectionDetails BriefConnectionDetails `json:"connection_details"`
}


func (c *Client) ListChannels() ([]ChannelInfo, error) {
	var err error
	req, err := NewHTTPRequest(c, "GET", "channels")
	if err != nil {
		return []ChannelInfo{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return []ChannelInfo{}, err
	}

	var rec []ChannelInfo
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}


//
// GET /api/connections/{name}
//

func (c *Client) GetConnection(name string) (ConnectionInfo, error) {
	var err error
	req, err := NewHTTPRequest(c, "GET", "connections/" + name)
	if err != nil {
		return ConnectionInfo{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return ConnectionInfo{}, err
	}

	var rec ConnectionInfo
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}


//
// GET /api/channels/{name}
//

func (c *Client) GetChannel(name string) (ChannelInfo, error) {
	var err error
	req, err := NewHTTPRequest(c, "GET", "channels/" + name)
	if err != nil {
		return ChannelInfo{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return ChannelInfo{}, err
	}

	var rec ChannelInfo
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}


//
// GET /api/exchanges
//

type ExchangeInfo struct {
	Name       string                 `json:"name"`
	Vhost      string                 `json:"vhost"`
	Type       string                 `json:"type"`
	Durable    bool                   `json:"durable"`
	AutoDelete bool                   `json:"auto_delete"`
	Internal   bool                   `json:"internal"`
	Arguments  map[string]interface{} `json:"arguments"`

	MessageStats IngressStats         `json:"message_stats"`
}

type IngressStats struct {
	PublishIn        int      `json:"publish_in"`
	PublishInDetails RateDetails `json:"publish_in_details"`
}


func (c *Client) ListExchanges() ([]ExchangeInfo, error) {
	var err error
	req, err := NewHTTPRequest(c, "GET", "exchanges")
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
	// TODO: percent encoding, e.g. / => %2F. MK.
	req, err := NewHTTPRequest(c, "GET", "exchanges/" + vhost)
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