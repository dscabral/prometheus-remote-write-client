package api

import (
	"context"
	kithttp "github.com/go-kit/kit/transport/http"
	"github.com/go-zoo/bone"
	"io/ioutil"
	"net/http"
	prom "prometheus_remote_client"
	"strings"
)

func MakeHandler(svcName string, svc prom.Service) http.Handler {
	opts := []kithttp.ServerOption{
		kithttp.ServerErrorEncoder(encodeError),
	}
	r := bone.New()

	r.Post("/prom/push", kithttp.NewServer(
		addPostEnpoint(svc),
		decodeAddRequest,
		EncodeResponse,
		opts...,
	))

	return r
}

func decodeAddRequest(_ context.Context, r *http.Request) (interface{}, error) {
	if !strings.Contains(r.Header.Get("Content-Type"), "application/json") {
		return nil, prom.ErrUnsupportedContentType
	}

	vals := bone.GetQuery(r, "url")
	if len(vals) > 1 {
		return "", prom.ErrInvalidQueryParams
	}

	req := addReq{
		url:       vals[0],
		token:     r.Header.Get("Authorization"),
	}

	data, err := ioutil.ReadAll(r.Body)
	if err == nil && data != nil {
    	req.body = data
	} else {
		return nil, prom.ErrMalformedEntity
	}
	return req, nil
}

func encodeError(_ context.Context, err error, w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	switch err {
	case prom.ErrNotFound:
		w.WriteHeader(http.StatusNotFound)
	case prom.ErrMalformedEntity:
		w.WriteHeader(http.StatusBadRequest)
	case prom.ErrUnsupportedContentType:
		w.WriteHeader(http.StatusUnsupportedMediaType)
	case prom.ErrUnauthorizedAccess:
		w.WriteHeader(http.StatusUnauthorized)
	default:
		w.WriteHeader(http.StatusInternalServerError)
	}
}
