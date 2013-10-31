package rabbithole

//
// GET /api/permissions
//

// Example response:
//
// [{"user":"guest","vhost":"/","configure":".*","write":".*","read":".*"}]

type PermissionInfo struct {
	User       string `json:"user"`
	Vhost      string `json:"vhost"`

	Configure  string `json:"configure"`
	Write      string `json:"write"`
	Read       string `json:"read"`
}

func (c *Client) ListPermissions() (rec []PermissionInfo, err error) {
	req, err := newGETRequest(c, "permissions/")
	if err != nil {
		return []PermissionInfo{}, err
	}

	if err = executeAndParseRequest(req, &rec); err != nil {
		return []PermissionInfo{}, err
	}

	return rec, nil
}