# CloudPulse
**A Cloud-Native Uptime & SLO Monitoring Platform built with Go and Terraform**

## Overview
CloudPulse is a production-style end-to-end cloud engineering project.

It provisions all infrastructure with **Terraform** and runs lightweight **Go** services on **AWS ECS Fargate**, exposing metrics, API endpoints, and dashboards for uptime and latency tracking.

It periodically probes registered endpoints (HTTP/TCP) to measure uptime, response latency, and error rates.  

All infrastructure, from networking to monitoring, is defined as code — allowing reproducible environments and automated deployments.


## Tech Stack

| Category | Technology | Purpose |
|-----------|-------------|----------|
| Language | **Go** | High-performance API + runner services |
| Infrastructure as Code | **Terraform** | AWS provisioning (VPC, ECS, DDB, SSM, CloudWatch) |
| Cloud Platform | **AWS** | ECS Fargate, Lambda/EventBridge, API Gateway, DynamoDB |
| CI/CD | **GitHub Actions** | Build, test, lint, Terraform plan/apply, ECR deploy |
| Metrics & Logs | **Prometheus**, **CloudWatch**, **OpenTelemetry** | Uptime metrics, traces, alerts |
| Containerization | **Docker** | Local dev + AWS deployment packaging |
| Local Emulation | **LocalStack** | Run AWS services locally for dev/test |
---

## Status
- In active development. 

## Development Timeline
- 0: Bootstrap repo (Go modules, folder layout)    (#0)
- 1: Local Docker Compose with Prometheus          (#1)
- 2: Minimal API (/health, /metrics)               (#2)
- 3: Terraform backend (S3 + DynamoDB lock)        (#3)
- 4: VPC module                                    (#4)
- 5: ECS Fargate API + ALB                         (#5)
- 6: DynamoDB results table                        (#6) 
- 7: Runner (EventBridge schedule)                 (#7)
- 8: CloudWatch alarms & dashboards                (#8)
- 9: CI/CD (GitHub Actions)                        (#9)




