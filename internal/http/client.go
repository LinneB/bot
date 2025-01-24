package http

import (
	"net/http"
	"net/url"
)

type Client struct {
	Client         *http.Client
	DefaultHeaders map[string]string
	// Default headers per hostname.
	// Hostnames are the same as [url.URL.Host], like "example.com".
	URLHeaders map[string]map[string]string
}

type Request struct {
	Method string
	URL    string
}

func (c *Client) GenericRequest(r Request) (res *http.Response, err error) {
	url, err := url.Parse(r.URL)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequest(r.Method, url.String(), nil)
	if err != nil {
		return nil, err
	}
	// Set default headers
	for key, value := range c.DefaultHeaders {
		request.Header.Set(key, value)
	}
	if headers, ok := c.URLHeaders[url.Host]; ok {
		for key, value := range headers {
			request.Header.Set(key, value)
		}
	}
	return c.Client.Do(request)
}
