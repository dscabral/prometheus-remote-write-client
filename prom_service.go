package prometheus_remote_client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"os"
	"prometheus_remote_client/prometheus"
)

var (
	// ErrMalformedEntity indicates malformed entity specification (e.g. invalid author or content).
	ErrMalformedEntity = errors.New("malformed entity specification")
	// ErrNotFound indicates a non-existent entity request.
	ErrNotFound = errors.New("non-existent entity")
	// ErrUnsupportedContentType indicates a unsupported content-type (should be application/json)
	ErrUnsupportedContentType = errors.New("unsupported content-type")
	// ErrUnauthorizedAccess indicates missing credentials
	ErrUnauthorizedAccess = errors.New("unauthorized access")
	//ErrInvalidQueryParams indicates problems with the params
	ErrInvalidQueryParams = errors.New("invalid query params")
)

type PromParticle struct {
	Name  string
	Label string
	Value int64
}

type Service interface {
	PromRemoteWrite(particles []PromParticle, url string, token string) error
}

var _ Service = (*promService)(nil)

type promService struct {
	logger *zap.Logger
}

func (p promService) PromRemoteWrite(particles []PromParticle, url string, token string) error {
	var tsList = prometheus.TSList{}
	convertToPromParticle(particles, &tsList)

	p.logger.Info("writing to", zap.String("url", url))

	//var writeURLFlag string
	//flag.StringVar(&writeURLFlag, "u", config.Url, "remote write endpoint")
	cfg := prometheus.NewConfig(
		prometheus.WriteURLOption(url),
	)

	promClient, err := prometheus.NewClient(cfg)
	if err != nil {
		p.logger.Error("unable to construct client", zap.Error(err))
	}

	var headers = make(map[string]string)
	headers["Authorization"] = token
	result, writeErr := promClient.WriteTimeSeries(context.Background(), tsList,
		prometheus.WriteOptions{Headers: headers})
	if err := error(writeErr); err != nil {
		json.NewEncoder(os.Stdout).Encode(struct {
			Success    bool   `json:"success"`
			Error      string `json:"error"`
			StatusCode int    `json:"statusCode"`
		}{
			Success:    false,
			Error:      err.Error(),
			StatusCode: writeErr.StatusCode(),
		})
		os.Stdout.Sync()

		p.logger.Error("remote write error", zap.Error(err))
	}

	json.NewEncoder(os.Stdout).Encode(struct {
		Success    bool `json:"success"`
		StatusCode int  `json:"statusCode"`
	}{
		Success:    true,
		StatusCode: result.StatusCode,
	})
	os.Stdout.Sync()

	p.logger.Info("write success")
	return nil
}

func convertToPromParticle(p []PromParticle, tsList *prometheus.TSList) {
	for _, particle := range p {
		tsList = makePromParticle(particle.Label, particle.Name, particle.Value, tsList)
	}
}

func makePromParticle(labelName string, k string, v interface{}, tsList *prometheus.TSList) *prometheus.TSList {
	mapQuantiles := make(map[string]float64)
	mapQuantiles["P50"] = 0.50
	mapQuantiles["P90"] = 0.90
	mapQuantiles["P95"] = 0.95
	mapQuantiles["P99"] = 0.99

	var dpFlag prometheus.Dp
	var labelsListFlag prometheus.LabelList
	labelsListFlag.Set(fmt.Sprintf("__name__:%s", labelName))
	labelsListFlag.Set("instance:demo_project")
	if k != "" {
		if value, ok := mapQuantiles[k]; ok {
			labelsListFlag.Set(fmt.Sprintf("quantile:%.2f", value))
			fmt.Printf("%s{intance=%q, quantile=%q} %d\n", labelName, "demo_project", k, value)
		} else {
			labelsListFlag.Set(fmt.Sprintf("name:%s", k))
			fmt.Printf("%s{intance=%q, name=%q} %d\n", labelName, "demo_project", k, v)
		}
	} else {
		fmt.Printf("%s{intance=%q} %d\n", labelName, "demo_project", v)
	}
	dpFlag.Set(fmt.Sprintf("now,%d", v))
	*tsList = append(*tsList, prometheus.TimeSeries{
		Labels:    []prometheus.Label(labelsListFlag),
		Datapoint: prometheus.Datapoint(dpFlag),
	})

	return tsList
}

func New(logger *zap.Logger) Service {
	return &promService{
		logger: logger,
	}
}
