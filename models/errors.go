package models

import (
	"fmt"
	"net/url"
)

type APIError struct {
	// Reponse code, if any
	Status int
	URL    *url.URL
	// Error returned, if any
	Err error
}

func (ae *APIError) Error() string {
	if ae.Err != nil {
		return fmt.Sprintf("API Error: %s (Status: %d, URL Path: %s): %s", ae.URL.Host, ae.Status, ae.URL.RequestURI(), ae.Err)
	}
	return fmt.Sprintf("API Error: %s (Status: %d, URL Path: %s)", ae.URL.Host, ae.Status, ae.URL.RequestURI())
}

func (ae *APIError) Unwrap() error {
	return ae.Err
}
