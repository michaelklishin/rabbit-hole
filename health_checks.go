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

// Aliveness represents response from aliveness-test endpoint
type Aliveness struct {
	Status string `json:"status"`
}

// Aliveness endpoint declares a test queue, then publishes a message and consumes a message
func (c *Client) Aliveness(vhost string) (rec *Aliveness, err error) {
	req, err := newGETRequest(c, "aliveness-test/"+vhost)
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(c, req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}
