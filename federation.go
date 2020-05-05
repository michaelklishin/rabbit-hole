package rabbithole

import (
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
func (c *Client) GetFederationUpstream(vhost, name string) (up *FederationUpstream, err error) {
	param, err := c.GetRuntimeParameter("federation-upstream", vhost, name)
	if err != nil {
		return nil, err
	}

	// TODO: extract to function when migrating List* methods.
	up = &FederationUpstream{
		Name:      param.Name,
		Vhost:     param.Vhost,
		Component: param.Component,
	}

	def := FederationDefinition{}
	m := param.Value.(map[string]interface{})

	if v, ok := m["uri"].(string); ok {
		def.Uri = v
	}

	if v, ok := m["expires"].(float64); ok {
		def.Expires = int(v)
	}

	if v, ok := m["message-ttl"].(float64); ok {
		def.MessageTTL = int32(v)
	}

	if v, ok := m["max-hops"].(float64); ok {
		def.MaxHops = int(v)
	}

	if v, ok := m["prefetch-count"].(float64); ok {
		def.PrefetchCount = int(v)
	}

	if v, ok := m["reconnect-delay"].(float64); ok {
		def.ReconnectDelay = int(v)
	}

	if v, ok := m["ack-mode"].(string); ok {
		def.AckMode = v
	}

	if v, ok := m["trust-user-id"].(bool); ok {
		def.TrustUserId = v
	}

	if v, ok := m["exchange"].(string); ok {
		def.Exchange = v
	}

	if v, ok := m["queue"].(string); ok {
		def.Queue = v
	}

	up.Definition = def

	return up, nil
}

//
// PUT /api/parameters/federation-upstream/{vhost}/{upstream}
//

// Updates a federation upstream
func (c *Client) PutFederationUpstream(vhost string, name string, def FederationDefinition) (res *http.Response, err error) {
	return c.PutRuntimeParameter("federation-upstream", vhost, name, def)
}

//
// DELETE /api/parameters/federation-upstream/{vhost}/{name}
//

// Deletes a federation upstream.
func (c *Client) DeleteFederationUpstream(vhost, name string) (res *http.Response, err error) {
	return c.DeleteRuntimeParameter("federation-upstream", vhost, name)
}
