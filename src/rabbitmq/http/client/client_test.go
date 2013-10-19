package client

import (
	"testing"
	"regexp"
)

func AssertRegexpMatch(t *testing.T, pattern string, payload []byte) {
	matched, err := regexp.Match(pattern, payload)
	if err != nil || !matched {
		t.Log(err)
		t.Error("%s does not match %s", string(payload), pattern)
		t.Fail()
	}
}

func AssertSuccessfulResponse(t *testing.T, response interface{}, err error) {
	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestOverviewWithValidCredentials(t *testing.T) {
	c := NewClient("http://127.0.0.1:15672", "guest", "guest")

	t.Log("GET /api/overview")
	res, err := c.Overview()

	AssertSuccessfulResponse(t, res, err)
	AssertRegexpMatch(t, `^3\.\d+\.\d+`, []byte(res.ManagementVersion))
}
