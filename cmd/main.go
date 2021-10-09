package main

import (
	"fmt"
	"go.uber.org/zap"
	"net/http"
	"os"
	"os/signal"
	prom "prometheus_remote_client"
	"prometheus_remote_client/api"
	"syscall"
)

const svcName = "prom_client"

func main() {
	logger, _ := zap.NewProduction()

	svc := prom.New(logger)
	svc = api.NewLoggingMiddleware(svc, logger)

	errs := make(chan error, 2)

	go startHTTPServer(svc, logger, errs)

	go func() {
		c := make(chan os.Signal)
		signal.Notify(c, syscall.SIGINT)
		errs <- fmt.Errorf("%s", <-c)
	}()

	err := <-errs
	logger.Error(fmt.Sprintf("Prom client service terminated: %s", err))
}

func startHTTPServer(svc prom.Service, logger *zap.Logger, errs chan error) {
	port := fmt.Sprintf(":%d", 8080)
	logger.Info(fmt.Sprintf("Prom client service started using http on port %d", 8080))
	errs <- http.ListenAndServe(port, api.MakeHandler(svcName, svc))
}