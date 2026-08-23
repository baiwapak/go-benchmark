# V1.1 – V1.3 路线图

## V1.1 实时曲线 + 历史报告列表

- **时序聚合**：`report.Aggregator` 按秒记录 RPS、平均延迟、成功率
- **SSE / 报告**：`ProgressSnapshot.Series`、`BenchReport.Series` 携带时序数据
- **前端图表**：`bench.html` Canvas 折线图，压测中与完成后实时更新
- **历史列表**：`GET /reports`，从 `data/reports/*.json` 读取摘要

## V1.2 Ramp-up + POST body

- **升压**：`ramp_sec` 参数，引擎线性增加活跃 worker 数
- **请求体**：`body_text` 字段，支持 POST/PUT/PATCH；默认 `Content-Type: application/json`
- **限制**：`MAX_RAMP_SEC`、`MAX_BODY_BYTES`（默认 64KB）

## V1.3 报告对比 + SSRF 防护

- **对比页**：`GET /reports/compare?a=&b=`，对比 RPS、成功率、延迟分位
- **SSRF**：`internal/security` 拦截 loopback、私网、link-local 等地址

## 路由

| 方法 | 路径 | 说明 |
|------|------|------|
| GET | `/reports` | 历史报告列表 |
| POST | `/reports/:id/delete` | 删除单条报告 |
| POST | `/reports/clear` | 清空全部报告 |
| GET | `/reports/compare` | 报告对比 |

## 验收

- 压测进行中可看到 RPS/延迟/成功率曲线
- `/reports` 可浏览历史记录
- 可选 ramp_sec、body_text 并成功压测 API
- `http://127.0.0.1` 等内网地址被拒绝
