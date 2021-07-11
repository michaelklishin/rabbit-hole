package rabbithole

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// VhostLimitsValue are properties used to modify virtual hosts limits.
type VhostLimitsValue struct {
	// Maximum number of connections
	MaxConnections int `json:"max-connections"`
}

type VhostLimitsInfo struct {
	Vhost string           `json:"vhost"`
	Value VhostLimitsValue `json:"value"`
}

type VhostLimitsMaxConnections struct {
	Value int `json:"value"`
}

func (c *Client) GetVhostLimits(vhostname string) (rec []VhostLimitsInfo, err error) {
	req, err := newGETRequest(c, "vhost-limits/"+url.PathEscape(vhostname))
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

func (c *Client) PutVhostLimitsMaxConnections(vhostname string, maxConnections int) (res *http.Response, err error) {
	body, err := json.Marshal(VhostLimitsMaxConnections{Value: maxConnections})
	if err != nil {
		return nil, err
	}

	req, err := newRequestWithBody(c, "PUT", "vhost-limits/"+url.PathEscape(vhostname)+"/max-connections", body)
	if err != nil {
		return nil, err
	}

	if res, err = executeRequest(c, req); err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) DeleteVhostLimitsMaxConnections(vhostname string) (res *http.Response, err error) {
	req, err := newRequestWithBody(c, "DELETE", "vhost-limits/"+url.PathEscape(vhostname)+"/max-connections", nil)
	if err != nil {
		return nil, err
	}

	if res, err = executeRequest(c, req); err != nil {
		return nil, err
	}

	return res, nil
}
