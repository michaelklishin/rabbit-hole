package rabbithole

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
)

// OperatorPolicy represents an operator policy.
type OperatorPolicy struct {
	// Virtual host this policy is in.
	Vhost string `json:"vhost"`
	// Regular expression pattern used to match queues.
	Pattern string `json:"pattern"`
	// What this policy applies to: "queues".
	ApplyTo  string `json:"apply-to"`
	Name     string `json:"name"`
	Priority int    `json:"priority"`
	// Additional arguments added to matching queues.
	Definition PolicyDefinition `json:"definition"`
}

// HasCMQKeys returns true if this policy's definition contains CMQ keys.
func (p OperatorPolicy) HasCMQKeys() bool {
	return p.Definition.HasCMQKeys()
}

// DoesMatchName returns true if this policy would apply to a queue with the given name and target.
func (p OperatorPolicy) DoesMatchName(vhost, name string, target PolicyTarget) bool {
	if p.Vhost != vhost {
		return false
	}
	if !targetsMatch(PolicyTarget(p.ApplyTo), target) {
		return false
	}
	matched, err := regexp.MatchString(p.Pattern, name)
	return err == nil && matched
}

//
// GET /api/operator-policies
//

// ListOperatorPolicies returns all operator policies (across all virtual hosts).
func (c *Client) ListOperatorPolicies() (rec []OperatorPolicy, err error) {
	req, err := newGETRequest(c, "operator-policies")
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

//
// GET /api/operator-policies/{vhost}
//

// ListOperatorPoliciesIn returns operator policies in a specific virtual host.
func (c *Client) ListOperatorPoliciesIn(vhost string) (rec []OperatorPolicy, err error) {
	req, err := newGETRequest(c, "operator-policies/"+url.PathEscape(vhost))
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

//
// GET /api/operator-policies/{vhost}/{name}
//

// GetOperatorPolicy returns an operator policy by name.
func (c *Client) GetOperatorPolicy(vhost, name string) (rec *OperatorPolicy, err error) {
	req, err := newGETRequest(c, "operator-policies/"+url.PathEscape(vhost)+"/"+url.PathEscape(name))
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

//
// PUT /api/operator-policies/{vhost}/{name}
//

// PutOperatorPolicy creates or updates an operator policy.
func (c *Client) PutOperatorPolicy(vhost string, name string, operatorPolicy OperatorPolicy) (res *http.Response, err error) {
	body, err := json.Marshal(operatorPolicy)
	if err != nil {
		return nil, err
	}

	req, err := newRequestWithBody(c, "PUT", "operator-policies/"+url.PathEscape(vhost)+"/"+url.PathEscape(name), body)
	if err != nil {
		return nil, err
	}

	if res, err = executeRequest(c, req); err != nil {
		return nil, err
	}

	return res, nil
}

//
// DELETE /api/operator-policies/{vhost}/{name}
//

// DeleteOperatorPolicy deletes an operator policy.
func (c *Client) DeleteOperatorPolicy(vhost, name string) (res *http.Response, err error) {
	req, err := newRequestWithBody(c, "DELETE", "operator-policies/"+url.PathEscape(vhost)+"/"+url.PathEscape(name), nil)
	if err != nil {
		return nil, err
	}

	if res, err = executeRequest(c, req); err != nil {
		return nil, err
	}

	return res, nil
}

// ListOperatorPoliciesForTarget returns operator policies in a virtual host filtered by their apply-to target.
func (c *Client) ListOperatorPoliciesForTarget(vhost string, target PolicyTarget) (rec []OperatorPolicy, err error) {
	policies, err := c.ListOperatorPoliciesIn(vhost)
	if err != nil {
		return nil, err
	}

	for _, p := range policies {
		if PolicyTarget(p.ApplyTo) == target {
			rec = append(rec, p)
		}
	}

	return rec, nil
}

// ListMatchingOperatorPolicies returns operator policies that would match a given entity name and target.
func (c *Client) ListMatchingOperatorPolicies(vhost, name string, target PolicyTarget) (rec []OperatorPolicy, err error) {
	policies, err := c.ListOperatorPoliciesIn(vhost)
	if err != nil {
		return nil, err
	}

	for _, p := range policies {
		if p.DoesMatchName(vhost, name, target) {
			rec = append(rec, p)
		}
	}

	return rec, nil
}

// DeleteOperatorPoliciesIn deletes multiple operator policies in a virtual host by name.
func (c *Client) DeleteOperatorPoliciesIn(vhost string, names []string) error {
	for _, name := range names {
		_, err := c.DeleteOperatorPolicy(vhost, name)
		if err != nil {
			return err
		}
	}
	return nil
}
