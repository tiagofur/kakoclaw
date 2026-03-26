plugins {
    alias(libs.plugins.android.library)
    alias(libs.plugins.kotlin.android)
    alias(libs.plugins.compose.compiler)
}

android {
    namespace = "com.makoclaw.core.ui"
    compileSdk = 35

    defaultConfig { minSdk = 26 }

    compileOptions {
        sourceCompatibility = JavaVersion.VERSION_17
        targetCompatibility = JavaVersion.VERSION_17
    }

    kotlinOptions { jvmTarget = "17" }

    buildFeatures {
        compose = true
    }
}

dependencies {
    implementation(project(":core:core-model"))

    implementation(platform(libs.compose.bom))
    api(libs.compose.ui)
    api(libs.compose.ui.graphics)
    api(libs.compose.material3)
    api(libs.compose.material.icons.extended)
    api(libs.compose.animation)
    api(libs.compose.foundation)
    debugApi(libs.compose.ui.tooling)
    api(libs.compose.ui.tooling.preview)

    api(libs.lifecycle.runtime.compose)
    api(libs.lifecycle.viewmodel.compose)

    implementation(libs.coil.compose)
}
