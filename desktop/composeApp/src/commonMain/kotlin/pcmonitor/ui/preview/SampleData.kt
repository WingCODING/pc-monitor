package pcmonitor.ui.preview

import pcmonitor.model.CpuMetrics
import pcmonitor.model.DashboardMetrics
import pcmonitor.model.DiskSummary
import pcmonitor.model.MemoryMetrics
import pcmonitor.model.NetworkSummary
import pcmonitor.model.ProcessMetrics
import pcmonitor.model.SystemMetrics

/**
 * Dados de exemplo para montar a tela antes de existir integração.
 *
 * Os valores são os observados nesta máquina durante a validação do agente, e
 * não números redondos: um mock arredondado esconde justamente os problemas de
 * formatação e alinhamento que a tela precisa resolver.
 */
val sampleSystem = SystemMetrics(
    hostname = "desktop",
    os = "linux",
    osVersion = "arch",
    architecture = "x86_64",
    uptimeSeconds = 23_176,
)

val sampleDashboardMetrics = DashboardMetrics(
    timestamp = "2026-08-30T21:01:05Z",
    cpu = CpuMetrics(
        model = "AMD Ryzen 9 9950X3D 16-Core Processor",
        cores = 16,
        threads = 32,
        usagePercent = 5.967329658155839,
    ),
    memory = MemoryMetrics(
        total = 32_449_835_008,
        used = 13_086_269_440,
        available = 19_363_565_568,
        usagePercent = 40.32769176414544,
    ),
    disk = DiskSummary(
        total = 1_015_740_882_944,
        used = 389_397_245_952,
        usagePercent = 38.33627773486678,
    ),
    network = NetworkSummary(
        downloadBytesPerSecond = 2_233_281.078710775,
        uploadBytesPerSecond = 35_688.99671055571,
    ),
    uptimeSeconds = 23_176,
)

val sampleProcesses = listOf(
    ProcessMetrics(pid = 250_015, name = "chromium", cpuPercent = 17.476565559289227, memory = 690_499_584),
    ProcessMetrics(pid = 113_512, name = "idea", cpuPercent = 4.5069538699799025, memory = 2_381_553_664),
    ProcessMetrics(pid = 111_937, name = "chromium", cpuPercent = 9.079385889765089, memory = 971_685_888),
    ProcessMetrics(pid = 245_484, name = "chromium", cpuPercent = 7.575891076870988, memory = 470_220_800),
    ProcessMetrics(pid = 225_956, name = "Discord", cpuPercent = 5.092631972674747, memory = 617_496_576),
    ProcessMetrics(pid = 1_560, name = "Hyprland", cpuPercent = 5.496435560231467, memory = 229_658_624),
    ProcessMetrics(pid = 1, name = "systemd", cpuPercent = 0.041, memory = 24_129_536),
    ProcessMetrics(pid = 247_707, name = "pc-monitor-agent", cpuPercent = 0.9, memory = 18_874_368),
)
