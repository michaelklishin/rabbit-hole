package rabbithole

import (
	"encoding/json"
	"net/url"
)

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

//
// GET /api/connections
//

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
