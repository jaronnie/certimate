package handlers

import (
	"context"
	"net/http"

	"github.com/pocketbase/pocketbase/core"
	"github.com/pocketbase/pocketbase/tools/router"

	xenv "github.com/certimate-go/certimate/pkg/utils/env"
)

const (
	metricsContentType = "text/plain; version=0.0.4; charset=utf-8"
	metricsEnvEnabled  = "CERTIMATE_METRICS_ENABLED"
)

type metricsService interface {
	Render(ctx context.Context) ([]byte, string, error)
}

type MetricsHandler struct {
	service metricsService
}

func NewMetricsHandler(router *router.Router[*core.RequestEvent], service metricsService) {
	handler := &MetricsHandler{
		service: service,
	}

	router.GET("/metrics", handler.get)
}

func (handler *MetricsHandler) get(e *core.RequestEvent) error {
	if !xenv.GetBool(metricsEnvEnabled) {
		return e.String(http.StatusNotFound, "404 Not Found")
	}

	res, contentType, err := handler.service.Render(e.Request.Context())
	if err != nil {
		return e.String(http.StatusInternalServerError, err.Error())
	}

	if contentType == "" {
		contentType = metricsContentType
	}

	return e.Blob(http.StatusOK, contentType, res)
}
