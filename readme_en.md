# Go Benchmark — Open-source Website Load Testing

中文说明见 [README.md](./README.md)。

Repo: [github.com/baiwapak/go-benchmark](https://github.com/baiwapak/go-benchmark)

## Introduction

Go Benchmark is an open-source website load-testing tool written in Go.
Enter a URL in the browser, configure concurrency and duration, watch live
progress and charts, then download a bilingual report (with suggestions) as
PDF/JSON.

It is a lightweight Go + Gin web service without a database: jobs stay in
memory; reports are stored as JSON files.

## Features

- HTTP load test by URL (GET/POST/PUT/PATCH/DELETE/HEAD; concurrency, duration, timeout, ramp-up, headers, body)
- Presets from target host specs (CPU / RAM / bandwidth → suggested concurrency, duration, timeout)
- SSE live progress plus RPS / latency / success-rate charts
- Report: throughput, latency percentiles (p50/p90/p95/p99), status/error histograms
- Rule-based optimization suggestions (follow UI language)
- History: delete one report, clear all, compare two reports
- Glossary of metrics (RPS, percentiles, ramp-up, and more)
- PDF / JSON download
- Chinese / English UI (default: Chinese)
- Guardrails: authorization checkbox, SSRF (blocks loopback/private targets), create-job rate limit per IP
- Auto-detect local CPU/memory for concurrency and job caps (or set them in `.env`)

## Quick Start

```powershell
copy .env.example .env
go run ./cmd/server/
```

Open http://127.0.0.1:8000/

Only test targets you own or have written permission to test.
`localhost` and private-network addresses are rejected.

You can also create jobs via JSON API: `POST /api/v1/bench`.

## Configuration (`.env`)

If `MAX_CONCURRENCY`, `MAX_DURATION_SEC`, and `MAX_INFLIGHT_JOBS` are omitted,
limits are inferred from this machine's CPU and available memory.

| Key | Default | Description |
|-----|---------|-------------|
| `SERVER_ADDR` | `:8000` | Listen address |
| `SERVER_MODE` | `debug` | Gin mode (`debug` / `release`) |
| `DEFAULT_LANG` | `zh` | Default UI language (`zh` / `en`) |
| `REPORT_DIR` | `data/reports` | Report JSON directory |
| `MAX_CONCURRENCY` | auto | Max workers per job |
| `MAX_DURATION_SEC` | auto | Max test duration (seconds) |
| `MAX_TIMEOUT_SEC` | `60` | Max per-request timeout (seconds) |
| `MAX_INFLIGHT_JOBS` | auto | Max concurrent jobs |
| `CREATE_RATE_LIMIT_PER_MIN` | `10` | Create-job rate limit per IP |
| `MAX_RAMP_SEC` | `120` | Max ramp-up seconds |
| `MAX_BODY_BYTES` | `65536` | Max request body size |
| `GITHUB_REPO_URL` | this repo | Floating Star button URL |
| `CORS_ALLOW_ORIGINS` | `*` in debug | CORS origins (comma-separated; empty in release = same-origin only) |

## Presets

On the home page you can fill in reference parameters from a target host profile
(starting points only — tune for page size, HTTPS, database, and so on):

| Profile | Specs | Concurrency | Duration | Timeout |
|---------|-------|-------------|----------|---------|
| Starter | 1 CPU / 1GB / 1Mbps | 5 | 30s | 15s |
| Basic | 2 CPU / 2GB / 5Mbps | 10 | 45s | 12s |
| Standard | 4 CPU / 4GB / 10Mbps | 20 | 60s | 10s |
| Growth | 4 CPU / 8GB / 20Mbps | 35 | 60s | 10s |
| Enterprise | 8 CPU / 16GB / 50Mbps | 60 | 90s | 10s |
| Performance | 16 CPU / 32GB / 100Mbps | 100 | 120s | 8s |

## PDF fonts

PDF generation tries system CJK fonts (e.g. Windows `simhei.ttf`; on Linux install `fonts-noto-cjk`).
You may also place `NotoSansSC-Regular.ttf` / `.otf` under `fonts/`.

## Disclaimer

Only test targets you own or have written permission to test.
Unauthorized load testing may be illegal.

## License

MIT License — see [LICENSE](./LICENSE).
