package rabbithole

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
