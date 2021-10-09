package api

import (
	"go.uber.org/zap"
	prom "prometheus_remote_client"
	"time"
)

var _ prom.Service = (*loggingMiddleware)(nil)

type loggingMiddleware struct {
	logger *zap.Logger
	svc    prom.Service
}

func (l loggingMiddleware) PromRemoteWrite(particles []prom.PromParticle, url string, token string) (err error) {
	defer func(begin time.Time) {
		if err != nil {
			l.logger.Warn("method call: prom_push",
				zap.Error(err),
				zap.Duration("duration", time.Since(begin)))
		} else {
			l.logger.Info("method call: prom_push",
				zap.Duration("duration", time.Since(begin)))
		}
	}(time.Now())
	return l.svc.PromRemoteWrite(particles, url, token)
}

func NewLoggingMiddleware(svc prom.Service, logger *zap.Logger) prom.Service {
	return &loggingMiddleware{
		logger: logger,
		svc:    svc,
	}
}

