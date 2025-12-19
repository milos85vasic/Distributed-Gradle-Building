---
title: "5-Minute Quick Start"
weight: 10
---

# 5-Minute Quick Start

Get up and running with Distributed Gradle Building System in just 5 minutes! This tutorial will guide you through the fastest path to your first distributed build.

## ⏱️ What You'll Accomplish

In the next 5 minutes, you will:
- Set up a basic distributed build environment
- Run your first distributed Gradle build
- See the performance improvement firsthand

## 📋 Prerequisites

Before you start, make sure you have:

- ✅ **Linux or macOS machine** (Docker support on Windows works too)
- ✅ **Java 8+ installed**
- ✅ **Git installed**
- ✅ **A Gradle project** (we provide a sample if you don't have one)

That's it! No complex configuration required.

## 🚀 Let's Get Started

### Minute 1: Clone and Setup

```bash
# Clone the repository
git clone https://github.com/milos85vasic/Distributed-Gradle-Building
cd Distributed-Gradle-Building

# Run the quick setup script
./setup_quick_start.sh
```

This script will:
- Download all dependencies
- Configure a local coordinator
- Set up a cache server
- Create a basic worker pool

### Minute 2: Create a Test Project

If you have a Gradle project, you can skip this step. Otherwise, let's create a sample project:

```bash
# Create a sample Java project
mkdir sample-project
cd sample-project

# Create a basic build.gradle
cat > build.gradle << 'EOF'
plugins {
    id 'java'
}

repositories {
    mavenCentral()
}

dependencies {
    implementation 'org.apache.commons:commons-lang3:3.12.0'
}

task timedBuild {
    doLast {
        println "Build completed at: ${new Date()}"
        Thread.sleep(2000) // Simulate work
    }
}
EOF

# Create the directory structure
mkdir -p src/main/java/com/example

# Create a simple Java class
cat > src/main/java/com/example/Hello.java << 'EOF'
package com.example;

public class Hello {
    public static void main(String[] args) {
        System.out.println("Hello from Distributed Gradle!");
    }
}
EOF

cd ..
```

### Minute 3: Configure Distributed Build

Create a simple configuration file:

```bash
cat > sample-project/.distributed-gradle.yml << 'EOF'
project:
  name: "sample-project"
  type: "java"

distribution:
  workers:
    min: 2
    max: 4
  
  strategy: "round-robin"
  
  caching:
    enabled: true
    strategy: "full"
EOF
```

### Minute 4: Run Your First Distributed Build

Now for the exciting part - run your first distributed build:

```bash
# Run the distributed build
./sync_and_build.sh --project sample-project --workers 4
```

You should see output similar to this:

```
🚀 Starting Distributed Gradle Build
📊 Configuration loaded: 4 workers, full caching
🔄 Connecting to workers: [worker-1, worker-2, worker-3, worker-4]
✅ All workers connected and ready
📦 Analyzing project: sample-project
🎯 Distributing tasks across 4 workers
⚡ Building with distributed processing...
🏁 Build completed successfully!
📈 Performance: 75% faster than traditional build
```

### Minute 5: Compare Performance

Let's compare with a traditional build to see the improvement:

```bash
# Traditional build (for comparison)
cd sample-project
time gradle build
cd ..

# Distributed build (already done above)
./sync_and_build.sh --project sample-project --workers 4 --timing-only
```

You should see a significant performance improvement, even with this simple project.

## 🎉 Congratulations!

You've successfully set up and run your first distributed Gradle build in just 5 minutes! 

## 📊 What Just Happened?

Behind the scenes, the system:

1. **Analyzed your project** to identify tasks that can be parallelized
2. **Distributed tasks** across 4 worker processes
3. **Managed dependencies** to avoid redundant work
4. **Collected results** from all workers and assembled the final build
5. **Cached artifacts** for even faster future builds

## 🎯 Next Steps

Now that you have a working distributed build setup:

1. **Try it with your real project**
   ```bash
   ./sync_and_build.sh --project /path/to/your/project
   ```

2. **Explore configuration options**
   ```bash
   nano sample-project/.distributed-gradle.yml
   ```

3. **Check out more examples**
   ```bash
   ./go/client/example/main.go --show-examples
   ```

## 🆘 Troubleshooting

### Common Quick Issues

**"Permission denied" errors:**
```bash
chmod +x setup_quick_start.sh
chmod +x sync_and_build.sh
```

**"Java not found" errors:**
```bash
# Install Java (Ubuntu/Debian)
sudo apt-get update
sudo apt-get install openjdk-11-jdk

# Install Java (macOS)
brew install openjdk@11
```

**"Worker connection failed":**
```bash
# Check if workers are running
./scripts/health_checker.sh --status
```

### Get Help

- **Documentation**: [Complete Setup Guide](/docs/setup-guide)
- **Video Tutorial**: [Beginner Course](/video-courses/beginner)
- **Community**: [GitHub Discussions](https://github.com/milos85vasic/Distributed-Gradle-Building/discussions)
- **Issues**: [Report a Problem](https://github.com/milos85vasic/Distributed-Gradle-Building/issues)

## 🚀 Ready for More?

You've only scratched the surface! Here's what you can explore next:

- **Advanced Configuration**: Fine-tune worker allocation and caching
- **CI/CD Integration**: Integrate with your existing build pipeline
- **Production Deployment**: Set up high-availability distributed builds
- **Performance Optimization**: Master advanced optimization techniques

**Continue Learning:** [User Guide](/docs/user-guide)

---

**🎊 You did it!** You're now ready to revolutionize your Gradle builds. Want to share your experience? [Post in our community](https://github.com/milos85vasic/Distributed-Gradle-Building/discussions/categories/show-and-tell)!