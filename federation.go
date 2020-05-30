package rabbithole

import (
	"net/http"
)

// FederationDefinition represents settings
// that will be used by federation links.
type FederationDefinition struct {
	Uri            string `json:"uri"`
	Expires        int    `json:"expires,omitempty"`
	MessageTTL     int32  `json:"message-ttl"`
	MaxHops        int    `json:"max-hops"`
	PrefetchCount  int    `json:"prefetch-count"`
	ReconnectDelay int    `json:"reconnect-delay"`
	AckMode        string `json:"ack-mode,omitempty"`
	TrustUserId    bool   `json:"trust-user-id"`
	Exchange       string `json:"exchange"`
	Queue          string `json:"queue"`
}

// FederationUpstream represents a configured federation upstream.
type FederationUpstream struct {
	Name       string               `json:"name"`
	Vhost      string               `json:"vhost"`
	Component  string               `json:"component"`
	Definition FederationDefinition `json:"value"`
}

// FederationUpstreamComponent is the name of the runtime parameter component
// used by federation upstreams.
const FederationUpstreamComponent string = "federation-upstream"

//
// GET /api/parameters/federation-upstream
//

// ListFederationUpstreams returns a list of all federation upstreams.
func (c *Client) ListFederationUpstreams() (ups []FederationUpstream, err error) {
	params, err := c.ListRuntimeParametersFor(FederationUpstreamComponent)
	if err != nil {
		return nil, err
	}

	for _, p := range params {
		up := paramToUpstream(&p)
		ups = append(ups, *up)
	}
	return ups, nil
}

//
// GET /api/parameters/federation-upstream/{vhost}
//

// ListFederationUpstreamsIn returns a list of all federation upstreams in a vhost.
func (c *Client) ListFederationUpstreamsIn(vhost string) (ups []FederationUpstream, err error) {
	params, err := c.ListRuntimeParametersIn(FederationUpstreamComponent, vhost)
	if err != nil {
		return nil, err
	}

	for _, p := range params {
		up := paramToUpstream(&p)
		ups = append(ups, *up)
	}
	return ups, nil
}

//
// GET /api/parameters/federation-upstream/{vhost}/{upstream}
//

// GetFederationUpstream returns information about a federation upstream.
func (c *Client) GetFederationUpstream(vhost, name string) (up *FederationUpstream, err error) {
	p, err := c.GetRuntimeParameter(FederationUpstreamComponent, vhost, name)
	if err != nil {
		return nil, err
	}
	return paramToUpstream(p), nil
}

//
// PUT /api/parameters/federation-upstream/{vhost}/{upstream}
//

// PutFederationUpstream creates or updates a federation upstream configuration.
func (c *Client) PutFederationUpstream(vhost, name string, def FederationDefinition) (res *http.Response, err error) {
	return c.PutRuntimeParameter(FederationUpstreamComponent, vhost, name, def)
}

//
// DELETE /api/parameters/federation-upstream/{vhost}/{name}
//

// DeleteFederationUpstream removes a federation upstream.
func (c *Client) DeleteFederationUpstream(vhost, name string) (res *http.Response, err error) {
	return c.DeleteRuntimeParameter(FederationUpstreamComponent, vhost, name)
}

// paramToUpstream maps from a RuntimeParameter structure to a FederationUpstream structure.
func paramToUpstream(p *RuntimeParameter) (up *FederationUpstream) {
	up = &FederationUpstream{
		Name:      p.Name,
		Vhost:     p.Vhost,
		Component: p.Component,
	}

	m, ok := p.Value.(map[string]interface{})
	if !ok {
		return up
	}

	def := FederationDefinition{}

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
	return up
}
