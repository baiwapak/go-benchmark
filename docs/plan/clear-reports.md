# 历史报告清空

## 目标

在 `GET /reports` 历史报告页支持删除单条报告与清空全部，避免 `data/reports/` 无限堆积。

## 行为

- **删除单条**：每行「删除」，确认后移除对应 JSON，并从内存任务表剔除（进行中的任务不可删）。
- **清空全部**：页头「清空全部」，确认后删除全部已落盘报告；跳过 `pending` / `running` 任务。
- 操作不可恢复；仅删除 `.json` 报告文件，不影响正在运行的压测。

## 实现

| 层 | 变更 |
|----|------|
| 存储 | `FileStore.Delete` / `ClearExcept`，ID 限十六进制以防路径穿越 |
| 服务 | `Manager.DeleteReport` / `ClearReports`，同步清理内存中的已结束任务 |
| 路由 | `POST /reports/:id/delete`、`POST /reports/clear`，成功后回列表 |
| UI | `reports.html` 删除/清空按钮 + `confirm` |
| 文案 | `reports.delete` / `reports.clear` 及确认提示（中英） |

## 验收

- 列表有数据时可见「清空全部」与每行「删除」
- 确认后报告从列表与磁盘消失
- 取消确认不删除
- 空列表不显示清空按钮
