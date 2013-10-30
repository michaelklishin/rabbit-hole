package rabbithole

func (c *Client) EnabledProtocols() (rec []string, err error) {
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
