pluginManagement {
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}

dependencyResolution {
    repositories {
        google()
        mavenCentral()
    }
}

rootProject.name = "MakoClaw"

include(":app")

// Core modules
include(":core:core-common")
include(":core:core-model")
include(":core:core-network")
include(":core:core-database")
include(":core:core-datastore")
include(":core:core-security")
include(":core:core-ui")

// Feature modules
include(":feature:feature-auth")
include(":feature:feature-chat")
include(":feature:feature-tasks")
include(":feature:feature-dashboard")
include(":feature:feature-agents")
include(":feature:feature-knowledge")
include(":feature:feature-memory")
include(":feature:feature-history")
include(":feature:feature-settings")
include(":feature:feature-skills")
include(":feature:feature-workflows")
include(":feature:feature-metrics")
include(":feature:feature-cron")
include(":feature:feature-files")
include(":feature:feature-mcp")
include(":feature:feature-reports")
