package api

import prom "prometheus_remote_client"

type addReq struct {
	url       string
	token     string
	particles []particle
}

func (req addReq) validate() error {
	if req.token == "" || req.url == "" {
		return prom.ErrUnauthorizedAccess
	}

	if len(req.particles) == 0 {
		return prom.ErrMalformedEntity
	}

	return nil
}

type particle struct {
	Name  string `json:"name"`
	Label string `json:"label,omitempty"`
	Value int64  `json:"value"`
}
