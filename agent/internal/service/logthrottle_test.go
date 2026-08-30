package service

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
	"time"
)

// capturarLog redireciona o slog padrão para um buffer durante o teste.
func capturarLog(t *testing.T) *bytes.Buffer {
	t.Helper()

	buffer := &bytes.Buffer{}
	anterior := slog.Default()

	slog.SetDefault(slog.New(slog.NewTextHandler(buffer, &slog.HandlerOptions{Level: slog.LevelWarn})))
	t.Cleanup(func() { slog.SetDefault(anterior) })

	return buffer
}

func TestWarnThrottleRegistraAPrimeiraOcorrencia(t *testing.T) {
	buffer := capturarLog(t)

	newWarnThrottle(time.Minute).warn("cpu", "cpu indisponível")

	if !strings.Contains(buffer.String(), "cpu indisponível") {
		t.Errorf("primeira ocorrência não foi registrada: %q", buffer.String())
	}
}

// TestWarnThrottleSilenciaRepeticoes reproduz o cenário que motivou o
// throttle: o loop do WebSocket coletando a cada segundo com um collector
// quebrado.
func TestWarnThrottleSilenciaRepeticoes(t *testing.T) {
	buffer := capturarLog(t)

	instante := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	throttle := newWarnThrottle(30 * time.Second)
	throttle.now = func() time.Time { return instante }

	// 20 coletas, uma por segundo, todas falhando.
	for range 20 {
		throttle.warn("cpu", "cpu indisponível")
		instante = instante.Add(time.Second)
	}

	if linhas := strings.Count(buffer.String(), "cpu indisponível"); linhas != 1 {
		t.Errorf("%d linhas registradas, esperado 1", linhas)
	}
}

// TestWarnThrottleInformaAsOmitidas: sem a contagem, o log sugeriria uma falha
// esporádica onde havia uma falha contínua.
func TestWarnThrottleInformaAsOmitidas(t *testing.T) {
	buffer := capturarLog(t)

	instante := time.Date(2026, 8, 30, 12, 0, 0, 0, time.UTC)
	throttle := newWarnThrottle(30 * time.Second)
	throttle.now = func() time.Time { return instante }

	for range 40 {
		throttle.warn("cpu", "cpu indisponível")
		instante = instante.Add(time.Second)
	}

	saida := buffer.String()

	if linhas := strings.Count(saida, "cpu indisponível"); linhas != 2 {
		t.Fatalf("%d linhas registradas, esperado 2", linhas)
	}

	if !strings.Contains(saida, "repeticoesOmitidas=29") {
		t.Errorf("contagem de omitidas ausente ou errada: %q", saida)
	}
}

// TestWarnThrottleSeparaChaves garante que uma falha de disco não silencia uma
// falha de CPU.
func TestWarnThrottleSeparaChaves(t *testing.T) {
	buffer := capturarLog(t)

	throttle := newWarnThrottle(time.Minute)

	throttle.warn("cpu", "cpu indisponível")
	throttle.warn("disk", "disco indisponível")
	throttle.warn("cpu", "cpu indisponível")

	saida := buffer.String()

	if strings.Count(saida, "cpu indisponível") != 1 {
		t.Errorf("repetição de cpu não foi silenciada: %q", saida)
	}

	if !strings.Contains(saida, "disco indisponível") {
		t.Errorf("falha de disco silenciada por outra chave: %q", saida)
	}
}
