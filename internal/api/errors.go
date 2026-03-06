package api

import "fmt"

type APIError struct {
	StatusCode int    `json:"-"`
	Code       string `json:"error"`
	Message    string `json:"message"`
}

func (e *APIError) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("%s: %s", e.Code, e.Message)
	}
	return e.Code
}

func (e *APIError) ExitCode() int {
	switch {
	case e.StatusCode == 401 || e.StatusCode == 403:
		return 2
	case e.StatusCode == 422 || e.StatusCode == 400:
		return 3
	case e.StatusCode == 404:
		return 4
	case e.StatusCode == 429:
		return 5
	case e.StatusCode >= 500:
		return 6
	default:
		return 1
	}
}
