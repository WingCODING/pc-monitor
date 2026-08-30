import org.jetbrains.compose.desktop.application.dsl.TargetFormat

plugins {
    alias(libs.plugins.kotlinMultiplatform)
    alias(libs.plugins.composeMultiplatform)
    alias(libs.plugins.composeCompiler)
    alias(libs.plugins.kotlinSerialization)
}

kotlin {
    jvmToolchain(21)
    jvm()

    sourceSets {
        val commonMain by getting {
            dependencies {
                implementation(compose.runtime)
                implementation(compose.foundation)
                implementation(compose.material3)
                implementation(compose.ui)
                implementation(libs.kotlinx.coroutines.core)
                implementation(libs.kotlinx.serialization.json)
                implementation(libs.ktor.client.core)
                implementation(libs.ktor.client.content.negotiation)
                implementation(libs.ktor.serialization.kotlinx.json)
                implementation(libs.ktor.client.websockets)
            }
        }
        val commonTest by getting {
            dependencies {
                implementation(kotlin("test"))
                implementation(libs.kotlinx.coroutines.test)
                // MockEngine responde no lugar da rede: os testes de
                // repository exercitam o mesmo caminho de desserialização e
                // de erro que a aplicação usa, sem servidor nenhum.
                implementation(libs.ktor.client.mock)
            }
        }

        val jvmMain by getting {
            dependencies {
                implementation(compose.desktop.currentOs)
                implementation(libs.ktor.client.cio)
            }
        }
    }
}

compose.desktop {
    application {
        mainClass = "pcmonitor.MainKt"

        nativeDistributions {
            // createDistributable é o caminho principal: gera um diretório com
            // a aplicação e uma JVM enxuta via jlink, sem depender de
            // ferramenta de empacotamento do sistema. O .deb fica disponível
            // para quem tiver dpkg, mas não é requisito para rodar.
            targetFormats(TargetFormat.Deb)

            packageName = "pc-monitor"
            packageVersion = "1.0.0"
            description = "Monitor de recursos do computador em tempo real"
            vendor = "PC Monitor"

            // O agente responde em JSON sobre HTTP e WebSocket; sem estes
            // módulos o jlink enxuga a JVM até quebrar a rede em tempo de
            // execução, e o erro só aparece no aplicativo empacotado.
            modules("java.net.http", "jdk.crypto.ec", "java.management")
        }
    }
}
