package pcmonitor

import pcmonitor.network.DEFAULT_AGENT_URL
import java.io.File
import java.net.HttpURLConnection
import java.net.InetAddress
import java.net.ServerSocket
import java.net.URI
import java.util.concurrent.TimeUnit

/**
 * Conexão com o agente usada pelo aplicativo desktop.
 *
 * Em desenvolvimento, ou quando PCMON_AGENT_URL foi informado, o agente é
 * externo. Num pacote nativo, o executável incluído nos recursos é iniciado
 * como processo filho e encerrado junto com a janela.
 */
internal class AgentProcess private constructor(
    val baseUrl: String,
    private val process: Process?,
) : AutoCloseable {
    override fun close() {
        val child = process ?: return
        child.destroy()
        if (!child.waitFor(2, TimeUnit.SECONDS)) {
            child.destroyForcibly()
            child.waitFor(2, TimeUnit.SECONDS)
        }
    }

    companion object {
        fun start(): AgentProcess {
            System.getenv("PCMON_AGENT_URL")?.let {
                return AgentProcess(it, null)
            }

            val executable = packagedAgent() ?: return AgentProcess(DEFAULT_AGENT_URL, null)
            val port = availableLoopbackPort()
            val baseUrl = "http://127.0.0.1:$port"
            val child = ProcessBuilder(executable.absolutePath)
                .redirectErrorStream(true)
                .redirectOutput(ProcessBuilder.Redirect.INHERIT)
                .apply {
                    environment()["PCMON_HOST"] = "127.0.0.1"
                    environment()["PCMON_PORT"] = port.toString()
                }
                .start()

            try {
                awaitHealth(child, baseUrl)
            } catch (error: Exception) {
                child.destroyForcibly()
                throw error
            }

            return AgentProcess(baseUrl, child)
        }

        private fun packagedAgent(): File? {
            val resources = System.getProperty("compose.application.resources.dir") ?: return null
            val executableName = if (System.getProperty("os.name").startsWith("Windows", ignoreCase = true)) {
                "pc-monitor-agent.exe"
            } else {
                "pc-monitor-agent"
            }

            return File(resources, executableName).takeIf { it.isFile }
        }

        private fun availableLoopbackPort(): Int =
            ServerSocket(0, 1, InetAddress.getByName("127.0.0.1")).use { it.localPort }

        private fun awaitHealth(child: Process, baseUrl: String) {
            val deadline = System.nanoTime() + TimeUnit.SECONDS.toNanos(5)

            while (System.nanoTime() < deadline) {
                if (!child.isAlive) {
                    throw IllegalStateException("O agente encerrou durante a inicialização.")
                }

                val ready = runCatching {
                    (URI("$baseUrl/health").toURL().openConnection() as HttpURLConnection).run {
                        connectTimeout = 300
                        readTimeout = 300
                        requestMethod = "GET"
                        try {
                            responseCode in 200..299
                        } finally {
                            disconnect()
                        }
                    }
                }.getOrDefault(false)

                if (ready) return
                Thread.sleep(100)
            }

            throw IllegalStateException("O agente não respondeu em $baseUrl.")
        }
    }
}
