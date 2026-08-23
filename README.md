# Go Benchmark — 开源网站压测工具

English: [readme_en.md](./readme_en.md)

仓库：[github.com/baiwapak/go-benchmark](https://github.com/baiwapak/go-benchmark)

## 简介

Go Benchmark 是用 Go 编写的开源网站压测软件。通过浏览器输入目标 URL，配置并发与时长后发起压测，实时查看进度与曲线，并生成中英文压测报告、优化建议与 PDF 下载。

基于 Go + Gin 的轻量 Web 服务：任务状态保存在内存，报告以 JSON 落盘，无需数据库。

## 主要功能

- 输入网址发起 HTTP 压测（GET/POST/PUT/PATCH/DELETE/HEAD；可配并发、时长、超时、升压、自定义 Header 与请求体）
- 按目标主机规格选择预设档位（CPU / 内存 / 带宽 → 参考并发、时长、超时）
- SSE 实时进度 + RPS / 延迟 / 成功率曲线
- 压测报告：吞吐、延迟分位（p50/p90/p95/p99）、状态码与错误分布
- 基于阈值的优化建议（随界面语言切换）
- 历史报告列表：删除单条、清空全部、两份报告对比
- 指标名词解释页（RPS、分位延迟、升压等）
- 开源下载 PDF / JSON 报告
- 中英文界面，默认中文
- 安全护栏：授权确认、SSRF（拒绝回环/私网等地址）、创建任务 IP 限流
- 本机 CPU/内存自动推算并发与任务上限（也可在 `.env` 中写死）

## 快速开始

```powershell
copy .env.example .env
go run ./cmd/server/
```

打开浏览器：http://127.0.0.1:8000/

勾选授权确认后提交目标 URL（请仅测试您拥有或已获授权的站点）。`localhost` 与内网地址会被拒绝。

也可通过 JSON API 创建任务：`POST /api/v1/bench`。

## 常用配置（`.env`）

未设置 `MAX_CONCURRENCY` / `MAX_DURATION_SEC` / `MAX_INFLIGHT_JOBS` 时，会按本机 CPU 与可用内存自动推算。

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `SERVER_ADDR` | `:8000` | 监听地址 |
| `SERVER_MODE` | `debug` | Gin 模式（`debug` / `release`） |
| `DEFAULT_LANG` | `zh` | 默认语言（`zh` / `en`） |
| `REPORT_DIR` | `data/reports` | 报告 JSON 目录 |
| `MAX_CONCURRENCY` | 自动 | 单任务最大并发 |
| `MAX_DURATION_SEC` | 自动 | 最大压测时长（秒） |
| `MAX_TIMEOUT_SEC` | `60` | 单请求最大超时（秒） |
| `MAX_INFLIGHT_JOBS` | 自动 | 同时运行任务上限 |
| `CREATE_RATE_LIMIT_PER_MIN` | `10` | 每 IP 每分钟创建任务上限 |
| `MAX_RAMP_SEC` | `120` | 升压最长秒数 |
| `MAX_BODY_BYTES` | `65536` | 请求体最大字节数 |
| `GITHUB_REPO_URL` | 本仓库地址 | 页面浮动 Star 按钮链接 |
| `CORS_ALLOW_ORIGINS` | debug 下为 `*` | 跨域来源（逗号分隔；release 且为空则仅同源） |

## 压测预设

首页可按目标站点规格一键填入参考参数（仅为起点，需按页面大小、HTTPS、数据库等再微调）：

| 档位 | 规格 | 并发 | 时长 | 超时 |
|------|------|------|------|------|
| 入门主机 | 1 核 / 1GB / 1Mbps | 5 | 30s | 15s |
| 轻量云主机 | 2 核 / 2GB / 5Mbps | 10 | 45s | 12s |
| 标准云主机 | 4 核 / 4GB / 10Mbps | 20 | 60s | 10s |
| 成长型服务 | 4 核 / 8GB / 20Mbps | 35 | 60s | 10s |
| 企业级主机 | 8 核 / 16GB / 50Mbps | 60 | 90s | 10s |
| 高性能主机 | 16 核 / 32GB / 100Mbps | 100 | 120s | 8s |

## PDF 中文说明

PDF 会尝试加载系统中文字体（Windows 常见：`simhei.ttf`；Linux 可安装 `fonts-noto-cjk`）。  
也可将 `NotoSansSC-Regular.ttf` / `.otf` 放到项目 `fonts/` 目录。

## 免责声明

请仅对您拥有或已获书面授权的目标进行压测。未经授权的压力测试可能违法，开发者不对滥用行为负责。

## 许可证

MIT License（见 [LICENSE](./LICENSE)）
