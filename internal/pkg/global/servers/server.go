package servers

import (
	"fmt"

	"ghippo.io/api/wssdk/v1alpha1/types"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	hertzconfig "github.com/cloudwego/hertz/pkg/common/config"
	"github.com/hertz-contrib/obs-opentelemetry/provider"
	hertztracing "github.com/hertz-contrib/obs-opentelemetry/tracing"

	"github.com/auth-engine/config"
	"github.com/auth-engine/internal/pkg/routers/routeinit"
)

const (
	engineService = "auth-engine"
)

type ServerOpt struct {
	ServiceName string
	Opts        []hertzconfig.Option
	AppHandler  []app.HandlerFunc
}

func NewServerOpt(serverConfig config.Server, tracerConfig config.Tracer) (*ServerOpt, error) {
	opt := ServerOpt{
		ServiceName: engineService,
		Opts:        nil,
		AppHandler:  nil,
	}
	if tracerConfig.Enable {
		tracer, cfg := hertztracing.NewServerTracer()
		opt.Opts = append(opt.Opts, tracer)
		opt.AppHandler = append(opt.AppHandler, hertztracing.ServerMiddleware(cfg))
	}
	host := fmt.Sprintf("0.0.0.0:%d", serverConfig.Port)
	opt.Opts = append(opt.Opts, server.WithHostPorts(host), server.WithMaxRequestBodySize(1*1024*1024*1024))
	return &opt, nil
}

type WebServer struct {
	ghippoSdk types.Interface
	Web       *server.Hertz
	routeInit *routeinit.RouteInit
}

func NewWebServer(hertzServer *server.Hertz, ghippoSdk types.Interface, tracerConfig config.Tracer, routeInit *routeinit.RouteInit) (*WebServer, error) {
	if tracerConfig.Enable {
		_ = provider.NewOpenTelemetryProvider(
			provider.WithServiceName(engineService),
			// Support setting ExportEndpoint via environment variables: OTEL_EXPORTER_OTLP_ENDPOINT
			provider.WithExportEndpoint(tracerConfig.Endpoint),
			provider.WithInsecure(),
		)
	}
	return &WebServer{
		ghippoSdk: ghippoSdk,
		Web:       hertzServer,
		routeInit: routeInit,
	}, nil
}

func (w *WebServer) Run() {
	defer w.ghippoSdk.Stop()
	w.Web.Spin()
}
