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

// Represents general response from health check endpoints if no dedicated representation is defined
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

// Checks if there are alarms in effect in the cluster
func (c *Client) HealthCheckAlarms() (rec Health, err error) {
	err = executeCheck(c, "health/checks/alarms", &rec)
	return rec, err
}

// Checks if there are local alarms in effect on the target node
func (c *Client) HealthCheckLocalAlarms() (rec Health, err error) {
	err = executeCheck(c, "health/checks/local-alarms", &rec)
	return rec, err
}

// Checks the expiration date on the certificates for every listener configured to use TLS.
// Valid units: days, weeks, months, years. The value of the within argument is the number of units.
// So, when within is 2 and unit is "months", the expiration period used by the check will be the next two months.
func (c *Client) HealthCheckCertificateExpiration(within uint, unit TimeUnit) (rec Health, err error) {
	err = executeCheck(c, "health/checks/certificate-expiration/"+strconv.Itoa(int(within))+"/"+string(unit), &rec)
	return rec, err
}

// Represents the response from HealthCheckPortListener
type PortListenerHealth struct {
	Check
	Status  string `json:"status"`
	Reason  string `json:"reason"`
	Missing string `json:"missing"`
	Port    uint   `json:"port"`
	Ports   []uint `json:"ports"`
}

// Checks if there is an active listener on the give port
func (c *Client) HealthCheckPortListener(port uint) (rec PortListenerHealth, err error) {
	err = executeCheck(c, "health/checks/port-listener/"+strconv.Itoa(int(port)), &rec)
	return rec, err
}

// Represents the response from HealthCheckProtocolListener
type ProtocolListenerHealth struct {
	Check
	Status    string   `json:"status"`
	Reason    string   `json:"reason"`
	Missing   string   `json:"missing"`
	Protocols []string `json:"protocols"`
}

// Checks if there is an active listener for the given protocol
// Valid protocol names are: amqp091, amqp10, mqtt, stomp, web-mqtt, web-stomp.
func (c *Client) HealthCheckProtocolListener(protocol Protocol) (rec ProtocolListenerHealth, err error) {
	err = executeCheck(c, "health/checks/protocol-listener/"+string(protocol), &rec)
	return rec, err
}

// Checks if all virtual hosts and running on the target node
func (c *Client) HealthCheckVirtualHosts() (rec Health, err error) {
	err = executeCheck(c, "health/checks/virtual-hosts", &rec)
	return rec, err
}

// Checks if there are classic mirrored queues without synchronised mirrors online (queues that would potentially lose data if the target node is shut down).
func (c *Client) HealthCheckNodeIsMirrorSyncCritical() (rec Health, err error) {
	err = executeCheck(c, "health/checks/node-is-mirror-sync-critical", &rec)
	return rec, err
}

// Checks if there are quorum queues with minimum online quorum (queues that would lose their quorum and availability if the target node is shut down).
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
