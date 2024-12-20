package krakend

import (
	"fmt"

	apikeyauth "github.com/anshulgoel27/krakend-apikey-auth"
	apikeyauthgin "github.com/anshulgoel27/krakend-apikey-auth/gin"
	basicauth "github.com/anshulgoel27/krakend-basic-auth/gin"
	ipfilter "github.com/anshulgoel27/krakend-ipfilter"
	ratelimit "github.com/anshulgoel27/krakend-ratelimit/v3/router/gin"
	botdetector "github.com/krakendio/krakend-botdetector/v2/gin"
	jose "github.com/krakendio/krakend-jose/v2"
	ginjose "github.com/krakendio/krakend-jose/v2/gin"
	lua "github.com/krakendio/krakend-lua/v2/router/gin"
	metrics "github.com/krakendio/krakend-metrics/v2/gin"
	opencensus "github.com/krakendio/krakend-opencensus/v2/router/gin"
	"github.com/luraproject/lura/v2/config"
	"github.com/luraproject/lura/v2/logging"
	"github.com/luraproject/lura/v2/proxy"
	router "github.com/luraproject/lura/v2/router/gin"
	"github.com/luraproject/lura/v2/transport/http/server"

	"github.com/gin-gonic/gin"
)

// NewHandlerFactory returns a HandlerFactory with a rate-limit and a metrics collector middleware injected
func NewHandlerFactory(logger logging.Logger, metricCollector *metrics.Metrics, rejecter jose.RejecterFactory, apiKeyAuthManager *apikeyauth.AuthKeyLookupManager) router.HandlerFactory {
	handlerFactory := router.CustomErrorEndpointHandler(logger, server.DefaultToHTTPError)

	handlerFactory = ratelimit.NewRateLimiterMw(logger, handlerFactory)
	handlerFactory = ratelimit.NewTriredRateLimiterMw(logger, handlerFactory)
	handlerFactory = lua.HandlerFactory(logger, handlerFactory)
	handlerFactory = ginjose.HandlerFactory(handlerFactory, logger, rejecter)
	handlerFactory = metricCollector.NewHTTPHandlerFactory(handlerFactory)
	handlerFactory = opencensus.New(handlerFactory)
	handlerFactory = botdetector.New(handlerFactory, logger)
	handlerFactory = basicauth.New(handlerFactory, logger)
	if apiKeyAuthManager != nil {
		handlerFactory = apikeyauthgin.NewHandlerFactory(apiKeyAuthManager, handlerFactory, logger)
	}
	handlerFactory = ipfilter.HandlerFactory(handlerFactory, logger)

	return func(cfg *config.EndpointConfig, p proxy.Proxy) gin.HandlerFunc {
		logger.Debug(fmt.Sprintf("[ENDPOINT: %s] Building the http handler", cfg.Endpoint))
		return handlerFactory(cfg, p)
	}
}

type handlerFactory struct{}

func (handlerFactory) NewHandlerFactory(l logging.Logger, m *metrics.Metrics, r jose.RejecterFactory, apiKeyAuthManager *apikeyauth.AuthKeyLookupManager) router.HandlerFactory {
	return NewHandlerFactory(l, m, r, apiKeyAuthManager)
}
