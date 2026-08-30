package collector

// clampPercent limita um percentual ao intervalo 0–100.
//
// Leituras do sistema operacional podem estourar levemente esse intervalo por
// arredondamento entre núcleos, ou ficar negativas quando um contador é
// reiniciado. A API sempre publica um valor dentro da faixa.
func clampPercent(value float64) float64 {
	switch {
	case value < 0:
		return 0
	case value > 100:
		return 100
	default:
		return value
	}
}

// percentOf devolve quanto used representa de total, em percentual.
//
// Total zero devolve 0 em vez de NaN: uma partição ou dispositivo sem
// capacidade conhecida não deve contaminar o JSON com um valor não numérico,
// que quebraria a desserialização no cliente.
func percentOf(used, total uint64) float64 {
	if total == 0 {
		return 0
	}

	return clampPercent(float64(used) / float64(total) * 100)
}
