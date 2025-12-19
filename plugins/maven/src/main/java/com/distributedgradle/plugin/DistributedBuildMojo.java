package com.distributedgradle.plugin;

import okhttp3.*;
import com.google.gson.Gson;
import org.apache.maven.plugin.AbstractMojo;
import org.apache.maven.plugin.MojoExecutionException;
import org.apache.maven.plugin.MojoFailureException;
import org.apache.maven.plugins.annotations.Mojo;
import org.apache.maven.plugins.annotations.Parameter;
import org.apache.maven.project.MavenProject;

import java.io.IOException;
import java.util.HashMap;
import java.util.Map;
import java.util.concurrent.TimeUnit;

/**
 * Maven plugin for distributed Gradle building
 */
@Mojo(name = "distributed-build", requiresProject = true)
public class DistributedBuildMojo extends AbstractMojo {

    /**
     * The Maven project
     */
    @Parameter(defaultValue = "${project}", readonly = true, required = true)
    private MavenProject project;

    /**
     * URL of the distributed build service
     */
    @Parameter(property = "distributed.serviceUrl", defaultValue = "http://localhost:8080")
    private String serviceUrl;

    /**
     * Authentication token for the service
     */
    @Parameter(property = "distributed.authToken")
    private String authToken;

    /**
     * Task name to execute (defaults to package)
     */
    @Parameter(property = "distributed.taskName", defaultValue = "package")
    private String taskName;

    /**
     * Whether to enable caching
     */
    @Parameter(property = "distributed.cacheEnabled", defaultValue = "true")
    private boolean cacheEnabled;

    /**
     * Timeout in minutes for build completion
     */
    @Parameter(property = "distributed.timeoutMinutes", defaultValue = "30")
    private int timeoutMinutes;

    @Override
    public void execute() throws MojoExecutionException, MojoFailureException {
        getLog().info("Submitting distributed build to: " + serviceUrl);
        getLog().info("Task: " + taskName + ", Cache enabled: " + cacheEnabled);

        OkHttpClient client = new OkHttpClient.Builder()
            .connectTimeout(30, TimeUnit.SECONDS)
            .readTimeout(timeoutMinutes, TimeUnit.MINUTES)
            .build();

        // Prepare build request
        Map<String, Object> buildRequest = new HashMap<>();
        buildRequest.put("project_path", project.getBasedir().getAbsolutePath());
        buildRequest.put("task_name", taskName);
        buildRequest.put("cache_enabled", cacheEnabled);
        buildRequest.put("build_options", new HashMap<String, String>());

        Gson gson = new Gson();
        String jsonRequest = gson.toJson(buildRequest);

        RequestBody body = RequestBody.create(jsonRequest, MediaType.parse("application/json"));

        Request.Builder requestBuilder = new Request.Builder()
            .url(serviceUrl + "/api/build")
            .post(body);

        if (authToken != null && !authToken.isEmpty()) {
            requestBuilder.addHeader("X-Auth-Token", authToken);
        }

        Request request = requestBuilder.build();

        try (Response response = client.newCall(request).execute()) {
            if (!response.isSuccessful()) {
                throw new MojoExecutionException("Build submission failed: " + response.code() + " " + response.message());
            }

            String responseBody = response.body().string();
            BuildResponse buildResponse = gson.fromJson(responseBody, BuildResponse.class);

            getLog().info("Build submitted successfully. Build ID: " + buildResponse.buildId);

            // Wait for build completion
            waitForBuildCompletion(client, buildResponse.buildId);

        } catch (IOException e) {
            throw new MojoExecutionException("Failed to submit or monitor build", e);
        }
    }

    private void waitForBuildCompletion(OkHttpClient client, String buildId) throws MojoExecutionException, IOException {
        long startTime = System.currentTimeMillis();
        long timeoutMillis = timeoutMinutes * 60 * 1000;

        while (System.currentTimeMillis() - startTime < timeoutMillis) {
            try {
                Thread.sleep(5000); // Check every 5 seconds
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                throw new MojoExecutionException("Build monitoring interrupted", e);
            }

            Request.Builder statusRequestBuilder = new Request.Builder()
                .url(serviceUrl + "/api/build/" + buildId)
                .get();

            if (authToken != null && !authToken.isEmpty()) {
                statusRequestBuilder.addHeader("X-Auth-Token", authToken);
            }

            Request statusRequest = statusRequestBuilder.build();

            try (Response statusResponse = client.newCall(statusRequest).execute()) {
                if (!statusResponse.isSuccessful()) {
                    getLog().warn("Failed to get build status: " + statusResponse.message());
                    continue;
                }

                String statusBody = statusResponse.body().string();
                Gson gson = new Gson();
                BuildStatus status = gson.fromJson(statusBody, BuildStatus.class);

                getLog().info("Build status: " + status.status + " (Duration: " + status.duration + "ms)");

                if ("completed".equals(status.status) || "failed".equals(status.status)) {
                    if ("completed".equals(status.status)) {
                        getLog().info("Build completed successfully!");
                        if (status.artifacts != null && !status.artifacts.isEmpty()) {
                            getLog().info("Artifacts: " + status.artifacts);
                        }
                    } else {
                        throw new MojoFailureException("Build failed: " + status.errorMessage);
                    }
                    return;
                }
            }
        }

        throw new MojoExecutionException("Build timeout after " + timeoutMinutes + " minutes");
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