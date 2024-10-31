package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	infisical "github.com/infisical/go-sdk"
)

var (
	hostUrl                   = flag.String("host-url", "https://app.infisical.com", "infisical instance URL")
	listenAddr                = flag.String("listen-addr", "/opt/infisicalkms.socket", "gRPC listen address")
	kmsKeyId                  = flag.String("kms-key", "", "Infisical KMS key ID")
	caCertificate             = flag.String("ca-certificate", "", "instance certificate for SSL/TLS")
	identityId                = flag.String("identity-id", "", "machine identity ID")
	uaClientId                = flag.String("ua-client-id", "", "universal auth client ID")
	uaClientSecret            = flag.String("ua-client-secret", "", "universal auth client secret")
	resource                  = flag.String("azure-resource", "", "azure resource")
	serviceAccountKeyfilePath = flag.String("service-account-keyfile-path", "", "path of the service account file")
	healthzPort               = flag.Int("healthz-port", 8787, "port for health check")
	healthzPath               = flag.String("healthz-path", "/healthz", "path for health check")
	healthzTimeout            = flag.Duration("healthz-timeout", 20*time.Second, "RPC timeout for health check")
)

func run(grpcService *GRPCService, healthCheckService *HealthCheckService) error {
	signalsChan := make(chan os.Signal, 1)
	signal.Notify(signalsChan, syscall.SIGINT, syscall.SIGTERM)

	healthErrorChan := healthCheckService.Start()
	pluginErrorChan := grpcService.Start()

	for {
		select {
		case sig := <-signalsChan:
			return fmt.Errorf("captured %v, shutting down plugin", sig)
		case pluginErr := <-pluginErrorChan:
			return pluginErr
		case healthErr := <-healthErrorChan:
			log.Printf("Health check error: %+v", healthErr)
			return nil
		}
	}
}

func main() {
	flag.Parse()
	if *kmsKeyId == "" {
		log.Fatalln("Error: KMS key ID is missing")
	}

	infisicalClient := infisical.NewInfisicalClient(context.Background(), infisical.Config{
		SiteUrl:       *hostUrl,
		CaCertificate: *caCertificate,
	})

	auth := AuthHandler{
		infisicalClient:           &infisicalClient,
		clientId:                  *uaClientId,
		clientSecret:              *uaClientSecret,
		identityId:                *identityId,
		resource:                  *resource,
		serviceAccountKeyfilePath: *serviceAccountKeyfilePath,
	}
	if err := auth.login(); err != nil {
		log.Fatalf("Error authenticating with Infisical: %+v\n", err)
	}

	healthCheckService := NewHealthCheckService(
		*healthzPath,
		*healthzPort,
		*healthzTimeout,
		*listenAddr,
	)

	grpcService := NewGRPCService(*listenAddr, 20*time.Second, NewInfisicalKmsService(&infisicalClient, *kmsKeyId))

	if err := run(grpcService, healthCheckService); err != nil {
		log.Fatalf("Error running services: %+v\n", err)
	}
}
