pipeline {
    agent any

    environment {
        GO_VERSION = '1.21'
        DOCKER_IMAGE_TAG = "${env.BUILD_NUMBER}"
        REGISTRY = 'your-registry.com'
        APP_NAME = 'distributed-gradle-building'
    }

    stages {
        stage('Checkout') {
            steps {
                checkout scm
            }
        }

        stage('Setup Go Environment') {
            steps {
                script {
                    // Install Go if not available
                    sh '''
                        if ! command -v go &> /dev/null; then
                            echo "Installing Go ${GO_VERSION}..."
                            wget -q https://golang.org/dl/go${GO_VERSION}.linux-amd64.tar.gz
                            sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
                            export PATH=$PATH:/usr/local/go/bin
                            echo "export PATH=$PATH:/usr/local/go/bin" >> ~/.bashrc
                        fi
                        go version
                    '''
                }
            }
        }

        stage('Dependencies') {
            steps {
                dir('go') {
                    sh 'go mod download'
                    sh 'go mod tidy'
                }
            }
        }

        stage('Unit Tests') {
            steps {
                dir('go') {
                    sh '''
                        go test ./coordinatorpkg ./workerpkg ./cachepkg ./monitorpkg -v -race -coverprofile=coverage.out -covermode=atomic
                        go tool cover -func=coverage.out
                    '''
                }
            }
            post {
                always {
                    dir('go') {
                        publishCoverage adapters: [coberturaAdapter('coverage.out')]
                    }
                }
            }
        }

        stage('Integration Tests') {
            steps {
                dir('go') {
                    sh '''
                        go test ./tests/integration -v -timeout 60s -coverprofile=integration-coverage.out
                        go tool cover -func=integration-coverage.out
                    '''
                }
            }
        }

        stage('Security Tests') {
            steps {
                dir('go') {
                    sh '''
                        go test ./tests/security -v -timeout 30s -coverprofile=security-coverage.out
                        # Run gosec security scanner
                        if command -v gosec &> /dev/null; then
                            gosec ./...
                        fi
                    '''
                }
            }
        }

        stage('Build Binaries') {
            steps {
                dir('go') {
                    sh '''
                        mkdir -p bin

                        echo "Building coordinator..."
                        go build -ldflags="-X main.version=${BUILD_NUMBER}" -o bin/coordinator main.go

                        echo "Building worker..."
                        go build -ldflags="-X main.version=${BUILD_NUMBER}" -o bin/worker worker.go

                        echo "Building cache server..."
                        go build -ldflags="-X main.version=${BUILD_NUMBER}" -o bin/cache_server cache_server.go

                        echo "Building monitor..."
                        go build -ldflags="-X main.version=${BUILD_NUMBER}" -o bin/monitor monitor.go

                        echo "Build completed successfully"
                        ls -la bin/
                    '''
                }
            }
        }

        stage('Build Docker Images') {
            steps {
                script {
                    // Build coordinator image
                    sh """
                        docker build -t ${REGISTRY}/${APP_NAME}-coordinator:${DOCKER_IMAGE_TAG} \\
                            --build-arg BINARY_PATH=go/bin/coordinator \\
                            -f docker/Dockerfile.coordinator .
                    """

                    // Build worker image
                    sh """
                        docker build -t ${REGISTRY}/${APP_NAME}-worker:${DOCKER_IMAGE_TAG} \\
                            --build-arg BINARY_PATH=go/bin/worker \\
                            -f docker/Dockerfile.worker .
                    """

                    // Build cache server image
                    sh """
                        docker build -t ${REGISTRY}/${APP_NAME}-cache:${DOCKER_IMAGE_TAG} \\
                            --build-arg BINARY_PATH=go/bin/cache_server \\
                            -f docker/Dockerfile.cache .
                    """

                    // Build monitor image
                    sh """
                        docker build -t ${REGISTRY}/${APP_NAME}-monitor:${DOCKER_IMAGE_TAG} \\
                            --build-arg BINARY_PATH=go/bin/monitor \\
                            -f docker/Dockerfile.monitor .
                    """
                }
            }
        }

        stage('Push Docker Images') {
            when {
                anyOf {
                    branch 'main'
                    branch 'master'
                    tag '*'
                }
            }
            steps {
                script {
                    sh """
                        docker push ${REGISTRY}/${APP_NAME}-coordinator:${DOCKER_IMAGE_TAG}
                        docker push ${REGISTRY}/${APP_NAME}-worker:${DOCKER_IMAGE_TAG}
                        docker push ${REGISTRY}/${APP_NAME}-cache:${DOCKER_IMAGE_TAG}
                        docker push ${REGISTRY}/${APP_NAME}-monitor:${DOCKER_IMAGE_TAG}

                        # Tag as latest for main branch
                        if [ "${env.BRANCH_NAME}" = "main" ] || [ "${env.BRANCH_NAME}" = "master" ]; then
                            docker tag ${REGISTRY}/${APP_NAME}-coordinator:${DOCKER_IMAGE_TAG} ${REGISTRY}/${APP_NAME}-coordinator:latest
                            docker tag ${REGISTRY}/${APP_NAME}-worker:${DOCKER_IMAGE_TAG} ${REGISTRY}/${APP_NAME}-worker:latest
                            docker tag ${REGISTRY}/${APP_NAME}-cache:${DOCKER_IMAGE_TAG} ${REGISTRY}/${APP_NAME}-cache:latest
                            docker tag ${REGISTRY}/${APP_NAME}-monitor:${DOCKER_IMAGE_TAG} ${REGISTRY}/${APP_NAME}-monitor:latest

                            docker push ${REGISTRY}/${APP_NAME}-coordinator:latest
                            docker push ${REGISTRY}/${APP_NAME}-worker:latest
                            docker push ${REGISTRY}/${APP_NAME}-cache:latest
                            docker push ${REGISTRY}/${APP_NAME}-monitor:latest
                        fi
                    """
                }
            }
        }

        stage('Performance Tests') {
            when {
                anyOf {
                    branch 'main'
                    branch 'master'
                    tag '*'
                }
            }
            steps {
                dir('go') {
                    sh '''
                        go test ./tests/performance -v -timeout 120s -bench=. -run=^$ -count=3 | tee performance-bench.txt
                    '''
                }
            }
            post {
                always {
                    archiveArtifacts artifacts: 'go/performance-bench.txt', fingerprint: true
                }
            }
        }

        stage('Load Tests') {
            when {
                anyOf {
                    branch 'main'
                    branch 'master'
                    tag '*'
                }
            }
            steps {
                dir('go') {
                    sh '''
                        go test ./tests/load -v -timeout 120s
                    '''
                }
            }
        }

        stage('Deploy to Staging') {
            when {
                branch 'develop'
            }
            steps {
                script {
                    sh '''
                        echo "Deploying to staging environment..."
                        # Add your staging deployment commands here
                        # Example: kubectl apply -f k8s/staging/ or docker-compose up -d
                    '''
                }
            }
        }

        stage('Deploy to Production') {
            when {
                anyOf {
                    branch 'main'
                    branch 'master'
                    tag '*'
                }
            }
            steps {
                script {
                    timeout(time: 15, unit: 'MINUTES') {
                        input message: 'Deploy to production?', ok: 'Deploy'
                    }

                    sh '''
                        echo "Deploying to production environment..."
                        # Add your production deployment commands here
                        # Example: kubectl apply -f k8s/production/ or helm upgrade
                    '''
                }
            }
        }

        stage('Archive Artifacts') {
            steps {
                archiveArtifacts artifacts: 'go/bin/*', fingerprint: true
                archiveArtifacts artifacts: 'docker-compose*.yml', fingerprint: true
            }
        }
    }

    post {
        always {
            // Clean up workspace
            cleanWs()

            // Send notifications
            script {
                def buildStatus = currentBuild.currentResult
                def subject = "${env.JOB_NAME} - Build #${env.BUILD_NUMBER} - ${buildStatus}"
                def body = """
                    Build Status: ${buildStatus}
                    Build URL: ${env.BUILD_URL}
                    Branch: ${env.BRANCH_NAME}
                    Commit: ${env.GIT_COMMIT}
                """

                emailext subject: subject,
                         body: body,
                         to: 'dev-team@company.com',
                         attachLog: true
            }
        }

        success {
            echo 'Pipeline completed successfully!'
        }

        failure {
            echo 'Pipeline failed!'
        }
    }
}