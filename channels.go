package rabbithole

import "net/url"

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

func (c *Client) ListChannels() (rec []ChannelInfo, err error) {
	req, err := newGETRequest(c, "channels")
	if err != nil {
		return []ChannelInfo{}, err
	}

	if err = executeAndParseRequest(req, &rec); err != nil {
		return []ChannelInfo{}, err
	}

	return rec, nil
}

//
// GET /api/channels/{name}
//

func (c *Client) GetChannel(name string) (rec *ChannelInfo, err error) {
	req, err := newGETRequest(c, "channels/"+url.QueryEscape(name))
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}
