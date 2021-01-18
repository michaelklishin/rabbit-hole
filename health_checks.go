package rabbithole

import "strconv"

type TimeUnit string

const (
	DAYS   TimeUnit = "days"
	WEEKS  TimeUnit = "weeks"
	MONTHS TimeUnit = "months"
	YEARS  TimeUnit = "years"
)

type Protocol string

const (
	AMQP091   Protocol = "amqp091"
	AMQP10    Protocol = "amqp10"
	MQTT      Protocol = "mqtt"
	STOMP     Protocol = "stomp"
	WEB_MQTT  Protocol = "web-mqtt"
	WEB_STOMP Protocol = "web-stomp"
)

// Responds a 200 OK if there are no alarms in effect in the cluster, otherwise responds with a 503 Service Unavailable.
func (c *Client) HealthCheckAlarms() error {
	req, err := newGETRequest(c, "health/checks/alarms")
	if err != nil {
		return err
	}

	if _, err = executeRequest(c, req); err != nil {
		return err
	}

	return nil
}

// Responds a 200 OK if there are no local alarms in effect on the target node, otherwise responds with a 503 Service Unavailable.
func (c *Client) HealthCheckLocalAlarms() error {
	req, err := newGETRequest(c, "health/checks/local-alarms")
	if err != nil {
		return err
	}

	if _, err = executeRequest(c, req); err != nil {
		return err
	}

	return nil
}


// Checks the expiration date on the certificates for every listener configured to use TLS.
// Responds a 200 OK if all certificates are valid (have not expired), otherwise responds with a 503 Service Unavailable.
// Valid units: days, weeks, months, years. The value of the within argument is the number of units.
// So, when within is 2 and unit is "months", the expiration period used by the check will be the next two months.
func (c *Client) HealthCheckCertificateExpiration(within uint, unit TimeUnit) (rec *Health, err error) {
	req, err := newGETRequest(c, "health/checks/certificate-expiration/"+strconv.Itoa(int(within))+"/"+string(unit))
	if err != nil {
		return nil, err
	}

	if _, err = executeRequest(c, req); err != nil {
		return nil, err
	}

	return rec, nil
}

// Responds a 200 OK if there is an active listener on the give port, otherwise responds with a 503 Service Unavailable.
func (c *Client) HealthCheckPortListenerListener(port uint) error {
	req, err := newGETRequest(c, "health/checks/port-listener/"+strconv.Itoa(int(port)))
	if err != nil {
		return err
	}

	if _, err = executeRequest(c, req); err != nil {
		return err
	}

	return nil
}

// Responds a 200 OK if there is an active listener for the given protocol, otherwise responds with a 503 Service Unavailable.
// Valid protocol names are: amqp091, amqp10, mqtt, stomp, web-mqtt, web-stomp.
func (c *Client) HealthCheckProtocolListener(protocol Protocol) error {
	req, err := newGETRequest(c, "health/checks/protocol-listener/"+string(protocol))
	if err != nil {
		return err
	}

	if _, err = executeRequest(c, req); err != nil {
		return err
	}

	return nil
}

// Responds a 200 OK if all virtual hosts and running on the target node, otherwise responds with a 503 Service Unavailable.
func (c *Client) HealthCheckVirtualHosts() error {
	req, err := newGETRequest(c, "health/checks/virtual-hosts")
	if err != nil {
		return err
	}

	if _, err = executeRequest(c, req); err != nil {
		return err
	}

	return nil
}

// Checks if there are classic mirrored queues without synchronised mirrors online (queues that would potentially lose data if the target node is shut down).
// Responds a 200 OK if there are no such classic mirrored queues, otherwise responds with a 503 Service Unavailable.
func (c *Client) HealthCheckNodeIsMirrorSyncCritical() error {
	req, err := newGETRequest(c, "health/checks/node-is-mirror-sync-critical")
	if err != nil {
		return err
	}

	if _, err = executeRequest(c, req); err != nil {
		return err
	}

	return nil
}

// Checks if there are quorum queues with minimum online quorum (queues that would lose their quorum and availability if the target node is shut down).
// Responds a 200 OK if there are no such quorum queues, otherwise responds with a 503 Service Unavailable.
func (c *Client) HealthCheckNodeIsQuorumCritical() error {
	req, err := newGETRequest(c, "health/checks/node-is-quorum-critical")
	if err != nil {
		return err
	}

	if _, err = executeRequest(c, req); err != nil {
		return err
	}

	return nil
}

// Deprecated health check api

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
