package rabbithole

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type UserInfo struct {
	Name         string `json:"name"`
	PasswordHash string `json:"password_hash"`
	Tags         string `json:"tags"`

	// *never* returned by RabbitMQ. Set by the client
	// to create/update a user. MK.
	Password string `json:"password"`
}

//
// GET /api/users
//

// Example response:
// [{"name":"guest","password_hash":"8LYTIFbVUwi8HuV/dGlp2BYsD1I=","tags":"administrator"}]

func (c *Client) ListUsers() ([]UserInfo, error) {
	var err error
	req, err := NewGETRequest(c, "users/")
	if err != nil {
		return []UserInfo{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return []UserInfo{}, err
	}

	var rec []UserInfo
	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}

//
// GET /api/users/{name}
//

func (c *Client) GetUser(username string) (*UserInfo, error) {
	var (
		rec *UserInfo
		err error
	)

	req, err := NewGETRequest(c, "users/"+url.QueryEscape(username))
	if err != nil {
		return nil, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return nil, err
	}

	if IsNotFound(res) {
		return nil, errors.New("user not found")
	}

	decoder := json.NewDecoder(res.Body)
	decoder.Decode(&rec)

	return rec, nil
}

//
// PUT /api/users/{name}
//

func (c *Client) PutUser(username string, info UserInfo) (*http.Response, error) {
	var err error

	body, err := json.Marshal(info)
	if err != nil {
		return &http.Response{}, err
	}

	req, err := NewHTTPRequestWithBody(c, "PUT", "users/"+url.QueryEscape(username), body)
	if err != nil {
		return &http.Response{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return res, err
	}

	return res, nil
}

//
// DELETE /api/users/{name}
//

func (c *Client) DeleteUser(username string) (*http.Response, error) {
	var err error

	req, err := NewHTTPRequestWithBody(c, "DELETE", "users/"+url.QueryEscape(username), nil)
	if err != nil {
		return &http.Response{}, err
	}

	res, err := ExecuteHTTPRequest(c, req)
	if err != nil {
		return res, err
	}

	return res, nil
}
