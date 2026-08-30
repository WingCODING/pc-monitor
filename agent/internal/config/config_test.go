package config

import (
	"log/slog"
	"testing"
	"time"
)

// TestDefaultAtendeOsPadroesDoMVP trava os valores que a spec exige funcionar
// sem nenhuma configuração.
func TestDefaultAtendeOsPadroesDoMVP(t *testing.T) {
	cfg := Default()

	if cfg.Host != "127.0.0.1" {
		t.Errorf("Host = %q, esperado 127.0.0.1", cfg.Host)
	}

	if cfg.Port != 8080 {
		t.Errorf("Port = %d, esperado 8080", cfg.Port)
	}

	if cfg.CollectionInterval != time.Second {
		t.Errorf("CollectionInterval = %v, esperado 1s", cfg.CollectionInterval)
	}

	if cfg.Addr() != "127.0.0.1:8080" {
		t.Errorf("Addr() = %q, esperado 127.0.0.1:8080", cfg.Addr())
	}
}

func TestLoadSemAmbienteUsaOsPadroes(t *testing.T) {
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() falhou: %v", err)
	}

	if cfg != Default() {
		t.Errorf("Load() = %+v, esperado %+v", cfg, Default())
	}
}

func TestLoadLeOAmbiente(t *testing.T) {
	t.Setenv("PCMON_HOST", "0.0.0.0")
	t.Setenv("PCMON_PORT", "9090")
	t.Setenv("PCMON_COLLECTION_INTERVAL", "500ms")
	t.Setenv("PCMON_LOG_LEVEL", "DEBUG")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load() falhou: %v", err)
	}

	esperado := Config{
		Host:               "0.0.0.0",
		Port:               9090,
		CollectionInterval: 500 * time.Millisecond,
		LogLevel:           slog.LevelDebug,
	}

	if cfg != esperado {
		t.Errorf("Load() = %+v, esperado %+v", cfg, esperado)
	}

	if cfg.Addr() != "0.0.0.0:9090" {
		t.Errorf("Addr() = %q, esperado 0.0.0.0:9090", cfg.Addr())
	}
}

// TestLoadRejeitaValorInvalido documenta a escolha de falhar em vez de cair no
// padrão: quem exportou uma porta quer aquela porta, e subir na 8080
// esconderia o engano.
func TestLoadRejeitaValorInvalido(t *testing.T) {
	casos := []struct {
		nome     string
		variavel string
		valor    string
	}{
		{"porta não numérica", "PCMON_PORT", "oitenta"},
		{"porta zero", "PCMON_PORT", "0"},
		{"porta acima do máximo", "PCMON_PORT", "70000"},
		{"host vazio", "PCMON_HOST", ""},
		{"intervalo sem unidade", "PCMON_COLLECTION_INTERVAL", "1"},
		{"intervalo curto demais", "PCMON_COLLECTION_INTERVAL", "10ms"},
		{"intervalo longo demais", "PCMON_COLLECTION_INTERVAL", "5m"},
		{"nível inexistente", "PCMON_LOG_LEVEL", "verboso"},
	}

	for _, caso := range casos {
		t.Run(caso.nome, func(t *testing.T) {
			t.Setenv(caso.variavel, caso.valor)

			if _, err := Load(); err == nil {
				t.Errorf("Load() aceitou %s=%q", caso.variavel, caso.valor)
			}
		})
	}
}

func TestParseLevelAceitaVariacoes(t *testing.T) {
	casos := map[string]slog.Level{
		"debug":   slog.LevelDebug,
		"INFO":    slog.LevelInfo,
		" warn ":  slog.LevelWarn,
		"warning": slog.LevelWarn,
		"Error":   slog.LevelError,
	}

	for entrada, esperado := range casos {
		obtido, err := parseLevel(entrada)
		if err != nil {
			t.Errorf("parseLevel(%q) falhou: %v", entrada, err)

			continue
		}

		if obtido != esperado {
			t.Errorf("parseLevel(%q) = %v, esperado %v", entrada, obtido, esperado)
		}
	}
}
