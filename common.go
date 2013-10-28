package rabbithole

type Properties map[string]interface{}

type Port int

type RateDetails struct {
	Rate int `json:"rate"`
}

type BrokerContext struct {
	Node        string `json:"node"`
	Description string `json:"description"`
	Path        string `json:"path"`
	Port        Port   `json:"port"`
	Ignore      bool   `json:"ignore_in_use"`
}

type MessageStats struct {
	Publish        int         `json:"publish"`
	PublishDetails RateDetails `json:"publish_details"`
}
