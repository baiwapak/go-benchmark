# go-benchmark V1 规划（落地文档）

> 对应实现仓库：`go-benchmark`  
> 决策确认：纯 Go SSR（1A）+ 无数据库落盘 JSON/PDF（2C）

## 已确认决策

- **交互**：纯 Go SSR（表单 → 实时进度 → 报告页），不引入 Vue
- **存储**：无数据库；任务态内存 + 报告 JSON/PDF 落盘
- **语言**：中英双语，**默认中文**；报告与建议随当前语言输出
- **文档**：根目录 `README.md`（中文）与 `readme_en.md`（英文），相互链接

## 产品范围（V1）

| 能力 | 说明 |
|------|------|
| 发起压测 | 输入目标 URL；可配并发数、持续时间、HTTP 方法、请求超时、可选 Header |
| 实时进度 | SSE 推送已发请求数、当前 RPS、成功率、平均延迟 |
| 压测报告 | 吞吐、延迟分位（p50/p90/p95/p99）、状态码分布、错误分类、耗时 |
| 优化建议 | 基于阈值的规则引擎，按报告语言输出可执行建议 |
| PDF | 开源可下载，内容与当前语言报告一致 |
| 安全护栏 | 启动页醒目授权声明；配置硬上限；创建任务 IP 限流 |

**不做（V1）**：分布式压测节点、登录/RBAC、MySQL、历史账号体系、录制回放脚本。

## 技术栈

- Go 1.22+ / Gin / godotenv
- 分层：`handler → service`；`pkg/response` + `pkg/apperr`
- 压测引擎：自研 `net/http` worker 池
- PDF：`jung-kurt/gofpdf` + 系统/本地 CJK 字体
- 默认端口：`8000`

## 目录结构

```
go-benchmark/
  README.md
  readme_en.md
  LICENSE
  docs/plan/plan-v1.md
  cmd/server/main.go
  internal/{config,handler,service,middleware,i18n,dto}
  pkg/{response,apperr}
  web/{templates,static}
  data/reports/
```

## 主要路由

| 方法 | 路径 | 用途 |
|------|------|------|
| GET | `/` | 首页表单 |
| POST | `/bench` | 创建任务 |
| GET | `/bench/:id` | 进度/报告页 |
| GET | `/bench/:id/events` | SSE |
| POST | `/bench/:id/stop` | 停止 |
| GET | `/bench/:id/report.json` | JSON |
| GET | `/bench/:id/pdf` | PDF |
| GET | `/health` | 健康检查 |
| GET | `/lang/:code` | 切换语言 |

另提供 `/api/v1/bench*` JSON API。

## 验收标准

- `go run ./cmd/server` 后可对授权 URL 压测并看到中文报告
- 切到英文后报告与 PDF 为英文
- 无 MySQL/Redis；重启后可通过 `data/reports/{id}.json` 打开历史报告
- `README.md`（中文）与 `readme_en.md`（英文）相互链接
