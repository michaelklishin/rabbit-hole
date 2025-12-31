package rabbithole

import (
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
)

// PolicyDefinition is a map of additional arguments
// added to the entities (queues, exchanges or both)
// that match a policy.
type PolicyDefinition map[string]interface{}

// CMQPolicyKeys are the policy keys used by classic mirrored queues.
// These are deprecated and should not be used with quorum queues or streams.
var CMQPolicyKeys = []string{
	"ha-mode",
	"ha-params",
	"ha-sync-mode",
	"ha-sync-batch-size",
	"ha-promote-on-shutdown",
	"ha-promote-on-failure",
}

// QuorumQueueIncompatibleKeys are policy keys that are not compatible with quorum queues.
var QuorumQueueIncompatibleKeys = []string{
	"ha-mode",
	"ha-params",
	"ha-sync-mode",
	"ha-sync-batch-size",
	"ha-promote-on-shutdown",
	"ha-promote-on-failure",
	"max-priority",
	"queue-master-locator",
	"queue-mode",
}

// HasCMQKeys reports whether any classic mirrored queue keys are present.
func (pd PolicyDefinition) HasCMQKeys() bool {
	for _, key := range CMQPolicyKeys {
		if _, ok := pd[key]; ok {
			return true
		}
	}
	return false
}

// WithoutCMQKeys returns a copy of the policy definition without
// classic mirrored queue keys.
func (pd PolicyDefinition) WithoutCMQKeys() PolicyDefinition {
	return pd.WithoutKeys(CMQPolicyKeys)
}

// WithoutQuorumQueueIncompatibleKeys returns a copy of the policy definition
// without keys that are incompatible with quorum queues.
func (pd PolicyDefinition) WithoutQuorumQueueIncompatibleKeys() PolicyDefinition {
	return pd.WithoutKeys(QuorumQueueIncompatibleKeys)
}

// WithoutKeys returns a copy of the policy definition without the specified keys.
func (pd PolicyDefinition) WithoutKeys(keys []string) PolicyDefinition {
	result := make(PolicyDefinition)
	excluded := make(map[string]bool, len(keys))
	for _, key := range keys {
		excluded[key] = true
	}
	for k, v := range pd {
		if !excluded[k] {
			result[k] = v
		}
	}
	return result
}

// PolicyTarget represents what type of entities a policy applies to.
type PolicyTarget string

const (
	// PolicyTargetQueues applies to all queue types.
	PolicyTargetQueues PolicyTarget = "queues"
	// PolicyTargetClassicQueues applies only to classic queues.
	PolicyTargetClassicQueues PolicyTarget = "classic_queues"
	// PolicyTargetQuorumQueues applies only to quorum queues.
	PolicyTargetQuorumQueues PolicyTarget = "quorum_queues"
	// PolicyTargetStreams applies only to streams.
	PolicyTargetStreams PolicyTarget = "streams"
	// PolicyTargetExchanges applies to exchanges.
	PolicyTargetExchanges PolicyTarget = "exchanges"
	// PolicyTargetAll applies to both queues and exchanges.
	PolicyTargetAll PolicyTarget = "all"
)

func (t PolicyTarget) MatchesQueues() bool {
	return t == PolicyTargetQueues || t == PolicyTargetClassicQueues ||
		t == PolicyTargetQuorumQueues || t == PolicyTargetStreams || t == PolicyTargetAll
}

func (t PolicyTarget) MatchesExchanges() bool {
	return t == PolicyTargetExchanges || t == PolicyTargetAll
}

func (t PolicyTarget) MatchesClassicQueues() bool {
	return t == PolicyTargetQueues || t == PolicyTargetClassicQueues || t == PolicyTargetAll
}

func (t PolicyTarget) MatchesQuorumQueues() bool {
	return t == PolicyTargetQueues || t == PolicyTargetQuorumQueues || t == PolicyTargetAll
}

func (t PolicyTarget) MatchesStreams() bool {
	return t == PolicyTargetQueues || t == PolicyTargetStreams || t == PolicyTargetAll
}

// Policy represents a configured policy.
type Policy struct {
	// Virtual host this policy is in.
	Vhost string `json:"vhost"`
	// Regular expression pattern used to match queues and exchanges.
	Pattern string `json:"pattern"`
	// What this policy applies to: "queues", "exchanges", etc.
	ApplyTo  string `json:"apply-to"`
	Name     string `json:"name"`
	Priority int    `json:"priority"`
	// Additional arguments added to matching entities.
	Definition PolicyDefinition `json:"definition"`
}

// HasCMQKeys returns true if this policy's definition contains CMQ keys.
func (p Policy) HasCMQKeys() bool {
	return p.Definition.HasCMQKeys()
}

// DoesMatchName returns true if this policy would apply to an entity with the given name and target.
func (p Policy) DoesMatchName(vhost, name string, target PolicyTarget) bool {
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
// GET /api/policies
//

// ListPolicies returns all policies (across all virtual hosts).
func (c *Client) ListPolicies() (rec []Policy, err error) {
	req, err := newGETRequest(c, "policies")
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

//
// GET /api/policies/{vhost}
//

// ListPoliciesIn returns policies in a specific virtual host.
func (c *Client) ListPoliciesIn(vhost string) (rec []Policy, err error) {
	req, err := newGETRequest(c, "policies/"+url.PathEscape(vhost))
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

//
// GET /api/policies/{vhost}/{name}
//

// GetPolicy returns a policy by name.
func (c *Client) GetPolicy(vhost, name string) (rec *Policy, err error) {
	req, err := newGETRequest(c, "policies/"+url.PathEscape(vhost)+"/"+url.PathEscape(name))
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

//
// PUT /api/policies/{vhost}/{name}
//

// PutPolicy creates or updates a policy.
func (c *Client) PutPolicy(vhost string, name string, policy Policy) (res *http.Response, err error) {
	body, err := json.Marshal(policy)
	if err != nil {
		return nil, err
	}

	req, err := newRequestWithBody(c, "PUT", "policies/"+url.PathEscape(vhost)+"/"+url.PathEscape(name), body)
	if err != nil {
		return nil, err
	}

	if res, err = executeRequest(c, req); err != nil {
		return nil, err
	}

	return res, nil
}

//
// DELETE /api/policies/{vhost}/{name}
//

// DeletePolicy deletes a policy.
func (c *Client) DeletePolicy(vhost, name string) (res *http.Response, err error) {
	req, err := newRequestWithBody(c, "DELETE", "policies/"+url.PathEscape(vhost)+"/"+url.PathEscape(name), nil)
	if err != nil {
		return nil, err
	}

	if res, err = executeRequest(c, req); err != nil {
		return nil, err
	}

	return res, nil
}

// DeleteAllPolicies deletes all policies. Only meant to be used
// in integration tests.
func (c *Client) DeleteAllPolicies() (err error) {
	list, err := c.ListPolicies()
	if err != nil {
		return err
	}

	for _, p := range list {
		_, err = c.DeletePolicy(p.Vhost, p.Name)
		if err != nil {
			return err
		}
	}

	return nil
}

// ListPoliciesForTarget returns policies in a virtual host filtered by their apply-to target.
func (c *Client) ListPoliciesForTarget(vhost string, target PolicyTarget) (rec []Policy, err error) {
	policies, err := c.ListPoliciesIn(vhost)
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

// ListMatchingPolicies returns policies that would match a given entity name and target.
func (c *Client) ListMatchingPolicies(vhost, name string, target PolicyTarget) (rec []Policy, err error) {
	policies, err := c.ListPoliciesIn(vhost)
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

// targetsMatch checks if a policy target is compatible with a given target.
func targetsMatch(policyTarget, target PolicyTarget) bool {
	switch target {
	case PolicyTargetQueues:
		return policyTarget == PolicyTargetQueues || policyTarget == PolicyTargetAll
	case PolicyTargetClassicQueues:
		return policyTarget == PolicyTargetQueues || policyTarget == PolicyTargetClassicQueues || policyTarget == PolicyTargetAll
	case PolicyTargetQuorumQueues:
		return policyTarget == PolicyTargetQueues || policyTarget == PolicyTargetQuorumQueues || policyTarget == PolicyTargetAll
	case PolicyTargetStreams:
		return policyTarget == PolicyTargetQueues || policyTarget == PolicyTargetStreams || policyTarget == PolicyTargetAll
	case PolicyTargetExchanges:
		return policyTarget == PolicyTargetExchanges || policyTarget == PolicyTargetAll
	case PolicyTargetAll:
		return true
	default:
		return false
	}
}

// DeletePoliciesIn deletes multiple policies in a virtual host by name.
func (c *Client) DeletePoliciesIn(vhost string, names []string) error {
	for _, name := range names {
		_, err := c.DeletePolicy(vhost, name)
		if err != nil {
			return err
		}
	}
	return nil
}
