plugins {
    kotlin("multiplatform") version "1.9.24"
}

repositories {
    mavenCentral()
}

kotlin {
    wasmWasi {
        binaries.executable()
        nodejs()
    }
    sourceSets {
        val wasmWasiMain by getting {
            dependencies {
                implementation(kotlin("stdlib"))
            }
        }
    }
}
