package rabbithole

import (
	"encoding/json"
	"net/url"
)

type BackingQueueStatus struct {
	Q1                    int     `json:"q1"`
	Q2                    int     `json:"q2"`
	Q3                    int     `json:"q3"`
	Q4                    int     `json:"q4"`
	Length                int64   `json:"len"`
	PendingAcks           int64   `json:"pending_acks"`
	RAMMessageCount       int64   `json:"ram_msg_count"`
	RAMAckCount           int64   `json:"ram_ack_count"`
	PersistentCount       int64   `json:"persistent_count"`
	AverageIngressRate    float64 `json:"avg_ingress_rate"`
	AverageEgressRate     float64 `json:"avg_egress_rate"`
	AverageAckIngressRate float32 `json:"avg_ack_ingress_rate"`
	AverageAckEgressRate  float32 `json:"avg_ack_egress_rate"`
}

type OwnerPidDetails struct {
	Name     string `json:"name"`
	PeerPort Port   `json:"peer_port"`
	PeerHost string `json:"peer_host"`
}

type QueueInfo struct {
	Name       string                 `json:"name"`
	Vhost      string                 `json:"vhost"`
	Durable    bool                   `json:"durable"`
	AutoDelete bool                   `json:"auto_delete"`
	Arguments  map[string]interface{} `json:"arguments"`

	Node   string `json:"node"`
	Status string `json:"status"`

	Memory               int64  `json:"memory"`
	Consumers            int    `json:"consumers"`
	ExclusiveConsumerTag string `json:"exclusive_consumer_tag"`

	Policy string `json:"policy"`

	Messages        int         `json:"messages"`
	MessagesDetails RateDetails `json:"messages_details"`

	MessagesReady        int         `json:"messages_ready"`
	MessagesReadyDetails RateDetails `json:"messages_ready_details"`

	MessagesUnacknowledged        int         `json:"messages_unacknowledged"`
	MessagesUnacknowledgedDetails RateDetails `json:"messages_unacknowledged_details"`

	MessageStats MessageStats `json:"message_stats"`

	OwnerPidDetails OwnerPidDetails `json:"owner_pid_details"`

	BackingQueueStatus BackingQueueStatus `json:"backing_queue_status"`
}

type DetailedQueueInfo QueueInfo

//
// GET /api/queues
//

// [
//   {
//     "owner_pid_details": {
//       "name": "127.0.0.1:46928 -> 127.0.0.1:5672",
//       "peer_port": 46928,
//       "peer_host": "127.0.0.1"
//     },
//     "message_stats": {
//       "publish": 19830,
//       "publish_details": {
//         "rate": 5
//       }
//     },
//     "messages": 15,
//     "messages_details": {
//       "rate": 0
//     },
//     "messages_ready": 15,
//     "messages_ready_details": {
//       "rate": 0
//     },
//     "messages_unacknowledged": 0,
//     "messages_unacknowledged_details": {
//       "rate": 0
//     },
//     "policy": "",
//     "exclusive_consumer_tag": "",
//     "consumers": 0,
//     "memory": 143112,
//     "backing_queue_status": {
//       "q1": 0,
//       "q2": 0,
//       "delta": [
//         "delta",
//         "undefined",
//         0,
//         "undefined"
//       ],
//       "q3": 0,
//       "q4": 15,
//       "len": 15,
//       "pending_acks": 0,
//       "target_ram_count": "infinity",
//       "ram_msg_count": 15,
//       "ram_ack_count": 0,
//       "next_seq_id": 19830,
//       "persistent_count": 0,
//       "avg_ingress_rate": 4.9920127795527,
//       "avg_egress_rate": 4.9920127795527,
//       "avg_ack_ingress_rate": 0,
//       "avg_ack_egress_rate": 0
//     },
//     "status": "running",
//     "name": "amq.gen-QLEaT5Rn_ogbN3O8ZOQt3Q",
//     "vhost": "rabbit\/hole",
//     "durable": false,
//     "auto_delete": false,
//     "arguments": {
//       "x-message-ttl": 5000
//     },
//     "node": "rabbit@marzo"
//   }
// ]

func (c *Client) ListQueues() ([]QueueInfo, error) {
	var err error
	req, err := NewGETRequest(c, "queues")
	if err != nil {
		return []QueueInfo{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return []QueueInfo{}, err
	}

	var rec []QueueInfo
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}

//
// GET /api/queues/{vhost}
//

func (c *Client) ListQueuesIn(vhost string) ([]QueueInfo, error) {
	var err error
	req, err := NewGETRequest(c, "queues/"+url.QueryEscape(vhost))
	if err != nil {
		return []QueueInfo{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return []QueueInfo{}, err
	}

	var rec []QueueInfo
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}

//
// GET /api/queues/{vhost}/{name}
//

func (c *Client) GetQueue(vhost, queue string) (DetailedQueueInfo, error) {
	var err error
	req, err := NewGETRequest(c, "queues/"+url.QueryEscape(vhost)+"/"+queue)
	if err != nil {
		return DetailedQueueInfo{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return DetailedQueueInfo{}, err
	}

	var rec DetailedQueueInfo
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}
