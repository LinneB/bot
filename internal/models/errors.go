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

// SystemError is a generic error type, for boring things like HTTP or JSON parsing, etc.
type SystemError struct {
	Err error
}

func NewSystemError(err error) *SystemError {
	return &SystemError{Err: err}
}

func (se *SystemError) Error() string {
	return fmt.Sprintf("System Error: %s", se.Err)
}

func (se *SystemError) Unwrap() error {
	return se.Err
}
