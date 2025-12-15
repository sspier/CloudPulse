# CloudPulse

**A Cloud-Native Uptime & SLO Monitoring Platform built with Go and Terraform**

## Overview

CloudPulse is a production-style end-to-end cloud engineering project.

It provisions all infrastructure with **Terraform** and runs lightweight **Go** services on **AWS ECS Fargate**, exposing metrics, API endpoints, and dashboards for uptime and latency tracking.

It periodically probes registered endpoints (HTTP/TCP) to measure uptime, response latency, and error rates.

All infrastructure, from networking to monitoring, is defined as code - allowing reproducible environments and automated deployments.

## Tech Stack

| Category               | Technology                                        | Purpose                                                |
| ---------------------- | ------------------------------------------------- | ------------------------------------------------------ |
| Language               | **Go**                                            | High-performance API + runner services                 |
| Infrastructure as Code | **Terraform**                                     | AWS provisioning (VPC, ECS, DDB, SSM, CloudWatch)      |
| Cloud Platform         | **AWS**                                           | ECS Fargate, Lambda/EventBridge, API Gateway, DynamoDB |
| CI/CD                  | **GitHub Actions**                                | Build, test, lint, Terraform plan/apply, ECR deploy    |
| Metrics & Logs         | **Prometheus**, **CloudWatch**, **OpenTelemetry** | Uptime metrics, traces, alerts                         |
| Containerization       | **Docker**                                        | Local dev + AWS deployment packaging                   |
| Local Emulation        | **DynamoDB Local**                                | Run DynamoDB locally for dev/test                      |

---

## Architecture Overview

CloudPulse uses a split-service architecture for robustness and scalability:

1.  **API Service**: Handles user requests (add target, view results) and serves as the front door.
    - **Mode 1: In-Memory (Standalone)**: Stores data in RAM.
    - **Mode 2: Cloud (Distributed)**: Stores data in DynamoDB.
2.  **Runner System**: Polls targets and records results.
    - **Mode 1: Internal Ticker**: Runs as a background goroutine within the API service (Standalone mode).
    - **Mode 2: Lambda Runner**: Runs as a decoupled AWS Lambda function triggered by EventBridge (Cloud mode).

---

## Kubernetes Local Development

Run the entire CloudPulse stack (API, Runner, DynamoDB) on a local Kubernetes cluster.

### Prerequisites

- A running Kubernetes cluster (Rancher Desktop, kind, k3d, Docker Desktop)
- `kubectl` and `helm` installed
- `docker` installed

### 1. Build Images

Build the Docker images for the API and Runner services.
**Important**: Use the tag `v1` (or matching tag) throughout.

```bash
docker build -t cloudpulse-api:v1 -f apps/api/Dockerfile .
docker build -t cloudpulse-runner:v1 -f apps/runner/Dockerfile .
```

### 2. Infrastructure Setup

Deploy DynamoDB Local and initialize the tables.

```bash
# Create namespace
kubectl create namespace cloudpulse

# Deploy Database Infrastructure
kubectl apply -f deployments/kubernetes/dynamodb.yaml -n cloudpulse

# Initialize Tables (Job)
kubectl apply -f deployments/kubernetes/setup-tables-job.yaml -n cloudpulse
```

### 3. Deploy Applications

**API Service (via Helm)**
Connects to the local DynamoDB service on port 8000.

```bash
helm upgrade --install cloudpulse-api deployments/helm/cloudpulse-api -n cloudpulse \
  --set image.repository=cloudpulse-api \
  --set image.tag=v1 \
  --set env.AWS_ENDPOINT="http://dynamodb-local:8000" \
  --set env.TABLE_NAME_TARGETS="cloudpulse-targets-local" \
  --set env.TABLE_NAME_RESULTS="cloudpulse-probe-results-local"
```

**Runner Service (via Manifest)**
Background worker that polls targets.

```bash
kubectl apply -f deployments/kubernetes/runner.yaml -n cloudpulse
```

### 4. Monitoring (Optional)

Deploy Prometheus to scrape metrics from the API.

```bash
# Create ConfigMap and Deployment
kubectl apply -f deployments/kubernetes/prometheus-configmap.yaml -n cloudpulse
kubectl apply -f deployments/kubernetes/prometheus.yaml -n cloudpulse

# Access Prometheus Dashboard (localhost:9090)
kubectl port-forward svc/prometheus 9090:9090 -n cloudpulse
```

### 5. Verification

**Connections**
Port-forward the API Service (recommended over Pod forwarding for stability).

```bash
kubectl port-forward svc/cloudpulse-api 8080:80 -n cloudpulse
```

**Test Commands (PowerShell)**

```powershell
# Create a Target
Invoke-RestMethod -Uri "http://localhost:8080/targets" -Method Post -ContentType "application/json" -Body '{ "name": "Google", "url": "https://google.com" }'

# View Results
Invoke-RestMethod -Uri "http://localhost:8080/results" | Format-Table
```

**Test Commands (Bash/Curl)**

```bash
# Create a Target
curl -X POST http://localhost:8080/targets \
  -H "Content-Type: application/json" \
  -d '{ "name": "Google", "url": "https://google.com" }'

# View Results
curl http://localhost:8080/results
```

## API Documentation

Create a new target to monitor:

```bash
curl -v http://localhost:8080/targets -H "Content-Type: application/json" -d '{ "name": "My Blog", "url": "https://example.com"}'

or

POST http://localhost:8080/targets

Request:

{
  "name": "Example",
  "url": "https://example.com"
}
```

Response:

```json
{
  "id": "20251120184323.874139000",
  "name": "Example",
  "url": "https://example.com"
}
```

Return all targets:

```bash
curl -v http://localhost:8080/targets
or
GET http://localhost:8080/targets
```

Return the latest probe result for each target:

```bash
curl -v http://localhost:8080/results
or
GET http://localhost:8080/results
```

Return the full probe history for the given target ID:

```bash
curl -v http://localhost:8080/results/abc123
or
GET http://localhost:8080/results/abc123
```
