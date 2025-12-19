# Distributed Gradle Plugin

This Gradle plugin enables distributed building by offloading your Gradle builds to a remote distributed build service.

## Installation

Add the plugin to your `build.gradle`:

```gradle
plugins {
    id 'com.distributedgradle.plugin' version '1.0.0'
}
```

## Configuration

Configure the plugin in your `build.gradle`:

```gradle
distributedGradle {
    serviceUrl = 'http://your-build-service:8080'
    authToken = 'your-auth-token'  // optional
    taskName = 'build'             // default: 'build'
    cacheEnabled = true            // default: true
    timeoutMinutes = 30            // default: 30
}
```

## Usage

Run your build using the distributed service:

```bash
./gradlew distributedBuild
```

## Configuration Options

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `serviceUrl` | String | `http://localhost:8080` | URL of the distributed build service |
| `authToken` | String | `null` | Authentication token for the service |
| `taskName` | String | `build` | Gradle task to execute |
| `cacheEnabled` | Boolean | `true` | Whether to enable build caching |
| `timeoutMinutes` | Integer | `30` | Build timeout in minutes |

## Example

```gradle
plugins {
    id 'com.distributedgradle.plugin' version '1.0.0'
}

distributedGradle {
    serviceUrl = 'https://build-service.company.com'
    authToken = System.getenv('BUILD_TOKEN')
    taskName = 'assemble'
    cacheEnabled = true
    timeoutMinutes = 60
}
```

Then run:

```bash
./gradlew distributedBuild
```

The plugin will:
1. Submit your project to the distributed build service
2. Monitor the build progress
3. Report the final status and any artifacts

## Requirements

- Gradle 6.0 or higher
- Java 11 or higher
- Access to a running distributed build service