# Distributed Gradle Maven Plugin

This Maven plugin enables distributed building by offloading your Maven builds to a remote distributed build service.

## Installation

Add the plugin to your `pom.xml`:

```xml
<build>
    <plugins>
        <plugin>
            <groupId>com.distributedgradle</groupId>
            <artifactId>distributed-gradle-maven-plugin</artifactId>
            <version>1.0.0</version>
        </plugin>
    </plugins>
</build>
```

## Configuration

Configure the plugin in your `pom.xml`:

```xml
<build>
    <plugins>
        <plugin>
            <groupId>com.distributedgradle</groupId>
            <artifactId>distributed-gradle-maven-plugin</artifactId>
            <version>1.0.0</version>
            <configuration>
                <serviceUrl>https://build-service.company.com</serviceUrl>
                <authToken>${env.BUILD_TOKEN}</authToken>
                <taskName>package</taskName>
                <cacheEnabled>true</cacheEnabled>
                <timeoutMinutes>30</timeoutMinutes>
            </configuration>
        </plugin>
    </plugins>
</build>
```

## Usage

Run your build using the distributed service:

```bash
mvn distributed-gradle:distributed-build
```

## Configuration Options

| Property | Type | Default | Description |
|----------|------|---------|-------------|
| `serviceUrl` | String | `http://localhost:8080` | URL of the distributed build service |
| `authToken` | String | `null` | Authentication token for the service |
| `taskName` | String | `package` | Maven phase/goal to execute |
| `cacheEnabled` | Boolean | `true` | Whether to enable build caching |
| `timeoutMinutes` | Integer | `30` | Build timeout in minutes |

## Command Line Usage

You can also configure via command line:

```bash
mvn distributed-gradle:distributed-build \
    -Ddistributed.serviceUrl=https://build-service.company.com \
    -Ddistributed.authToken=your-token \
    -Ddistributed.taskName=install \
    -Ddistributed.cacheEnabled=true \
    -Ddistributed.timeoutMinutes=60
```

## Example

```xml
<project>
    <build>
        <plugins>
            <plugin>
                <groupId>com.distributedgradle</groupId>
                <artifactId>distributed-gradle-maven-plugin</artifactId>
                <version>1.0.0</version>
                <configuration>
                    <serviceUrl>https://build-service.company.com</serviceUrl>
                    <authToken>${env.BUILD_TOKEN}</authToken>
                    <taskName>deploy</taskName>
                    <cacheEnabled>true</cacheEnabled>
                    <timeoutMinutes>45</timeoutMinutes>
                </configuration>
            </plugin>
        </plugins>
    </build>
</project>
```

Then run:

```bash
mvn distributed-gradle:distributed-build
```

The plugin will:
1. Submit your project to the distributed build service
2. Monitor the build progress
3. Report the final status and any artifacts

## Requirements

- Maven 3.6.0 or higher
- Java 11 or higher
- Access to a running distributed build service