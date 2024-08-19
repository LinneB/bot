package helix

import (
	"fmt"
	"net/http"
)

type Client struct {
	ClientID   string
	HelixURL   string
	HttpClient *http.Client
	Token      string

	userIDCache map[string]int
}

func (c *Client) NewRequest(method string, endpoint string) (*http.Request, error) {
	req, err := http.NewRequest(method, endpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))
	req.Header.Set("Client-ID", c.ClientID)
	return req, nil
}

func (c *Client) ValidateToken() (bool, error) {
	url := "https://id.twitch.tv/oauth2/validate"
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.Token))

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return false, err
	}

	if res.StatusCode == 200 {
		return true, nil
	} else {
		return false, nil
	}
}
