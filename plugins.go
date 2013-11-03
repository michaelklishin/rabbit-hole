package rabbithole

func (c *Client) EnabledProtocols() ([]string, error) {
	var err error
	overview, err := c.Overview()
	if err != nil {
		return []string{}, err
	}

	// we really need to implement Map/Filter/etc. MK.
	l := len(overview.Listeners)
	var xs = make([]string, l)
	for i, lnr := range overview.Listeners {
		xs[i] = lnr.Protocol
	}

	return xs, nil
}

func (c *Client) ProtocolPorts() (map[string]Port, error) {
	var (
		err error
		res map[string]Port = map[string]Port{}
	)
	overview, err := c.Overview()
	if err != nil {
		return map[string]Port{}, err
	}

	for _, lnr := range overview.Listeners {
		res[lnr.Protocol] = lnr.Port
	}

	return res, nil
}
