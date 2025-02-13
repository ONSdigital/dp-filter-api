package main

import (
	"context"
	goerrors "errors"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/ONSdigital/dp-filter-api/config"
	"github.com/ONSdigital/dp-filter-api/service"
	dpotelgo "github.com/ONSdigital/dp-otel-go"
	"github.com/ONSdigital/log.go/v2/log"
	"github.com/pkg/errors"
)

const serviceName = "dp-filter-api"

var (
	// BuildTime represents the time in which the service was built
	BuildTime string
	// GitCommit represents the commit (SHA-1) hash of the service that is running
	GitCommit string
	// Version represents the version of the service that is running
	Version string

	/* NOTE: replace the above with the below to run code with for example vscode debugger.
	BuildTime string = "1601119818"
	GitCommit string = "6584b786caac36b6214ffe04bf62f058d4021538"
	Version   string = "v0.1.0"

	*/
)

func main() {
	log.Namespace = serviceName
	ctx := context.Background()

	if err := run(ctx); err != nil {
		log.Error(ctx, "application unexpectedly failed", err)
		os.Exit(1)
	}
}

func run(ctx context.Context) error {
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	svcErrors := make(chan error, 1)

	// Read config
	cfg, err := config.Get()
	if err != nil {
		log.Fatal(ctx, "unable to retrieve configuration", err)
		return err
	}

	// sensitive fields are omitted from config.String()
	log.Info(ctx, "loaded config", log.Data{"config": cfg})

	// Set up Open Telemetry
	otelConfig := dpotelgo.Config{
		OtelServiceName:          cfg.OTServiceName,
		OtelExporterOtlpEndpoint: cfg.OTExporterOTLPEndpoint,
		OtelBatchTimeout:         cfg.OTBatchTimeout,
	}

	otelShutdown, oErr := dpotelgo.SetupOTelSDK(ctx, otelConfig)
	if oErr != nil {
		return fmt.Errorf("error setting up OpenTelemetry - hint: ensure OTEL_EXPORTER_OTLP_ENDPOINT is set. %w", oErr)
	}
	// Handle shutdown properly so nothing leaks.
	defer func() {
		err = goerrors.Join(err, otelShutdown(context.Background()))
	}()

	// Run the service
	svc := service.New()
	if err := svc.Init(ctx, cfg, BuildTime, GitCommit, Version); err != nil {
		return errors.Wrap(err, "running service failed")
	}
	svc.Start(ctx, svcErrors)

	// Blocks until an os interrupt or a fatal error occurs
	select {
	case err := <-svcErrors:
		log.Error(ctx, "service error received", err)
		svc.Close(ctx)
		return err
	case sig := <-signals:
		log.Info(ctx, "os signal received", log.Data{"signal": sig})
	}
	return svc.Close(ctx)
}
