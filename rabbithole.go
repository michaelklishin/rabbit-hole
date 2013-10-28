package rabbithole

import (
	"encoding/json"
	"net/http"
	"net/url"
	"bytes"
)

type Client struct {
	Endpoint, Host, Username, Password string
}

// TODO: this probably should be fixed in RabbitMQ management plugin
type OsPid string

type Properties map[string]interface{}
type Port int

type NameDescriptionEnabled struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Enabled     bool   `json:"enabled"`
}

type NameDescriptionVersion struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Version     string `json:"version"`
}

type NameAndVhost struct {
	Name  string `json:"name"`
	Vhost string `json:"vhost"`
}

type ExchangeType NameDescriptionEnabled

type RateDetails struct {
	Rate int `json:"rate"`
}

type MessageStats struct {
	Publish        int         `json:"publish"`
	PublishDetails RateDetails `json:"publish_details"`
}

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

type BrokerContext struct {
	Node        string `json:"node"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Port        Port   `json:"port"`
	Ignore      bool   `json:"ignore_in_use"`
}

//
// GET /api/overview
//

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
// GET /api/nodes
//

type AuthMechanism NameDescriptionEnabled
type ErlangApp NameDescriptionVersion

type NodeInfo struct {
	Name      string `json:"name"`
	NodeType  string `json:"type"`
	IsRunning bool   `json:"running"`
	OsPid     OsPid  `json:"os_pid"`

	FdUsed        int  `json:"fd_used"`
	FdTotal       int  `json:"fd_total"`
	SocketsUsed   int  `json:"sockets_used"`
	SocketsTotal  int  `json:"sockets_total"`
	MemUsed       int  `json:"mem_used"`
	MemLimit      int  `json:"mem_limit"`
	MemAlarm      bool `json:"mem_alarm"`
	DiskFreeAlarm bool `json:"disk_free_alarm"`

	ExchangeTypes  []ExchangeType  `json:"exchange_types"`
	AuthMechanisms []AuthMechanism `json:"auth_mechanisms"`
	ErlangApps     []ErlangApp     `json:"applications"`
	Contexts       []BrokerContext `json:"contexts"`
}

func (c *Client) ListNodes() ([]NodeInfo, error) {
	var err error
	req, err := NewGETRequest(c, "nodes")
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
	Name     string `json:"name"`
	Node     string `json:"node"`
	Channels int    `json:"channels"`
	State    string `json:"state"`
	Type     string `json:"type"`

	Port     Port `json:"port"`
	PeerPort Port `json:"peer_port"`

	Host     string `json:"host"`
	PeerHost string `json:"peer_host"`

	LastBlockedBy  string `json:"last_blocked_by"`
	LastBlockedAge string `json:"last_blocked_age"`

	UsesTLS          bool   `json:"ssl"`
	PeerCertSubject  string `json:"peer_cert_subject"`
	PeerCertValidity string `json:"peer_cert_validity"`
	PeerCertIssuer   string `json:"peer_cert_issuer"`

	SSLProtocol    string `json:"ssl_protocol"`
	SSLKeyExchange string `json:"ssl_key_exchange"`
	SSLCipher      string `json:"ssl_cipher"`
	SSLHash        string `json:"ssl_hash"`

	Protocol string `json:"protocol"`
	User     string `json:"user"`
	Vhost    string `json:"vhost"`

	Timeout  int `json:"timeout"`
	FrameMax int `json:"frame_max"`

	ClientProperties Properties `json:"client_properties"`

	RecvOct        uint64      `json:"recv_oct"`
	SendOct        uint64      `json:"send_oct"`
	RecvCount      uint64      `json:"recv_cnt"`
	SendCount      uint64      `json:"send_cnt"`
	SendPendi      uint64      `json:"send_pend"`
	RecvOctDetails RateDetails `json:"recv_oct_details"`
	SendOctDetails RateDetails `json:"send_oct_details"`
}

func (c *Client) ListConnections() ([]ConnectionInfo, error) {
	var err error
	req, err := NewGETRequest(c, "connections")
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
	Name     string `json:"name"`
	PeerPort Port   `json:"peer_port"`
	PeerHost string `json:"peer_host"`
}

type ChannelInfo struct {
	Number int    `json:"number"`
	Name   string `json:"name"`

	PrefetchCount int `json:"prefetch_count"`
	ConsumerCount int `json:"consumer_count"`

	UnacknowledgedMessageCount int `json:"messages_unacknowledged"`
	UnconfirmedMessageCount    int `json:"messages_unconfirmed"`
	UncommittedMessageCount    int `json:"messages_uncommitted"`
	UncommittedAckCount        int `json:"acks_uncommitted"`

	// TODO: custom deserializer to date/time?
	IdleSince string `json:"idle_since"`

	UsesPublisherConfirms bool `json:"confirm"`
	Transactional         bool `json:"transactional"`
	ClientFlowBlocked     bool `json:"client_flow_blocked"`

	User  string `json:"user"`
	Vhost string `json:"vhost"`
	Node  string `json:"node"`

	ConnectionDetails BriefConnectionDetails `json:"connection_details"`
}

func (c *Client) ListChannels() ([]ChannelInfo, error) {
	var err error
	req, err := NewGETRequest(c, "channels")
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
	req, err := NewGETRequest(c, "connections/"+url.QueryEscape(name))
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
	req, err := NewGETRequest(c, "channels/"+url.QueryEscape(name))
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

	MessageStats IngressEgressStats `json:"message_stats"`
}

type IngressEgressStats struct {
	PublishIn        int         `json:"publish_in"`
	PublishInDetails RateDetails `json:"publish_in_details"`

	PublishOut        int         `json:"publish_out"`
	PublishOutDetails RateDetails `json:"publish_out_details"`
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

//
// GET /api/queues
//

// [
//   {
//     "owner_pid_details": {
//       "name": "127.0.0.1:46928 -> 127.0.0.1:5672",
//       "peer_port": 46928,
//       "peer_host": "127.0.0.1"
//     },
//     "message_stats": {
//       "publish": 19830,
//       "publish_details": {
//         "rate": 5
//       }
//     },
//     "messages": 15,
//     "messages_details": {
//       "rate": 0
//     },
//     "messages_ready": 15,
//     "messages_ready_details": {
//       "rate": 0
//     },
//     "messages_unacknowledged": 0,
//     "messages_unacknowledged_details": {
//       "rate": 0
//     },
//     "policy": "",
//     "exclusive_consumer_tag": "",
//     "consumers": 0,
//     "memory": 143112,
//     "backing_queue_status": {
//       "q1": 0,
//       "q2": 0,
//       "delta": [
//         "delta",
//         "undefined",
//         0,
//         "undefined"
//       ],
//       "q3": 0,
//       "q4": 15,
//       "len": 15,
//       "pending_acks": 0,
//       "target_ram_count": "infinity",
//       "ram_msg_count": 15,
//       "ram_ack_count": 0,
//       "next_seq_id": 19830,
//       "persistent_count": 0,
//       "avg_ingress_rate": 4.9920127795527,
//       "avg_egress_rate": 4.9920127795527,
//       "avg_ack_ingress_rate": 0,
//       "avg_ack_egress_rate": 0
//     },
//     "status": "running",
//     "name": "amq.gen-QLEaT5Rn_ogbN3O8ZOQt3Q",
//     "vhost": "rabbit\/hole",
//     "durable": false,
//     "auto_delete": false,
//     "arguments": {
//       "x-message-ttl": 5000
//     },
//     "node": "rabbit@marzo"
//   }
// ]

type QueueInfo struct {
	Name       string                 `json:"name"`
	Vhost      string                 `json:"vhost"`
	Durable    bool                   `json:"durable"`
	AutoDelete bool                   `json:"auto_delete"`
	Arguments  map[string]interface{} `json:"arguments"`

	Node   string `json:"node"`
	Status string `json:"status"`

	Memory               int64  `json:"memory"`
	Consumers            int    `json:"consumers"`
	ExclusiveConsumerTag string `json:"exclusive_consumer_tag"`

	Policy string `json:"policy"`

	Messages        int         `json:"messages"`
	MessagesDetails RateDetails `json:"messages_details"`

	MessagesReady        int         `json:"messages_ready"`
	MessagesReadyDetails RateDetails `json:"messages_ready_details"`

	MessagesUnacknowledged        int         `json:"messages_unacknowledged"`
	MessagesUnacknowledgedDetails RateDetails `json:"messages_unacknowledged_details"`

	MessageStats MessageStats `json:"message_stats"`

	OwnerPidDetails OwnerPidDetails `json:"owner_pid_details"`

	BackingQueueStatus BackingQueueStatus `json:"backing_queue_status"`
}

type BackingQueueStatus struct {
	Q1                    int     `json:"q1"`
	Q2                    int     `json:"q2"`
	Q3                    int     `json:"q3"`
	Q4                    int     `json:"q4"`
	Length                int64   `json:"len"`
	PendingAcks           int64   `json:"pending_acks"`
	RAMMessageCount       int64   `json:"ram_msg_count"`
	RAMAckCount           int64   `json:"ram_ack_count"`
	PersistentCount       int64   `json:"persistent_count"`
	AverageIngressRate    float64 `json:"avg_ingress_rate"`
	AverageEgressRate     float64 `json:"avg_egress_rate"`
	AverageAckIngressRate float32 `json:"avg_ack_ingress_rate"`
	AverageAckEgressRate  float32 `json:"avg_ack_egress_rate"`
}

type OwnerPidDetails struct {
	Name     string `json:"name"`
	PeerPort Port   `json:"peer_port"`
	PeerHost string `json:"peer_host"`
}

func (c *Client) ListQueues() ([]QueueInfo, error) {
	var err error
	req, err := NewGETRequest(c, "queues")
	if err != nil {
		return []QueueInfo{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return []QueueInfo{}, err
	}

	var rec []QueueInfo
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}

//
// GET /api/queues/{vhost}
//

func (c *Client) ListQueuesIn(vhost string) ([]QueueInfo, error) {
	var err error
	req, err := NewGETRequest(c, "queues/"+url.QueryEscape(vhost))
	if err != nil {
		return []QueueInfo{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return []QueueInfo{}, err
	}

	var rec []QueueInfo
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}

//
// GET /api/queues/{vhost}/{name}
//

type DetailedQueueInfo QueueInfo

func (c *Client) GetQueue(vhost, queue string) (DetailedQueueInfo, error) {
	var err error
	req, err := NewGETRequest(c, "queues/"+url.QueryEscape(vhost)+"/"+queue)
	if err != nil {
		return DetailedQueueInfo{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return DetailedQueueInfo{}, err
	}

	var rec DetailedQueueInfo
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}

//
// GET /api/users
//

// Example response:
// [{"name":"guest","password_hash":"8LYTIFbVUwi8HuV/dGlp2BYsD1I=","tags":"administrator"}]

type UserInfo struct {
	Name         string `json:"name"`
	PasswordHash string `json:"password_hash"`
	Tags         string `json:"tags"`

	// *never* returned by RabbitMQ. Set by the client
	// to create/update a user. MK.
	Password     string `json:"password"`
}

func (c *Client) ListUsers() ([]UserInfo, error) {
	var err error
	req, err := NewGETRequest(c, "users/")
	if err != nil {
		return []UserInfo{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return []UserInfo{}, err
	}

	var rec []UserInfo
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}

//
// GET /api/users/{name}
//

func (c *Client) GetUser(username string) (UserInfo, error) {
	var err error
	req, err := NewGETRequest(c, "users/"+url.QueryEscape(username))
	if err != nil {
		return UserInfo{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return UserInfo{}, err
	}

	var rec UserInfo
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}


//
// PUT /api/users/{name}
//

func (c *Client) PutUser(username string, info UserInfo) (*http.Response, error) {
	var err error

	body, err := json.Marshal(info)
	if err != nil {
		return &http.Response{}, err
	}

	req, err := NewHTTPRequestWithBody(c, "PUT", "users/"+url.QueryEscape(username), body)
	if err != nil {
		return &http.Response{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return &http.Response{}, err
	}

	return res, nil
}


//
// Implementation
//

func NewClient(uri string, username string, password string) (*Client, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return &Client{}, err
	}

	me := &Client{
		Endpoint: uri,
		Host:     u.Host,
		Username: username,
		Password: password}

	return me, nil
}

func NewGETRequest(client *Client, path string) (*http.Request, error) {
	s := client.Endpoint + "/api/" + path

	req, err := http.NewRequest("GET", s, nil)
	req.SetBasicAuth(client.Username, client.Password)
	// set Opaque to preserve percent-encoded path. MK.
	req.URL.Opaque = "//" + client.Host + "/api/" + path

	return req, err
}

func NewHTTPRequestWithBody(client *Client, method string, path string, body []byte) (*http.Request, error) {
	s := client.Endpoint + "/api/" + path

	req, err := http.NewRequest(method, s, bytes.NewReader(body))
	req.SetBasicAuth(client.Username, client.Password)
	// set Opaque to preserve percent-encoded path. MK.
	req.URL.Opaque = "//" + client.Host + "/api/" + path

	req.Header.Add("Content-Type", "application/json")

	return req, err
}

func ExecuteHTTPRequest(client *Client, req *http.Request) (*http.Response, error) {
	httpc := &http.Client{}

	return httpc.Do(req)
}
