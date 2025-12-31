package rabbithole

import (
	"net/url"
)

// ListVhostChannels returns channels in a virtual host.
func (c *Client) ListVhostChannels(vhostname string) (rec []ChannelInfo, err error) {
	req, err := newGETRequest(c, "vhosts/"+url.PathEscape(vhostname)+"/channels")
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}
