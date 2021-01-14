package rabbithole

// Health represents response from healthchecks endpoint
type Health struct {
	Status string `json:"status"`
	Reason string `json:"reason"`
}

// HealthChecks endpoint checks if the application is running,
// channels and queues can be listed, and that no alarms are raised
func (c *Client) HealthCheck() (rec *Health, err error) {
	req, err := newGETRequest(c, "healthchecks/node")
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

// HealthChecks endpoint checks for a given node if the application is running,
// channels and queues can be listed, and that no alarms are raised
func (c *Client) HealthCheckFor(node string) (rec *Health, err error) {
	req, err := newGETRequest(c, "healthchecks/node/"+node)
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}
