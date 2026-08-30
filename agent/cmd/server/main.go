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
	"pc-monitor-agent/internal/service"
)

const (
	// defaultAddr mantém o agente restrito ao loopback: no MVP não há
	// autenticação, então a API não deve ficar exposta na rede.
	defaultAddr = "127.0.0.1:8080"

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
	// O listener é aberto antes do Serve para que falhas de bind (porta em uso,
	// permissão negada) sejam reportadas de imediato, e não em uma goroutine.
	listener, err := net.Listen("tcp", defaultAddr)
	if err != nil {
		return fmt.Errorf("não foi possível escutar em %s: %w", defaultAddr, err)
	}

	deps := api.Deps{
		CPU:    service.NewCPUService(collector.NewCPUCollector()),
		Memory: service.NewMemoryService(collector.NewMemoryCollector()),
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

	slog.Info("servidor iniciado", "addr", listener.Addr().String())

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
