package rabbithole

import (
	"encoding/json"
	"net/http"
	"net/url"
)

//
// GET /api/vhosts
//

// Example response:

// [
//   {
//     "message_stats": {
//       "publish": 78,
//       "publish_details": {
//         "rate": 0
//       }
//     },
//     "messages": 0,
//     "messages_details": {
//       "rate": 0
//     },
//     "messages_ready": 0,
//     "messages_ready_details": {
//       "rate": 0
//     },
//     "messages_unacknowledged": 0,
//     "messages_unacknowledged_details": {
//       "rate": 0
//     },
//     "recv_oct": 16653,
//     "recv_oct_details": {
//       "rate": 0
//     },
//     "send_oct": 40495,
//     "send_oct_details": {
//       "rate": 0
//     },
//     "name": "\/",
//     "tracing": false
//   },
//   {
//     "name": "29dd51888b834698a8b5bc3e7f8623aa1c9671f5",
//     "tracing": false
//   }
// ]

type VhostInfo struct {
	Name    string `json:"name"`
	Tracing bool   `json:"tracing"`

	Messages        int         `json:"messages"`
	MessagesDetails RateDetails `json:"messages_details"`

	MessagesReady        int         `json:"messages_ready"`
	MessagesReadyDetails RateDetails `json:"messages_ready_details"`

	MessagesUnacknowledged        int         `json:"messages_unacknowledged"`
	MessagesUnacknowledgedDetails RateDetails `json:"messages_unacknowledged_details"`

	RecvOct        uint64      `json:"recv_oct"`
	SendOct        uint64      `json:"send_oct"`
	RecvCount      uint64      `json:"recv_cnt"`
	SendCount      uint64      `json:"send_cnt"`
	SendPendi      uint64      `json:"send_pend"`
	RecvOctDetails RateDetails `json:"recv_oct_details"`
	SendOctDetails RateDetails `json:"send_oct_details"`
}

func (c *Client) ListVhosts() (rec []VhostInfo, err error) {
	req, err := newGETRequest(c, "vhosts/")
	if err != nil {
		return []VhostInfo{}, err
	}

	if err = executeAndParseRequest(req, &rec); err != nil {
		return []VhostInfo{}, err
	}

	return rec, nil
}

//
// GET /api/vhosts/{name}
//

func (c *Client) GetVhost(vhostname string) (rec *VhostInfo, err error) {
	req, err := newGETRequest(c, "vhosts/"+url.QueryEscape(vhostname))
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}


//
// PUT /api/vhosts/{name}
//

type VhostSettings struct {
	Tracing bool `json:"tracing"`
}

func (c *Client) PutVhost(vhostname string, settings VhostSettings) (res *http.Response, err error) {
	body, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}

	req, err := newRequestWithBody(c, "PUT", "vhosts/"+url.QueryEscape(vhostname), body)
	if err != nil {
		return nil, err
	}

	res, err = executeRequest(c, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

//
// DELETE /api/vhosts/{name}
//

func (c *Client) DeleteVhost(vhostname string) (res *http.Response, err error) {
	req, err := newRequestWithBody(c, "DELETE", "vhosts/"+url.QueryEscape(vhostname), nil)
	if err != nil {
		return nil, err
	}

	res, err = executeRequest(c, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
