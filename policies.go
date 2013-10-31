package rabbithole

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type PolicyDefinition map[string]interface{}

type NodeNames []string

type Policy struct {
	Vhost      string           `json:"vhost"`
	Pattern    string           `json:"pattern"`
	ApplyTo    string           `json:"apply-to"`
	Name       string           `json:"name"`
	Priority   int              `json:"priority"`
	Definition PolicyDefinition `json:"definition"`
}

//
// GET /api/policies/{name}
//

func (c *Client) GetPolicy(vhost, name string) (rec *Policy, err error) {
	req, err := newGETRequest(c, "policies/"+url.QueryEscape(vhost)+"/"+url.QueryEscape(name))
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

//
// PUT /api/policies/{vhost}/{name}
//

func (c *Client) PutPolicy(vhost string, name string, policy Policy) (res *http.Response, err error) {
	body, err := json.Marshal(policy)
	if err != nil {
		return nil, err
	}

	req, err := newRequestWithBody(c, "PUT", "policies/"+url.QueryEscape(vhost)+"/"+url.QueryEscape(name), body)
	if err != nil {
		return nil, err
	}

	res, err = executeRequest(c, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

//
// DELETE /api/policies/{vhost}/{name}
//

func (c *Client) DeletePolicy(vhost, name string) (res *http.Response, err error) {
	req, err := newRequestWithBody(c, "DELETE", "policies/"+url.QueryEscape(vhost)+"/"+url.QueryEscape(name), nil)
	if err != nil {
		return nil, err
	}

	res, err = executeRequest(c, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
