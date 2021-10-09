package api

import (
	"context"
	"encoding/json"
	"net/http"
)

const (
	ContentType = "application/json"
)

type ErrorRes struct {
	Err string `json:"error"`
}

// Response contains HTTP response specific methods.
type Response interface {
	// Code returns HTTP response code.
	Code() int

	// Headers returns map of HTTP headers with their values.
	Headers() map[string]string

	// Empty indicates if HTTP response has content.
	Empty() bool
}

func EncodeResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	w.Header().Set("Content-Type", ContentType)

	if ar, ok := response.(Response); ok {
		for k, v := range ar.Headers() {
			w.Header().Set(k, v)
		}

		w.WriteHeader(ar.Code())

		if ar.Empty() {
			return nil
		}
	}

	return json.NewEncoder(w).Encode(response)
}
