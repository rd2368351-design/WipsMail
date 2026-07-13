# Wispmail

Distributed, horizontally scalable email server for SaaS providers.

## Features

SMTP, IMAP, POP3, JMAP | Multi-tenant | DKIM, SPF, DMARC | Anti-spam, Anti-virus
S3, PostgreSQL, Redis, Elasticsearch | Webhooks, Billing | Raft, Cluster, Plugin
OpenTelemetry, Prometheus, Grafana | gRPC, GraphQL, REST

## Quick Start

cp .env.example .env
docker compose -f build/docker-compose.yml up -d
make dev

API: http://localhost:8080 | SMTP: localhost:2525 | Metrics: http://localhost:8080/metrics

## License

AGPL-3.0. Copyright (C) 2026 Wispmail Contributors.