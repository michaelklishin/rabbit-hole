package rabbithole

import (
	"encoding/json"
	"net/http"
	"strconv"
)

type TimeUnit string

const (
	SECONDS TimeUnit = "seconds"
	DAYS    TimeUnit = "days"
	MONTHS  TimeUnit = "months"
	YEARS   TimeUnit = "years"
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

type Check interface {
	// Returns true if the check is ok, otherwise false
	Ok() bool

	// Returns true if the check failed, otherwise false
	Failed() bool
}

// Health represents response from healthchecks endpoint
type Health struct {
	Check
	Status string `json:"status"`
	Reason string `json:"reason"`
}

func (h *Health) Ok() bool {
	return h.Status == "ok"
}

func (h *Health) Failed() bool {
	return !h.Ok()
}

// Responds a 200 OK if there are no alarms in effect in the cluster, otherwise responds with a 503 Service Unavailable.
func (c *Client) HealthCheckAlarms() (rec Health, err error) {
	err = executeCheck(c, "health/checks/alarms", &rec)
	return rec, err
}

// Responds a 200 OK if there are no local alarms in effect on the target node, otherwise responds with a 503 Service Unavailable.
func (c *Client) HealthCheckLocalAlarms() (rec Health, err error) {
	err = executeCheck(c, "health/checks/local-alarms", &rec)
	return rec, err
}

// Checks the expiration date on the certificates for every listener configured to use TLS.
// Responds a 200 OK if all certificates are valid (have not expired), otherwise responds with a 503 Service Unavailable.
// Valid units: days, weeks, months, years. The value of the within argument is the number of units.
// So, when within is 2 and unit is "months", the expiration period used by the check will be the next two months.
func (c *Client) HealthCheckCertificateExpiration(within uint, unit TimeUnit) (rec Health, err error) {
	err = executeCheck(c, "health/checks/certificate-expiration/"+strconv.Itoa(int(within))+"/"+string(unit), &rec)
	return rec, err
}

type PortListenerHealth struct {
	Check
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Missing string `json:"missing"'`
	Port    uint   `json:"port"`
	Ports   []uint `json:"ports"`
}

// Responds a 200 OK if there is an active listener on the give port, otherwise responds with a 503 Service Unavailable.
func (c *Client) HealthCheckPortListener(port uint) (rec PortListenerHealth, err error) {
	err = executeCheck(c, "health/checks/port-listener/"+strconv.Itoa(int(port)), &rec)
	return rec, err
}

type ProtocolListenerHealth struct {
	Check
	Status    string   `json:"status"`
	Reason    string   `json:"reason"`
	Missing   string   `json:"missing"`
	Protocols []string `json:"protocols"`
}

// Responds a 200 OK if there is an active listener for the given protocol, otherwise responds with a 503 Service Unavailable.
// Valid protocol names are: amqp091, amqp10, mqtt, stomp, web-mqtt, web-stomp.
func (c *Client) HealthCheckProtocolListener(protocol Protocol) (rec ProtocolListenerHealth, err error) {
	err = executeCheck(c, "health/checks/protocol-listener/"+string(protocol), &rec)
	return rec, err
}

// Responds a 200 OK if all virtual hosts and running on the target node, otherwise responds with a 503 Service Unavailable.
func (c *Client) HealthCheckVirtualHosts() (rec Health, err error) {
	err = executeCheck(c, "health/checks/virtual-hosts", &rec)
	return rec, err
}

// Checks if there are classic mirrored queues without synchronised mirrors online (queues that would potentially lose data if the target node is shut down).
// Responds a 200 OK if there are no such classic mirrored queues, otherwise responds with a 503 Service Unavailable.
func (c *Client) HealthCheckNodeIsMirrorSyncCritical() (rec Health, err error) {
	err = executeCheck(c, "health/checks/node-is-mirror-sync-critical", &rec)
	return rec, err
}

// Checks if there are quorum queues with minimum online quorum (queues that would lose their quorum and availability if the target node is shut down).
// Responds a 200 OK if there are no such quorum queues, otherwise responds with a 503 Service Unavailable.
func (c *Client) HealthCheckNodeIsQuorumCritical() (rec Health, err error) {
	err = executeCheck(c, "health/checks/node-is-quorum-critical", &rec)
	return rec, err
}

func executeCheck(client *Client, path string, rec interface{}) error {
	req, err := newGETRequest(client, path)
	httpc := &http.Client{
		Timeout: client.timeout,
	}
	if client.transport != nil {
		httpc.Transport = client.transport
	}
	resp, err := httpc.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()

	if resp.StatusCode < http.StatusBadRequest || resp.StatusCode == http.StatusServiceUnavailable {
		if err = json.NewDecoder(resp.Body).Decode(&rec); err != nil {
			return err
		}

		return nil
	}

	if err = parseResponseErrors(resp); err != nil {
		return err
	}

	return nil
}
