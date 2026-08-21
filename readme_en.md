# Go Benchmark — Open-source Website Load Testing

中文说明见 [README.md](./README.md)。

## Introduction

Go Benchmark is an open-source website load-testing tool written in Go.
Enter a URL in the browser, configure concurrency and duration, watch live
progress, then download a bilingual report (with suggestions) as PDF/JSON.

It is a lightweight Go + Gin web service without a database: jobs stay in
memory; reports are stored as JSON files.

## Features

- HTTP load test by URL (method, concurrency, duration, timeout, headers)
- SSE live progress
- Report: throughput, latency percentiles, status/error histograms
- Rule-based optimization suggestions (follow UI language)
- PDF / JSON download
- Chinese / English UI (default: Chinese)

## Quick Start

```powershell
copy .env.example .env
go run ./cmd/server/
```

Open http://127.0.0.1:8000/

Only test targets you own or have written permission to test.

## Configuration (`.env`)

| Key | Default | Description |
|-----|---------|-------------|
| `SERVER_ADDR` | `:8000` | Listen address |
| `DEFAULT_LANG` | `zh` | Default UI language (`zh` / `en`) |
| `REPORT_DIR` | `data/reports` | Report JSON directory |
| `MAX_CONCURRENCY` | `200` | Max workers per job |
| `MAX_DURATION_SEC` | `300` | Max test duration |
| `MAX_TIMEOUT_SEC` | `60` | Max per-request timeout |
| `MAX_INFLIGHT_JOBS` | `3` | Max concurrent jobs |
| `CREATE_RATE_LIMIT_PER_MIN` | `10` | Create-job rate limit per IP |

## PDF fonts

PDF generation tries system CJK fonts (e.g. Windows `simhei.ttf`; on Linux install `fonts-noto-cjk`).
You may also place `NotoSansSC-Regular.ttf` / `.otf` under `fonts/`.

## Disclaimer

Only test targets you own or have written permission to test.
Unauthorized load testing may be illegal.

## License

MIT License — see [LICENSE](./LICENSE).
