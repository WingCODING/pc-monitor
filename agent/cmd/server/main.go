// Command server sobe o agente HTTP do PC Monitor.
package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pc-monitor-agent/internal/api"
	"pc-monitor-agent/internal/collector"
	"pc-monitor-agent/internal/config"
	"pc-monitor-agent/internal/service"
)

const (
	readHeaderTimeout = 5 * time.Second
	shutdownTimeout   = 10 * time.Second
)

func main() {
	if err := run(); err != nil {
		slog.Error("agente encerrado com erro", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("configuração inválida: %w", err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: cfg.LogLevel})))

	// O listener é aberto antes do Serve para que falhas de bind (porta em uso,
	// permissão negada) sejam reportadas de imediato, e não em uma goroutine.
	listener, err := net.Listen("tcp", cfg.Addr())
	if err != nil {
		return fmt.Errorf("não foi possível escutar em %s: %w", cfg.Addr(), err)
	}

	cpuService := service.NewCPUService(collector.NewCPUCollector())
	memoryService := service.NewMemoryService(collector.NewMemoryCollector())
	systemService := service.NewSystemService(collector.NewSystemCollector())
	diskService := service.NewDiskService(collector.NewDiskCollector())
	networkService := service.NewNetworkService(collector.NewNetworkCollector())
	processService := service.NewProcessService(collector.NewProcessCollector())

	deps := api.Deps{
		CPU:     cpuService,
		Memory:  memoryService,
		System:  systemService,
		Disk:    diskService,
		Network: networkService,
		Process: processService,
		Metrics: service.NewMetricsService(cpuService, memoryService, systemService, diskService, networkService),
	}

	server := &http.Server{
		Handler:           api.NewRouter(deps),
		ReadHeaderTimeout: readHeaderTimeout,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	serveErr := make(chan error, 1)
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serveErr <- err
		}
	}()

	slog.Info("servidor iniciado",
		"addr", listener.Addr().String(),
		"collectionInterval", cfg.CollectionInterval,
		"logLevel", cfg.LogLevel,
	)

	select {
	case err := <-serveErr:
		return fmt.Errorf("servidor interrompido: %w", err)
	case <-ctx.Done():
		slog.Info("sinal recebido, encerrando servidor")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("encerramento não concluído: %w", err)
	}

	slog.Info("servidor encerrado")

	return nil
}
