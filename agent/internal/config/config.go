// Package config carrega a configuração do agente a partir do ambiente.
//
// O MVP não tem arquivo de configuração: variáveis de ambiente cobrem o que é
// preciso ajustar (porta ocupada, intervalo de coleta, verbosidade) sem
// nenhuma cerimônia, e os padrões fazem o agente funcionar sem configuração
// nenhuma.
package config

import (
	"fmt"
	"log/slog"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

// Prefixo das variáveis de ambiente reconhecidas.
const envPrefix = "PCMON_"

// Padrões do MVP.
const (
	// DefaultHost mantém o agente restrito ao loopback: não há autenticação,
	// então a API não deve ficar exposta na rede.
	DefaultHost = "127.0.0.1"

	DefaultPort = 8080

	// DefaultCollectionInterval é a cadência do loop que alimenta o WebSocket.
	DefaultCollectionInterval = time.Second

	DefaultLogLevel = slog.LevelInfo
)

// Limites do intervalo de coleta. Abaixo do mínimo a varredura de processos
// passaria o tempo todo lendo /proc; acima do máximo o "tempo real" da UI
// deixaria de merecer o nome.
const (
	MinCollectionInterval = 100 * time.Millisecond
	MaxCollectionInterval = time.Minute
)

// Config reúne o que pode ser ajustado sem recompilar.
type Config struct {
	Host               string
	Port               int
	CollectionInterval time.Duration
	LogLevel           slog.Level
}

// Default devolve a configuração usada quando nada é informado.
func Default() Config {
	return Config{
		Host:               DefaultHost,
		Port:               DefaultPort,
		CollectionInterval: DefaultCollectionInterval,
		LogLevel:           DefaultLogLevel,
	}
}

// Addr devolve o endereço de escuta no formato aceito por net.Listen.
func (c Config) Addr() string {
	return net.JoinHostPort(c.Host, strconv.Itoa(c.Port))
}

// Load lê o ambiente sobre os padrões.
//
// Um valor inválido é erro, não motivo para cair no padrão em silêncio: quem
// exportou PCMON_PORT=oitenta quer aquela porta, e subir na 8080 esconderia o
// engano até alguém estranhar a porta errada.
func Load() (Config, error) {
	cfg := Default()

	if raw, ok := lookup("HOST"); ok {
		if raw == "" {
			return Config{}, fmt.Errorf("%sHOST vazio", envPrefix)
		}

		cfg.Host = raw
	}

	if raw, ok := lookup("PORT"); ok {
		port, err := strconv.Atoi(raw)
		if err != nil || port < 1 || port > 65535 {
			return Config{}, fmt.Errorf("%sPORT=%q inválido: use um inteiro entre 1 e 65535", envPrefix, raw)
		}

		cfg.Port = port
	}

	if raw, ok := lookup("COLLECTION_INTERVAL"); ok {
		interval, err := time.ParseDuration(raw)
		if err != nil {
			return Config{}, fmt.Errorf("%sCOLLECTION_INTERVAL=%q inválido: use uma duração como 1s ou 500ms", envPrefix, raw)
		}

		if interval < MinCollectionInterval || interval > MaxCollectionInterval {
			return Config{}, fmt.Errorf("%sCOLLECTION_INTERVAL=%q fora do intervalo aceito (%v a %v)",
				envPrefix, raw, MinCollectionInterval, MaxCollectionInterval)
		}

		cfg.CollectionInterval = interval
	}

	if raw, ok := lookup("LOG_LEVEL"); ok {
		level, err := parseLevel(raw)
		if err != nil {
			return Config{}, err
		}

		cfg.LogLevel = level
	}

	return cfg, nil
}

// lookup devolve o valor da variável com o prefixo do agente.
func lookup(name string) (string, bool) {
	return os.LookupEnv(envPrefix + name)
}

// parseLevel aceita os nomes de nível do slog, sem diferenciar maiúsculas.
func parseLevel(raw string) (slog.Level, error) {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn", "warning":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return 0, fmt.Errorf("%sLOG_LEVEL=%q inválido: use debug, info, warn ou error", envPrefix, raw)
	}
}
