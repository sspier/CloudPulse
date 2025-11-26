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
| Local Emulation        | **LocalStack**                                    | Run AWS services locally for dev/test                  |

---

## API Documentation

Create a new target to monitor:

```bash
POST /targets

Request:

```json
{
  "name": "Example",
  "url": "https://example.com"
}

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
GET /targets
```

Return the latest probe result for each target:

```bash
GET /results
```

Return the full probe history for the given target ID:

```bash
GET /results/{id}
```



## Status

- In active development.

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
- Pause for documentation ([#24](../../pull/22))

## Changes Tracking

1. Added repo structure and go modules
2. Added Components -> API (initial) describing /health and /metrics.
3. Local Docker Compose with Prometheus
4. Terraform backend (S3)
5. VPC infrastructure
6. ECS Fargate API + ALB
7. DynamoDB results table
8. Runner (EventBridge schedule)
9. CloudWatch alarms & dashboards
10. CI/CD (GitHub Actions)
11. Recurring uptime checks and per-target result history
12. Pause for documentation
