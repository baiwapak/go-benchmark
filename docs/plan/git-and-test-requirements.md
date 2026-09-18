# Git 与测试注意事项（统一要求）

> 状态：已落地到 `.cursor/rules/00-git-commit-workflow.mdc`（或等价文件名）与 `AGENTS.md` 索引。

## 要求

- 每次改动完成后，都必须创建一个对应的 Git commit，以便后续追踪和回滚。
- 每次改动后，都必须编写或更新相关测试，并在交付给用户前，确保所有测试和验证全部通过。

## 落地位置

| 文件 | 作用 |
|------|------|
| `.cursor/rules/00-git-commit-workflow.mdc`（或 `git-commit-workflow.mdc`） | alwaysApply，Agent 强制遵守 |
| `AGENTS.md` | 规则索引入口 |
