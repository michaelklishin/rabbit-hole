package rabbithole

import (
	"net/url"
)

// ShovelStatus contains the configuration of a shovel
type ShovelStatus struct {
	// Shovel name
	Name string `json:"name"`
	// Virtual host this shovel belongs to
	Vhost string `json:"vhost"`
	// Type of this shovel
	Type string `json:"state"`
	// State of this shovel
	State string `json:"state"`
	// Timestamp of this shovel
	Timestamp string `json:"timestamp"`
}

//
// GET /api/shovels/{vhost}
//

// ListShovelStatus returns status of all shovels in a vhost
func (c *Client) ListShovelStatus(vhost string) (rec []ShovelStatus, err error) {
	req, err := newGETRequest(c, "shovels/" + url.PathEscape(vhost))
	if err != nil {
		return []ShovelStatus{}, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return []ShovelStatus{}, err
	}

	return rec, nil
}
