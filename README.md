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

## Local Development

You can run CloudPulse locally in two modes.

### 1. Standalone (In-Memory)

Simplest way to test the API logic. Data is lost on restart.

```bash
go run apps/api/main.go
# API listens on localhost:8080
```

### 2. Integrated (Local DynamoDB)

Test the full flow with persistence and the Runner service.

**Step 1: Start DynamoDB**

```bash
docker-compose up -d
# Starts DynamoDB Local and creates 'cloudpulse-targets-local' and 'cloudpulse-probe-results-local' tables
```

**Step 2: Start the API**

```bash
export TABLE_NAME_TARGETS=cloudpulse-targets-local
export TABLE_NAME_RESULTS=cloudpulse-probe-results-local
export AWS_ENDPOINT=http://localhost:8000
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=dummy
export AWS_SECRET_ACCESS_KEY=dummy

go run apps/api/main.go
```

**Step 3: Run the Probe Runner**

(In a separate terminal)

```bash
# Same env vars as above
export TABLE_NAME_TARGETS=cloudpulse-targets-local
export TABLE_NAME_RESULTS=cloudpulse-probe-results-local
export AWS_ENDPOINT=http://localhost:8000
export AWS_REGION=us-east-1
export AWS_ACCESS_KEY_ID=dummy
export AWS_SECRET_ACCESS_KEY=dummy

go run apps/runner/main.go
```

---

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

## Deploy with Helm

CloudPulse ships with a Helm chart for deploying the API to Kubernetes.

### Prerequisites

- A running Kubernetes cluster (Rancher Desktop, kind, k3d, EKS, etc.)
- `kubectl` configured to talk to the cluster
- `helm` installed
- Docker (or compatible runtime) for building the image

---

### Build the API image

From the repo root:

```bash
docker build -f apps/api/Dockerfile -t cloudpulse-api:dev .
```

### Install (or upgrade) the Helm release

```bash
helm upgrade --install cloudpulse-api deployments/helm/cloudpulse-api -n cloudpulse --create-namespace
```

### Check that everything is deployed

```bash
helm list -n cloudpulse
kubectl -n cloudpulse get deploy,svc,pods
```

You should see:

- A Deployment for the API
- A Service exposing it on port 80
- Pods in Running state

### Test the API

Port-forward the Service:

```bash
kubectl -n cloudpulse port-forward svc/cloudpulse-api 8080:80
```

In another terminal:

```bash
curl http://localhost:8080/health
```

Expected response:

```bash
ok
```

## Development Timeline

- Bootstrap repo (Go modules, folder layout) ([#2](../../pull/2))
- Minimal API (/health, /metrics) ([#4](../../pull/4))
- Local Docker Compose with Prometheus ([#6](../../pull/6))
- Terraform backend (S3) ([#8](../../pull/8))
- VPC infrastructure ([#10](../../pull/10))
- ECS Fargate API + ALB ([#12](../../pull/12))
- DynamoDB results table ([#14](../../pull/14))
- Runner (EventBridge schedule) ([#16](../../pull/16))
- CloudWatch alarms & dashboards ([#18](../../pull/18))
- CI/CD (GitHub Actions) ([#20](../../pull/20))
- Recurring uptime checks and per-target result history ([#22](../../pull/22))
- Pause for documentation ([#24](../../pull/24))
- Kuburnetes deployment and service ([#26](../../pull/26))
- Helm chart update ([#28](../../pull/28))
- DynamoDB Targets & Runner Service ([#30](../../pull/30))
