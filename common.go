package rabbithole

// Extra arguments as a map (on queues, bindings, etc)
type Properties map[string]interface{}

// Port used by RabbitMQ or clients
type Port int

// Rate of change of a numerical value
type RateDetails struct {
	Rate float32 `json:"rate"`
}

// RabbitMQ context (Erlang app) running on
// a node
type BrokerContext struct {
	Node        string `json:"node"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Port        Port   `json:"port"`
	Ignore      bool   `json:"ignore_in_use"`
}

// Basic published messages statistics
type MessageStats struct {
	Publish        int         `json:"publish"`
	PublishDetails RateDetails `json:"publish_details"`
}
