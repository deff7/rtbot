package rutracker

type Collection struct {
	query    string
	data     []TorrentFile
	offset   int
	pageSize int
}

func (c *Client) NewCollection(q string) (*Collection, error) {
	data, err := c.List(q)
	if err != nil {
		return nil, err
	}
	return &Collection{
		query: q,
		data:  data,
		// TODO: move to config
		pageSize: 10,
	}, nil
}

func (c *Collection) ListNext() []TorrentFile {
	if !c.HasNext() {
		return nil
	}

	limit := c.offset + c.pageSize
	if limit > len(c.data) {
		limit = len(c.data)
	}
	files := c.data[c.offset:limit]
	c.offset = limit
	return files
}

func (c *Collection) HasNext() bool {
	return len(c.data)-c.offset > c.pageSize
}
