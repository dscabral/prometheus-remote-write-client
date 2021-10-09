package api

import (
	"context"
	"encoding/json"
	"github.com/go-kit/kit/endpoint"
	prom "prometheus_remote_client"
)

func addPostEnpoint(svc prom.Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(addReq)
		if err := req.validate(); err != nil {
			return promRes{}, err
		}

		var particles map[string]interface{}
		err = json.Unmarshal(req.body, &particles)
		if err != nil {
			return promRes{}, err
		}

		svc.PromRemoteWrite(particles, req.url, req.token)

		return promRes{created: true}, nil
	}
}
