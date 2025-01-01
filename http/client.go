package http

import (
	"net/http"
)

type Client struct {
	Client         *http.Client
	DefaultHeaders map[string]string
}

type Request struct {
	Method string
	URL    string
}

func (c *Client) GenericRequest(r Request) (res *http.Response, err error) {
	request, err := http.NewRequest(r.Method, r.URL, nil)
	if err != nil {
		return nil, err
	}
	for key, value := range c.DefaultHeaders {
		request.Header.Set(key, value)
	}
	return c.Client.Do(request)
}
