pluginManagement {
    repositories {
        google()
        mavenCentral()
        gradlePluginPortal()
    }
}

dependencyResolutionManagement {
    repositoriesMode.set(RepositoriesMode.FAIL_ON_PROJECT_REPOS)
    repositories {
        google()
        mavenCentral()
    }
}

rootProject.name = "heroin-android"

include(":app")
include(":core-crypto")
include(":core-network") 
include(":core-storage")
include(":feature-auth")
include(":feature-messenger")
include(":feature-files")
include(":feature-settings")
