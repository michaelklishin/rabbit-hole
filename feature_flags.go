package rabbithole

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

// Feature flags are a mechanism that controls what features are considered to be enabled or available on all cluster nodes.
// If a FeatureFlag is enabled so is its associated feature (or behavior).
// If not then all nodes in the cluster will disable the feature (behavior).
type FeatureFlag struct {
	Name string `json:"name"`
	// Desc is the description of the feature flag.
	Desc string `json:"desc,omitempty"`
	// DocURL is the URL to a webpage to learn more about the feature flag.
	DocURL    string    `json:"doc_url,omitempty"`
	State     State     `json:"state,omitempty"`
	Stability Stability `json:"stability,omitempty"`
	// ProvidedBy is the RabbitMQ component or plugin which provides the feature flag.
	ProvidedBy string `json:"provided_by,omitempty"`
}

type State string

const (
	StateEnabled  State = "enabled"
	StateDisabled State = "disabled"
	// StateUnsupported means that one or more nodes in the cluster do not know this feature flag (and therefore it cannot be enabled).
	StateUnsupported State = "unsupported"
)

type Stability string

const (
	StabilityStable       Stability = "stable"
	StabilityExperimental Stability = "experimental"
)

//
// GET /api/feature-flags
//

// ListFeatureFlags lists all feature flags.
func (c *Client) ListFeatureFlags() (rec []FeatureFlag, err error) {
	req, err := newGETRequest(c, "feature-flags")
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

//
// PUT /api/feature-flags/{name}/enable
//

// EnableFeatureFlag enables a feature flag.
func (c *Client) EnableFeatureFlag(featureFlagName string) (res *http.Response, err error) {
	body, err := json.Marshal(FeatureFlag{Name: featureFlagName})
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("feature-flags/%s/enable", url.PathEscape(featureFlagName))
	req, err := newRequestWithBody(c, "PUT", path, body)
	if err != nil {
		return nil, err
	}

	if res, err = executeRequest(c, req); err != nil {
		return nil, err
	}

	return res, nil
}
