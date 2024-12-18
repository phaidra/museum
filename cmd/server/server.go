package server

import (
	"context"
	"fmt"
	docker "github.com/docker/docker/client"
	"github.com/nats-io/nats.go"
	etcd "go.etcd.io/etcd/client/v3"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"museum/config"
	proxymode "museum/config/proxy-mode"
	"museum/controller/api"
	"museum/controller/exhibit"
	"museum/controller/health"
	"museum/http"
	"museum/ioc"
	"museum/observability"
	"museum/persistence"
	"museum/service"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func Run() {
	ctx := context.Background()
	signal.NotifyContext(ctx, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGKILL, syscall.SIGSTOP)

	c := ioc.NewContainer()

	// register logger
	ioc.RegisterSingleton[*zap.SugaredLogger](c, observability.NewLogger)

	// register config
	ioc.RegisterSingleton[config.Config](c, config.NewEnvConfig)
	cfg := ioc.Get[config.Config](c)

	// register docker
	ioc.RegisterSingleton[*docker.Client](c, service.NewDockerClient)

	// register jaeger
	ioc.RegisterSingleton[tracesdk.SpanExporter](c, observability.NewSpanExporter)
	ioc.RegisterSingleton[*observability.TracerProviderFactory](c, observability.NewTracerProviderFactory)
	ioc.RegisterSingleton[trace.TracerProvider](c, observability.NewDefaultTracerProvider)

	// register NATS
	ioc.RegisterGenerator[*nats.Conn](c, persistence.NewNatsClient)

	// register eventing
	switch cfg.GetNatsHost() {
	case "":
		ioc.RegisterSingleton[persistence.Eventing](c, persistence.NewNoopEventing)
		break
	default:
		ioc.RegisterSingleton[persistence.Eventing](c, persistence.NewNatsEventing)
	}

	// register etcd
	ioc.RegisterSingleton[*etcd.Client](c, persistence.NewEtcdClient)

	// register shared state
	ioc.RegisterSingleton[persistence.State](c, persistence.NewEtcdState)

	// register services
	ioc.RegisterSingleton[service.VolumeProvisionerFactoryService](c, service.NewVolumeProvisionerFactoryService)
	ioc.RegisterSingleton[service.RewriteService](c, service.NewRewriteService)
	ioc.RegisterSingleton[service.EnvironmentTemplateResolverService](c, service.NewEnvironmentTemplateResolverService)
	ioc.RegisterSingleton[service.LockService](c, service.NewLockService)
	ioc.RegisterSingleton[service.RuntimeInfoService](c, service.NewRuntimeInfoService)
	ioc.RegisterSingleton[service.ExhibitService](c, service.NewExhibitService)
	ioc.RegisterSingleton[service.LastAccessedService](c, service.NewLastAccessedService)

	switch cfg.GetProxyMode() {
	case proxymode.ModeSwarm:
		ioc.RegisterSingleton[service.ApplicationResolverService](c, service.NewDockerHostApplicationResolverService)
		break
	case proxymode.ModeSwarmExt:
		ioc.RegisterSingleton[service.ApplicationResolverService](c, service.NewDockerExtHostApplicationResolverService)
		break
	}

	ioc.RegisterSingleton[service.ApplicationProxyService](c, service.NewDockerApplicationProxyService)

	// register livecheck
	ioc.RegisterSingleton[*service.HttpLivecheck](c, service.NewHttpLivecheck)
	ioc.RegisterSingleton[*service.ExecLivecheck](c, service.NewExecLivecheck)
	ioc.RegisterSingleton[service.LivecheckFactoryService](c, service.NewLivecheckFactoryService)

	// register services
	ioc.RegisterSingleton[service.ApplicationProvisionerService](c, service.NewDockerApplicationProvisionerService)
	ioc.RegisterSingleton[service.ApplicationProvisionerHandlerService](c, service.NewApplicationProvisionerHandlerService)
	ioc.RegisterSingleton[service.ExhibitCleanupService](c, service.NewExhibitCleanupService)

	// register router and routes
	ioc.RegisterSingleton[*http.Mux](c, http.NewMux)
	ioc.ForFunc(c, health.RegisterRoutes)
	ioc.ForFunc(c, exhibit.RegisterRoutes)
	ioc.ForFunc(c, api.RegisterRoutes)

	go ioc.ForFunc(c, startProxyServer)
	go ioc.ForFunc(c, startExhibitCleanup)

	<-ctx.Done()
}

func startExhibitCleanup(log *zap.SugaredLogger, cleanupService service.ExhibitCleanupService, exhibitService service.ExhibitService) {
	cleanup := func() {
		defer func() {
			if err := recover(); err != nil {
				log.Errorw("failed to cleanup exhibits", "error", err)
			}
		}()
		<-time.After(10 * time.Second)

		c := exhibitService.Count()
		log.Infow("checking for expired exhibits", "count", c)

		if c == 0 {
			log.Debugw("no exhibits to cleanup, skipping")
			return
		}

		err := cleanupService.Cleanup()
		if err != nil {
			log.Errorw("failed to cleanup exhibits", "error", err)
		}
	}

	for {
		cleanup()
	}
}

func startProxyServer(router *http.Mux, config config.Config, log *zap.SugaredLogger) {
	log.Infof("starting server on port %s", config.GetPort())

	if config.GetCertFile() != "" && config.GetKeyFile() != "" {
		log.Infof("using tls with cert %s and key %s", config.GetCertFile(), config.GetKeyFile())
		err := http.ConfigureTls(config.GetCertFile(), config.GetKeyFile())
		if err != nil {
			log.Panicw("failed to configure tls", "error", err)
		}
	}

	err := http.ListenAndServe(fmt.Sprintf(":%s", config.GetPort()), router)
	if err != nil {
		log.Panicw("failed to start server", "error", err)
	}
}
