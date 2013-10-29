package rabbithole

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
)

type Client struct {
	Endpoint, Host, Username, Password string
}

func NewClient(uri string, username string, password string) (me *Client, err error) {
	u, err := url.Parse(uri)
	if err != nil {
		return nil, err
	}

	me = &Client{
		Endpoint: uri,
		Host:     u.Host,
		Username: username,
		Password: password,
	}

	return me, nil
}

func newGETRequest(client *Client, path string) (*http.Request, error) {
	s := client.Endpoint + "/api/" + path

	req, err := http.NewRequest("GET", s, nil)
	req.SetBasicAuth(client.Username, client.Password)
	// set Opaque to preserve percent-encoded path. MK.
	req.URL.Opaque = "//" + client.Host + "/api/" + path

	return req, err
}

func newRequestWithBody(client *Client, method string, path string, body []byte) (*http.Request, error) {
	s := client.Endpoint + "/api/" + path

	req, err := http.NewRequest(method, s, bytes.NewReader(body))
	req.SetBasicAuth(client.Username, client.Password)
	// set Opaque to preserve percent-encoded path. MK.
	req.URL.Opaque = "//" + client.Host + "/api/" + path

	req.Header.Add("Content-Type", "application/json")

	return req, err
}

func executeRequest(client *Client, req *http.Request) (res *http.Response, err error) {
	httpc := &http.Client{}

	res, err = httpc.Do(req)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func executeAndParseRequest(req *http.Request, rec interface{}) (err error) {
	client := &http.Client{}

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	if isNotFound(res) {
		return errors.New("not found")
	}

	defer res.Body.Close()

	err = json.NewDecoder(res.Body).Decode(&rec)
	if err != nil {
		return err
	}

	return nil
}

func isNotFound(res *http.Response) bool {
	return res.StatusCode == http.StatusNotFound
}
