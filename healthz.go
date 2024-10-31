package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	kmsapi "k8s.io/kms/apis/v2" // The Kubernetes KMS API
)

type HealthCheckService struct {
	path           string
	port           int
	pluginUnixPath string
	timeout        time.Duration
}

func (h *HealthCheckService) Start() chan error {
	host := fmt.Sprintf("0.0.0.0:%d", *healthzPort)
	errorCh := make(chan error)
	mux := http.NewServeMux()
	mux.HandleFunc(h.path, h.HandlerFunc)
	log.Println("Starting up health check server...")
	go func() {
		defer close(errorCh)
		select {
		case errorCh <- http.ListenAndServe(host, mux):
		default:
		}
	}()

	return errorCh
}

func (h *HealthCheckService) HandlerFunc(w http.ResponseWriter, r *http.Request) {
	_, cancel := context.WithTimeout(r.Context(), h.timeout)
	defer cancel()

	conn, err := grpc.NewClient(fmt.Sprintf("unix://%s", h.pluginUnixPath), grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	defer conn.Close()

	client := kmsapi.NewKeyManagementServiceClient(conn)
	status, err := client.Status(context.Background(), &kmsapi.StatusRequest{})
	if err != nil {
		log.Printf("Health check failed: %v", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	log.Printf("Health check: %+v", status)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func NewHealthCheckService(path string, port int, timeout time.Duration, pluginUnixPath string) *HealthCheckService {
	return &HealthCheckService{
		path:           path,
		port:           port,
		timeout:        timeout,
		pluginUnixPath: pluginUnixPath,
	}
}
