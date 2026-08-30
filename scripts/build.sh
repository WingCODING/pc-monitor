#!/usr/bin/env bash
#
# Compila o agente Go e o aplicativo desktop, deixando os dois em dist/.
#
# O resultado não depende de Go nem de JDK instalados: o binário do agente é
# estático e o desktop leva uma JVM enxuta gerada pelo jlink.

set -euo pipefail

raiz="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
dist="$raiz/dist"

echo "==> agente Go"
mkdir -p "$dist"

# CGO desligado: o gopsutil lê /proc em Go puro no Linux, e sem CGO o binário
# não depende da libc da máquina que compilou.
(
    cd "$raiz/agent"
    CGO_ENABLED=0 go build -trimpath -ldflags "-s -w" -o "$dist/pc-monitor-agent" ./cmd/server
)

echo "==> aplicativo desktop"
(
    cd "$raiz/desktop"
    ./gradlew --quiet :composeApp:createDistributable
)

rm -rf "$dist/desktop"
cp -r "$raiz/desktop/composeApp/build/compose/binaries/main/app/pc-monitor" "$dist/desktop"

echo
echo "pronto:"
echo "  agente:  $dist/pc-monitor-agent"
echo "  desktop: $dist/desktop/bin/pc-monitor"
echo
echo "para subir os dois juntos: scripts/start.sh"
