package rabbithole

import (
	"encoding/json"
	"fmt"
)

type DeprecationPhase int

func (d *DeprecationPhase) UnmarshalJSON(b []byte) error {
	var deprecationPhase string
	err := json.Unmarshal(b, &deprecationPhase)
	if err != nil {
		return fmt.Errorf("error decoding deprecation phase: %w", err)
	}
	switch deprecationPhase {
	case "permitted_by_default":
		*d = DeprecationPermittedByDefault
	case "denied_by_default":
		*d = DeprecationDeniedByDefault
	case "disconnect":
		*d = DeprecationDisconnected
	case "removed":
		*d = DeprecationRemoved
	default:
		return fmt.Errorf("unknown deprecation phase: %s", deprecationPhase)
	}
	return nil
}

// Enum values for DeprecationPhase
const (
	DeprecationPermittedByDefault DeprecationPhase = iota
	DeprecationDeniedByDefault
	DeprecationDisconnected
	DeprecationRemoved
)

// DeprecatedFeature represents a deprecated feature in RabbitMQ
type DeprecatedFeature struct {
	// Short descriptive name of a deprecated feature
	Name string `json:"name"`
	// Detailed description of a deprecated feature. It can be empty string
	Description string `json:"desc"`
	// Current deprecation phase of this feature. Allowed values are:
	//   - DeprecationPermittedByDefault
	//   - DeprecationDeniedByDefault
	//   - DeprecationDisconnected
	//   - DeprecationRemoved
	Phase DeprecationPhase `json:"deprecation_phase"`
	// URL to RabbitMQ documentation with all details explaining the reasons for deprecation, and alternatives
	DocumentationUrl string `json:"doc_url"`
	// Where the deprecation comes from. It is 'rabbit', unless the deprecation comes from a plugin
	ProvidedBy string `json:"provided_by"`
}

// ListDeprecatedFeatures retrieves a list of deprecated features from the server.
//
// Returns:
//   - A slice of DeprecatedFeature objects.
//   - An error if there was an issue with the request or parsing the response.
func (c *Client) ListDeprecatedFeatures() ([]DeprecatedFeature, error) {
	req, err := newGETRequest(c, "deprecated-features")
	if err != nil {
		return nil, fmt.Errorf("error preparing deprecated features request: %w", err)
	}

	deprecatedFeatures := make([]DeprecatedFeature, 0, 10)
	if err := executeAndParseRequest(c, req, &deprecatedFeatures); err != nil {
		return nil, fmt.Errorf("error getting deprecated features from server: %w", err)
	}

	return deprecatedFeatures, nil
}

// ListDeprecatedFeaturesUsed retrieves a list of deprecated features currently in use by the server.
//
// Returns:
//   - A slice of DeprecatedFeature objects.
//   - An error if there was an issue with the request or parsing the response.
func (c *Client) ListDeprecatedFeaturesUsed() ([]DeprecatedFeature, error) {
	req, err := newGETRequest(c, "deprecated-features/used")
	if err != nil {
		return nil, fmt.Errorf("error preparing deprecated features request: %w", err)
	}

	deprecatedFeaturesInUsed := make([]DeprecatedFeature, 0, 10)
	if err := executeAndParseRequest(c, req, &deprecatedFeaturesInUsed); err != nil {
		return nil, fmt.Errorf("error getting deprecated features in use: %w", err)
	}
	return deprecatedFeaturesInUsed, nil
}
