package api

import prom "prometheus_remote_client"

type addReq struct {
	body []byte
	url string
	token string
}

func (req addReq) validate() error {
	if req.token == "" || req.url == "" {
		return prom.ErrUnauthorizedAccess
	}

	if req.body == nil {
		return prom.ErrMalformedEntity
	}
	return nil
}