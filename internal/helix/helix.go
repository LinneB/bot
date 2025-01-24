package helix

import (
	"bot/internal/http"
	"bot/internal/models"
	"encoding/json"
	"fmt"
	"strconv"
)

var helixURL = "https://api.twitch.tv/helix"
var userIDCache = make(map[string]int)

type ErrorStatus struct {
	StatusCode int
}

func (e *ErrorStatus) Error() string {
	return fmt.Sprintf("Requested resource returned unhandled status code: %d", e.StatusCode)
}

// Cached wrapper around [Client.GetUser] exclusively for user IDs.
func LoginToID(c http.Client, login string) (userid int, found bool, err error) {
	if id, found := userIDCache[login]; found {
		return id, true, nil
	}
	user, err := GetUser(c, login)
	if err != nil {
		return 0, false, err
	}
	if user != nil {
		id, err := strconv.Atoi(user.Id)
		if err != nil {
			return 0, false, err
		}
		userIDCache[login] = id
		return id, true, nil
	}
	return 0, false, nil
}

// Attempts to fetch a user using the /users endpoint.
// Returns nil if user was not found.
func GetUser(c http.Client, login string) (user *models.HelixUser, err error) {
	res, err := c.GenericRequest(http.Request{
		Method: "GET",
		URL:    helixURL + "/users?login=" + login,
	})
	if err != nil {
		return nil, err
	}
	if res.StatusCode != 200 {
		return nil, &ErrorStatus{
			res.StatusCode,
		}
	}
	var responseStruct struct {
		Data []models.HelixUser
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

func GetStream(c http.Client, name string) (stream models.HelixStream, found bool, err error) {
	res, err := c.GenericRequest(http.Request{
		Method: "GET",
		URL:    helixURL + "/streams?user_login=" + name,
	})
	if err != nil {
		return models.HelixStream{}, false, err
	}
	if res.StatusCode != 200 {
		return models.HelixStream{}, false, &ErrorStatus{
			res.StatusCode,
		}
	}

	var responseStruct struct {
		Data []models.HelixStream
	}
	err = json.NewDecoder(res.Body).Decode(&responseStruct)
	if err != nil {
		return models.HelixStream{}, false, err
	}

	if len(responseStruct.Data) > 0 {
		return responseStruct.Data[0], true, nil
	} else {
		return models.HelixStream{}, false, nil
	}
}

func ValidateToken(c http.Client) (bool, error) {
	res, err := c.GenericRequest(http.Request{
		Method: "GET",
		URL:    "https://id.twitch.tv/oauth2/validate",
	})
	if err != nil {
		return false, err
	}
	if res.StatusCode == 200 {
		return true, nil
	} else {
		return false, nil
	}
}
