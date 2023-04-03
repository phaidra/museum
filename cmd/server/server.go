package server

import (
	"fmt"
	"github.com/nats-io/nats.go"
	goredislib "github.com/redis/go-redis/v9"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"museum/config"
	"museum/http/api"
	"museum/http/exhibit"
	"museum/http/health"
	"museum/http/router"
	"museum/ioc"
	"museum/observability"
	"museum/persistence"
	"museum/service"
	"net/http"
)

func Run() {
	c := ioc.NewContainer()

	// register logger
	ioc.RegisterSingleton[*zap.SugaredLogger](c, observability.NewLogger)

	// register config
	ioc.RegisterSingleton[config.Config](c, config.NewEnvConfig)

	// register jaeger
	ioc.RegisterSingleton[tracesdk.SpanExporter](c, observability.NewSpanExporter)
	ioc.RegisterSingleton[*observability.TracerProviderFactory](c, observability.NewTracerProviderFactory)
	ioc.RegisterSingleton[trace.TracerProvider](c, observability.NewDefaultTracerProvider)

	// register redis
	ioc.RegisterSingleton[*goredislib.Client](c, persistence.NewRedisClient)
	ioc.RegisterSingleton[persistence.SharedPersistentState](c, persistence.NewRedisStateConnector)

	// register nats connection
	ioc.RegisterSingleton[*nats.Conn](c, persistence.NewNatsConn)

	// register nats consumer group
	ioc.RegisterSingleton[persistence.Consumer](c, persistence.NewNatsConsumer)

	// register nats producer
	ioc.RegisterSingleton[persistence.Emitter](c, persistence.NewNatsEmitter)

	// register shared state
	ioc.RegisterSingleton[persistence.SharedPersistentEmittedState](c, persistence.NewSharedPersistentEmittedState)

	// register services
	ioc.RegisterSingleton[service.ExhibitService](c, service.NewExhibitServiceImpl)

	// register router and routes
	ioc.RegisterSingleton[*router.Mux](c, router.NewMux)
	ioc.ForFunc(c, health.RegisterRoutes)
	ioc.ForFunc(c, exhibit.RegisterRoutes)
	ioc.ForFunc(c, api.RegisterRoutes)

	ioc.ForFunc(c, func(router *router.Mux, config config.Config, log *zap.SugaredLogger) {
		log.Infof("Starting server on port %s", config.GetPort())
		err := http.ListenAndServe(fmt.Sprintf(":%s", config.GetPort()), router)
		if err != nil {
			panic(err)
		}
	})

	graph := ioc.GenerateDependencyGraph(c)
	fmt.Println(graph)
}
