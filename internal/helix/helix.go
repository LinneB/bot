package helix

import (
	"bot/internal/http"
	"bot/internal/models"
	"encoding/json"
	"strconv"
)

var HelixURL = "https://api.twitch.tv/helix"
var userIDCache = make(map[string]int)

// Cached wrapper around [GetUser] for user IDs.
func LoginToID(c http.Client, login string) (userid int, found bool, err error) {
	if id, found := userIDCache[login]; found {
		return id, true, nil
	}
	user, found, err := GetUser(c, login)
	if err != nil {
		return 0, false, err
	}
	if found {
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
// Returned "found" value is false if the user doesn't exist.
func GetUser(c http.Client, login string) (user models.HelixUser, found bool, err error) {
	req := http.Request{
		Method: "GET",
		URL:    HelixURL + "/users?login=" + login,
	}
	res, err := c.GenericRequest(req)
	if err != nil {
		return models.HelixUser{}, false, &models.APIError{
			URL: req.Url(),
			Err: err,
		}
	}
	if res.StatusCode != 200 {
		return models.HelixUser{}, false, &models.APIError{
			Status: res.StatusCode,
			URL:    req.Url(),
		}
	}
	var responseStruct struct {
		Data []models.HelixUser
	}
	decoder := json.NewDecoder(res.Body)
	err = decoder.Decode(&responseStruct)
	if err != nil {
		return models.HelixUser{}, false, models.NewSystemError(err)
	}

	if len(responseStruct.Data) > 0 {
		return responseStruct.Data[0], true, nil
	} else {
		return models.HelixUser{}, false, nil
	}
}

// Attempts to fetch a stream using the /streams endpoint.
// Returned "found" value is false if the user exists, but is offline.
// Throws [models.APIError] with a 400 status if the user doesn't exist.
func GetStream(c http.Client, name string) (stream models.HelixStream, found bool, err error) {
	req := http.Request{
		Method: "GET",
		URL:    HelixURL + "/streams?user_login=" + name,
	}
	res, err := c.GenericRequest(req)
	if err != nil {
		return models.HelixStream{}, false, &models.APIError{
			URL: req.Url(),
			Err: err,
		}
	}
	if res.StatusCode != 200 {
		return models.HelixStream{}, false, &models.APIError{
			Status: res.StatusCode,
			URL:    req.Url(),
		}
	}

	var responseStruct struct {
		Data []models.HelixStream
	}
	err = json.NewDecoder(res.Body).Decode(&responseStruct)
	if err != nil {
		return models.HelixStream{}, false, models.NewSystemError(err)
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
