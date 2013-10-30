package rabbithole

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type UserInfo struct {
	Name         string `json:"name"`
	PasswordHash string `json:"password_hash"`
	Tags         string `json:"tags"`
}

type UserSettings struct {
	Name string `json:"name"`
	Tags string `json:"tags"`

	// *never* returned by RabbitMQ. Set by the client
	// to create/update a user. MK.
	Password string `json:"password"`
}

//
// GET /api/users
//

// Example response:
// [{"name":"guest","password_hash":"8LYTIFbVUwi8HuV/dGlp2BYsD1I=","tags":"administrator"}]

func (c *Client) ListUsers() (rec []UserInfo, err error) {
	req, err := newGETRequest(c, "users/")
	if err != nil {
		return []UserInfo{}, err
	}

	if err = executeAndParseRequest(req, &rec); err != nil {
		return []UserInfo{}, err
	}

	return rec, nil
}

//
// GET /api/users/{name}
//

func (c *Client) GetUser(username string) (rec *UserInfo, err error) {
	req, err := newGETRequest(c, "users/"+url.QueryEscape(username))
	if err != nil {
		return nil, err
	}

	if err = executeAndParseRequest(req, &rec); err != nil {
		return nil, err
	}

	return rec, nil
}

//
// PUT /api/users/{name}
//

func (c *Client) PutUser(username string, info UserSettings) (res *http.Response, err error) {
	body, err := json.Marshal(info)
	if err != nil {
		return nil, err
	}

	req, err := newRequestWithBody(c, "PUT", "users/"+url.QueryEscape(username), body)
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
// DELETE /api/users/{name}
//

func (c *Client) DeleteUser(username string) (res *http.Response, err error) {
	req, err := newRequestWithBody(c, "DELETE", "users/"+url.QueryEscape(username), nil)
	if err != nil {
		return nil, err
	}

	res, err = executeRequest(c, req)
	if err != nil {
		return nil, err
	}

	return res, nil
}
