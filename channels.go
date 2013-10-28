package rabbithole

import (
	"encoding/json"
	"net/url"
)

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

//
// GET /api/channels
//

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
