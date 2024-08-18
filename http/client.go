package http

import (
	"net/http"
)

type Client struct {
	Client         *http.Client
	BaseURL        string
	DefaultHeaders map[string]string
}

func (c *Client) NewRequest(method string, endpoint string) (res *http.Response, err error) {
	url := c.BaseURL + endpoint
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return res, err
	}
	for key, value := range c.DefaultHeaders {
		req.Header.Set(key, value)
	}

	return c.Client.Do(req)
}
