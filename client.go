package rabbithole

import (
	"bytes"
	"net/http"
	"net/url"
	"strings"
)

type Client struct {
	Endpoint, Host, Username, Password string
}

func NewClient(uri string, username string, password string) (*Client, error) {
	u, err := url.Parse(uri)
	if err != nil {
		return &Client{}, err
	}

	me := &Client{
		Endpoint: uri,
		Host:     u.Host,
		Username: username,
		Password: password}

	return me, nil
}

func NewGETRequest(client *Client, path string) (*http.Request, error) {
	s := client.Endpoint + "/api/" + path

	req, err := http.NewRequest("GET", s, nil)
	req.SetBasicAuth(client.Username, client.Password)
	// set Opaque to preserve percent-encoded path. MK.
	req.URL.Opaque = "//" + client.Host + "/api/" + path

	return req, err
}

func NewHTTPRequestWithBody(client *Client, method string, path string, body []byte) (*http.Request, error) {
	s := client.Endpoint + "/api/" + path

	req, err := http.NewRequest(method, s, bytes.NewReader(body))
	req.SetBasicAuth(client.Username, client.Password)
	// set Opaque to preserve percent-encoded path. MK.
	req.URL.Opaque = "//" + client.Host + "/api/" + path

	req.Header.Add("Content-Type", "application/json")

	return req, err
}

func ExecuteHTTPRequest(client *Client, req *http.Request) (*http.Response, error) {
	httpc := &http.Client{}

	return httpc.Do(req)
}

func IsNotFound(res *http.Response) bool {
	return strings.HasPrefix(res.Status, "404")
}
