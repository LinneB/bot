package helix

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type Client struct {
	ClientID   string
	HelixURL   string
	HttpClient *http.Client
	Token      string

	UserIDCache map[string]int
}

// Cached wrapper around [Client.GetUser] exclusively for user IDs.
// Return value is nil if the user is not found.
func (c *Client) LoginToID(login string) (*int, error) {
	if id, found := c.UserIDCache[login]; found {
		return &id, nil
	}
	user, err := c.GetUser(login)
	if err != nil {
		return nil, err
	}
	if user != nil {
		id, err := strconv.Atoi(user.Id)
		if err != nil {
			return nil, err
		}
		c.UserIDCache[login] = id
		return &id, nil
	}
	return nil, nil
}

// Attempts to fetch a user using the /users endpoint.
// Returns nil if user was not found.
func (c *Client) GetUser(login string) (user *User, err error) {
	req, err := c.NewRequest("GET", "/users?login="+login)
	if err != nil {
		return nil, err
	}
	res, err := c.HttpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, &ErrorStatus{
			res.StatusCode,
		}
	}

	var responseStruct struct {
		Data []User
	}
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&responseStruct)
	if err != nil {
		return nil, err
	}

	if len(responseStruct.Data) > 0 {
		return &responseStruct.Data[0], nil
	} else {
		return nil, nil
	}
}

func (c *Client) NewRequest(method string, endpoint string) (*http.Request, error) {
	req, err := http.NewRequest(method, c.HelixURL+endpoint, nil)
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
