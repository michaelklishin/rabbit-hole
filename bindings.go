package rabbithole

import (
	"net/url"
)

//
// GET /api/bindings
//

// Example response:
//
// [
//   {
//     "source": "",
//     "vhost": "\/",
//     "destination": "amq.gen-Dzw36tPTm_VsmILY9oTG9w",
//     "destination_type": "queue",
//     "routing_key": "amq.gen-Dzw36tPTm_VsmILY9oTG9w",
//     "arguments": {
//
//     },
//     "properties_key": "amq.gen-Dzw36tPTm_VsmILY9oTG9w"
//   }
// ]

type BindingInfo struct {
	Source          string                 `json:"source"`
	Vhost           string                 `json:"vhost"`
	Destination     string                 `json:"destination"`
	DestinationType string                 `json:"destination_type"`
	RoutingKey      string                 `json:"routing_key"`
	Arguments       map[string]interface{} `json:"arguments"`
	PropertiesKey   string                 `json:"properties_key"`
}

func (c *Client) ListBindings() (rec []BindingInfo, err error) {
	req, err := newGETRequest(c, "bindings/")
	if err != nil {
		return []BindingInfo{}, err
	}

	if err = executeAndParseRequest(req, &rec); err != nil {
		return []BindingInfo{}, err
	}

	return rec, nil
}

//
// GET /api/bindings/{vhost}
//

func (c *Client) ListBindingsIn(vhost string) (rec []BindingInfo, err error) {
	req, err := newGETRequest(c, "bindings/"+url.QueryEscape(vhost))
	if err != nil {
		return []BindingInfo{}, err
	}

	if err = executeAndParseRequest(req, &rec); err != nil {
		return []BindingInfo{}, err
	}

	return rec, nil
}

//
// GET /api/queues/{vhost}/{queue}/bindings
//

// Example response:
// [
//   {"source":"",
//    "vhost":"/",
//    "destination":"amq.gen-H0tnavWatL7g7uU2q5cAPA",
//    "destination_type":"queue",
//    "routing_key":"amq.gen-H0tnavWatL7g7uU2q5cAPA",
//    "arguments":{},
//    "properties_key":"amq.gen-H0tnavWatL7g7uU2q5cAPA"},
//   {"source":"temp",
//    "vhost":"/",
//    "destination":"amq.gen-H0tnavWatL7g7uU2q5cAPA",
//    "destination_type":"queue",
//    "routing_key":"",
//    "arguments":{},
//    "properties_key":"~"}
// ]

func (c *Client) ListBindingsOfQueue(vhost, queue string) (rec []BindingInfo, err error) {
	req, err := newGETRequest(c, "queues/"+url.QueryEscape(vhost)+"/"+url.QueryEscape(queue)+"/bindings")
	if err != nil {
		return []BindingInfo{}, err
	}

	if err = executeAndParseRequest(req, &rec); err != nil {
		return []BindingInfo{}, err
	}

	return rec, nil
}