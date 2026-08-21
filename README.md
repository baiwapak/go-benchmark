# Go Benchmark — 开源网站压测工具

English: [readme_en.md](./readme_en.md)

## 简介

Go Benchmark 是用 Go 编写的开源网站压测软件。通过浏览器输入目标 URL，配置并发与时长后发起压测，实时查看进度，并生成中英文压测报告、优化建议与 PDF 下载。

基于 Go + Gin 的轻量 Web 服务：任务状态保存在内存，报告以 JSON 落盘，无需数据库。

## 主要功能

- 输入网址发起 HTTP 压测（可配方法、并发、时长、超时、自定义 Header）
- SSE 实时进度（请求数、RPS、成功率、平均延迟）
- 压测报告：吞吐、延迟分位（p50/p90/p95/p99）、状态码与错误分布
- 基于阈值的优化建议（随界面语言切换）
- 开源下载 PDF / JSON 报告
- 中英文界面，默认中文

## 快速开始

```powershell
copy .env.example .env
go run ./cmd/server/
```

打开浏览器：http://127.0.0.1:8000/

勾选授权确认后提交目标 URL（请仅测试您拥有或已获授权的站点）。

## 常用配置（`.env`）

| 配置项 | 默认值 | 说明 |
|--------|--------|------|
| `SERVER_ADDR` | `:8000` | 监听地址 |
| `DEFAULT_LANG` | `zh` | 默认语言（`zh` / `en`） |
| `REPORT_DIR` | `data/reports` | 报告 JSON 目录 |
| `MAX_CONCURRENCY` | `200` | 单任务最大并发 |
| `MAX_DURATION_SEC` | `300` | 最大压测时长（秒） |
| `MAX_TIMEOUT_SEC` | `60` | 单请求最大超时（秒） |
| `MAX_INFLIGHT_JOBS` | `3` | 同时运行任务上限 |
| `CREATE_RATE_LIMIT_PER_MIN` | `10` | 每 IP 每分钟创建任务上限 |

## PDF 中文说明

PDF 会尝试加载系统中文字体（Windows 常见：`simhei.ttf`；Linux 可安装 `fonts-noto-cjk`）。  
也可将 `NotoSansSC-Regular.ttf` / `.otf` 放到项目 `fonts/` 目录。

## 免责声明

请仅对您拥有或已获书面授权的目标进行压测。未经授权的压力测试可能违法，开发者不对滥用行为负责。

## 许可证

MIT License（见 [LICENSE](./LICENSE)）
