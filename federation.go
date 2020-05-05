package rabbithole

import (
	"encoding/json"
	"net/http"
	"net/url"
)

// Federation definition: additional arguments
// added to the entities (queues, exchanges or both)
// that match a policy.
type FederationDefinition struct {
	Uri            string `json:"uri"`
	Expires        int    `json:"expires"`
	MessageTTL     int32  `json:"message-ttl"`
	MaxHops        int    `json:"max-hops"`
	PrefetchCount  int    `json:"prefetch-count"`
	ReconnectDelay int    `json:"reconnect-delay"`
	AckMode        string `json:"ack-mode,omitempty"`
	TrustUserId    bool   `json:"trust-user-id"`
	Exchange       string `json:"exchange"`
	Queue          string `json:"queue"`
}

// Represents a configured Federation upstream.
type FederationUpstream struct {
	Name       string               `json:"name"`
	Vhost      string               `json:"vhost"`
	Component  string               `json:"component"`
	Definition FederationDefinition `json:"value"`
}

// FederationDefinitionDTO provides a data transfer object for a FederationDefinition.
type FederationDefinitionDTO struct {
	Definition FederationDefinition `json:"value"`
}

//
// GET /api/parameters/federation-upstream
//

// ListFederationUpstreams returns all federation upstreams
func (c *Client) ListFederationUpstreams() (rec []FederationUpstream, err error) {
	req, err := newGETRequest(c, "parameters/federation-upstream")
	if err != nil {
		return []FederationUpstream{}, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return []FederationUpstream{}, err
	}

	return rec, nil
}

//
// GET /api/parameters/federation-upstream/{vhost}
//

// ListFederationUpstreamsIn returns all federation upstreams in a vhost
func (c *Client) ListFederationUpstreamsIn(vhost string) (rec []FederationUpstream, err error) {
	req, err := newGETRequest(c, "parameters/federation-upstream/"+url.PathEscape(vhost))
	if err != nil {
		return []FederationUpstream{}, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return []FederationUpstream{}, err
	}

	return rec, nil
}

//
// GET /api/parameters/federation-upstream/{vhost}/{upstream}
//

// GetFederationUpstream returns a federation upstream
func (c *Client) GetFederationUpstream(vhost, upstreamName string) (rec *FederationUpstream, err error) {
	req, err := newGETRequest(c, "parameters/federation-upstream/"+url.PathEscape(vhost)+"/"+url.PathEscape(upstreamName))

	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

//
// PUT /api/parameters/federation-upstream/{vhost}/{upstream}
//

// Updates a federation upstream
func (c *Client) PutFederationUpstream(vhost string, upstreamName string, fDef FederationDefinition) (res *http.Response, err error) {
	fedDTO := FederationDefinitionDTO{
		Definition: fDef,
	}

	body, err := json.Marshal(fedDTO)
	if err != nil {
		return nil, err
	}

	req, err := newRequestWithBody(c, "PUT", "parameters/federation-upstream/"+url.PathEscape(vhost)+"/"+url.PathEscape(upstreamName), body)
	if err != nil {
		return nil, err
	}

	if res, err = executeRequest(c, req); err != nil {
		return nil, err
	}

	return res, nil
}

func (c *Client) PutFederationUpstreamV2(vhost string, name string, def FederationDefinition) (res *http.Response, err error) {
	return c.PutRuntimeParameter("federation-upstream", vhost, name, def)
}

//
// DELETE /api/parameters/federation-upstream/{vhost}/{name}
//

// Deletes a federation upstream.
func (c *Client) DeleteFederationUpstream(vhost, upstreamName string) (res *http.Response, err error) {
	req, err := newRequestWithBody(c, "DELETE", "parameters/federation-upstream/"+url.PathEscape(vhost)+"/"+url.PathEscape(upstreamName), nil)
	if err != nil {
		return nil, err
	}

	if res, err = executeRequest(c, req); err != nil {
		return nil, err
	}

	return res, nil
}
