package com.distributedgradle.plugin;

import okhttp3.*;
import com.google.gson.Gson;
import org.gradle.api.Plugin;
import org.gradle.api.Project;
import org.gradle.api.tasks.TaskProvider;
import org.gradle.api.DefaultTask;
import org.gradle.api.tasks.TaskAction;
import org.gradle.api.provider.Property;

import java.io.IOException;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.TimeUnit;

public class DistributedGradlePlugin implements Plugin<Project> {

    @Override
    public void apply(Project project) {
        // Create extension for configuration
        DistributedGradleExtension extension = project.getExtensions().create("distributedGradle", DistributedGradleExtension.class);

        // Register the distributed build task
        TaskProvider<DistributedBuildTask> distributedBuildTask = project.getTasks().register("distributedBuild", DistributedBuildTask.class, task -> {
            task.getServiceUrl().set(extension.getServiceUrl());
            task.getAuthToken().set(extension.getAuthToken());
            task.getTaskName().set(extension.getTaskName());
            task.getCacheEnabled().set(extension.getCacheEnabled());
            task.getTimeoutMinutes().set(extension.getTimeoutMinutes());
        });
    }

    public static class DistributedGradleExtension {
        private final Property<String> serviceUrl;
        private final Property<String> authToken;
        private final Property<String> taskName;
        private final Property<Boolean> cacheEnabled;
        private final Property<Integer> timeoutMinutes;

        public DistributedGradleExtension(Project project) {
            serviceUrl = project.getObjects().property(String.class);
            serviceUrl.set("http://localhost:8080");

            authToken = project.getObjects().property(String.class);
            taskName = project.getObjects().property(String.class);
            taskName.set("build");

            cacheEnabled = project.getObjects().property(Boolean.class);
            cacheEnabled.set(true);

            timeoutMinutes = project.getObjects().property(Integer.class);
            timeoutMinutes.set(30);
        }

        public Property<String> getServiceUrl() { return serviceUrl; }
        public Property<String> getAuthToken() { return authToken; }
        public Property<String> getTaskName() { return taskName; }
        public Property<Boolean> getCacheEnabled() { return cacheEnabled; }
        public Property<Integer> getTimeoutMinutes() { return timeoutMinutes; }
    }

    public static class DistributedBuildTask extends DefaultTask {
        private final Property<String> serviceUrl = getProject().getObjects().property(String.class);
        private final Property<String> authToken = getProject().getObjects().property(String.class);
        private final Property<String> taskName = getProject().getObjects().property(String.class);
        private final Property<Boolean> cacheEnabled = getProject().getObjects().property(Boolean.class);
        private final Property<Integer> timeoutMinutes = getProject().getObjects().property(Integer.class);

        @TaskAction
        public void executeDistributedBuild() throws IOException {
            String url = serviceUrl.get();
            String token = authToken.getOrNull();
            String task = taskName.get();
            boolean cache = cacheEnabled.get();
            int timeout = timeoutMinutes.get();

            getLogger().info("Submitting distributed build to: {}", url);
            getLogger().info("Task: {}, Cache enabled: {}", task, cache);

            OkHttpClient client = new OkHttpClient.Builder()
                .connectTimeout(30, TimeUnit.SECONDS)
                .readTimeout(timeout, TimeUnit.MINUTES)
                .build();

            // Prepare build request
            Map<String, Object> buildRequest = new HashMap<>();
            buildRequest.put("project_path", getProject().getProjectDir().getAbsolutePath());
            buildRequest.put("task_name", task);
            buildRequest.put("cache_enabled", cache);
            buildRequest.put("build_options", new HashMap<String, String>());

            Gson gson = new Gson();
            String jsonRequest = gson.toJson(buildRequest);

            RequestBody body = RequestBody.create(jsonRequest, MediaType.parse("application/json"));

            Request.Builder requestBuilder = new Request.Builder()
                .url(url + "/api/build")
                .post(body);

            if (token != null && !token.isEmpty()) {
                requestBuilder.addHeader("X-Auth-Token", token);
            }

            Request request = requestBuilder.build();

            try (Response response = client.newCall(request).execute()) {
                if (!response.isSuccessful()) {
                    throw new IOException("Build submission failed: " + response.code() + " " + response.message());
                }

                String responseBody = response.body().string();
                BuildResponse buildResponse = gson.fromJson(responseBody, BuildResponse.class);

                getLogger().info("Build submitted successfully. Build ID: {}", buildResponse.buildId);

                // Wait for build completion
                waitForBuildCompletion(client, url, token, buildResponse.buildId, timeout);

            }
        }

        private void waitForBuildCompletion(OkHttpClient client, String baseUrl, String token, String buildId, int timeoutMinutes) throws IOException {
            long startTime = System.currentTimeMillis();
            long timeoutMillis = timeoutMinutes * 60 * 1000;

            while (System.currentTimeMillis() - startTime < timeoutMillis) {
                try {
                    Thread.sleep(5000); // Check every 5 seconds
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    throw new IOException("Build monitoring interrupted", e);
                }

                Request.Builder statusRequestBuilder = new Request.Builder()
                    .url(baseUrl + "/api/build/" + buildId)
                    .get();

                if (token != null && !token.isEmpty()) {
                    statusRequestBuilder.addHeader("X-Auth-Token", token);
                }

                Request statusRequest = statusRequestBuilder.build();

                try (Response statusResponse = client.newCall(statusRequest).execute()) {
                    if (!statusResponse.isSuccessful()) {
                        getLogger().warn("Failed to get build status: {}", statusResponse.message());
                        continue;
                    }

                    String statusBody = statusResponse.body().string();
                    Gson gson = new Gson();
                    BuildStatus status = gson.fromJson(statusBody, BuildStatus.class);

                    getLogger().info("Build status: {} (Duration: {}ms)", status.status, status.duration);

                    if ("completed".equals(status.status) || "failed".equals(status.status)) {
                        if ("completed".equals(status.status)) {
                            getLogger().info("Build completed successfully!");
                            if (status.artifacts != null && !status.artifacts.isEmpty()) {
                                getLogger().info("Artifacts: {}", status.artifacts);
                            }
                        } else {
                            getLogger().error("Build failed: {}", status.errorMessage);
                            throw new IOException("Build failed: " + status.errorMessage);
                        }
                        return;
                    }
                }
            }

            throw new IOException("Build timeout after " + timeoutMinutes + " minutes");
        }

        public Property<String> getServiceUrl() { return serviceUrl; }
        public Property<String> getAuthToken() { return authToken; }
        public Property<String> getTaskName() { return taskName; }
        public Property<Boolean> getCacheEnabled() { return cacheEnabled; }
        public Property<Integer> getTimeoutMinutes() { return timeoutMinutes; }
    }

    // Response classes
    public static class BuildResponse {
        public String buildId;
        public String status;
    }

    public static class BuildStatus {
        public String buildId;
        public String workerId;
        public String projectPath;
        public String taskName;
        public String status;
        public String startTime;
        public String endTime;
        public long duration;
        public boolean success;
        public double cacheHitRate;
        public java.util.List<String> artifacts;
        public String errorMessage;
    }
}