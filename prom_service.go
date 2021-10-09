package prometheus_remote_client

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"os"
	"prometheus_remote_client/prometheus"
	"regexp"
	"strings"
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

type Service interface {
	PromRemoteWrite(particles map[string]interface{}, url string, token string) error
}

var _ Service = (*promService)(nil)

type promService struct {
	logger *zap.Logger
}

func (p promService) PromRemoteWrite(particles map[string]interface{}, url string, token string) error {
	var tsList = prometheus.TSList{}
	//statsMap := structs.Map(particles)
	convertToPromParticle(particles, "", &tsList)

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

func convertToPromParticle(m map[string]interface{}, label string, tsList *prometheus.TSList) {
	for k, v := range m {
		switch c := v.(type) {
		case map[string]interface{}:
			if k == "name" || k == "estimate" {
				for _, value := range c {
					m, ok := value.(map[string]interface{})
					if !ok {
						continue
					}
					var lbl string
					var dtpt interface{}
					for k, v := range m {
						switch k {
						case "name":
							lbl = fmt.Sprintf("%v", v)
						case "estimate":
							dtpt = v
						}
					}
					tsList = makePromParticle(label+k, lbl, dtpt, tsList, false)
					//fmt.Printf("%s{intance=%q name:%q} %d\n", camelToSnake(label+k), "gw", lbl, dtpt)
				}
			} else {
				convertToPromParticle(c, label+k, tsList)
			}
		case interface{}:
			tsList = makePromParticle(label+k, "", v, tsList, false)
			//fmt.Printf("%s{intance=%q} %d\n", camelToSnake(label+k), "gw", v)
		}
	}
}

func makePromParticle(label string, k string, v interface{}, tsList *prometheus.TSList, quantile bool) *prometheus.TSList {
	var dpFlag prometheus.Dp
	var labelsListFlag prometheus.LabelList
	labelsListFlag.Set(fmt.Sprintf("__name__:%s", camelToSnake(label)))
	labelsListFlag.Set("instance:gw")
	labelsListFlag.Set(fmt.Sprintf("name:%s", k))
	dpFlag.Set(fmt.Sprintf("now,%d", v))
	*tsList = append(*tsList, prometheus.TimeSeries{
		Labels:    []prometheus.Label(labelsListFlag),
		Datapoint: prometheus.Datapoint(dpFlag),
	})
	return tsList
}

func camelToSnake(s string) string {
	var matchFirstCap = regexp.MustCompile("(.)([A-Z][a-z]+)")
	var matchAllCap = regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(s, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	lower := strings.ToLower(snake)
	return lower
}

func New(logger *zap.Logger) Service {
	return &promService{
		logger: logger,
	}
}
