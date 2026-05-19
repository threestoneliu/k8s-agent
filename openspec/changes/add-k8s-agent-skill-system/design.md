# Skill System Design

## Context

k8s-agent 当前使用 Function Calling 让 LLM 调用 k8s API。流程：
- 用户输入自然语言
- LLM 解析意图，调用 registered functions（resource_list, resource_get 等）
- Executor 执行 k8s 操作

**问题**：对于标准化流程（如巡检），每次都依赖 LLM 规划执行路径，结果不确定。

**需求**：固化流程，减少 LLM 规划依赖，提高结果一致性。

**参考标准**：Agent Skills 标准（agentskills.io）使用渐进式披露机制。

## Goals / Non-Goals

**Goals:**
- Skill 系统让 LLM 能调用预定义的 workflow（如巡检）
- 支持用户自定义 Skill
- 通过 System Prompt 嵌入 `<available_skills>` 让模型发现 Skill
- 采用渐进式披露：模型自行判断何时读取 Skill，执行后继续

**Non-Goals:**
- 不实现 Skill 市场或分享机制
- 不实现 Skill 的版本管理
- 不实现 Skill 之间的依赖关系

## Decisions

### 1. Skill 存储位置

**Decision**: `~/.config/k8s-agent/skills/<skill-name>/SKILL.md`

**Rationale**: 用户级目录，支持用户自定义 Skill，与 Agent Skills 标准一致。

### 2. Skill 格式

**Decision**: SKILL.md = YAML frontmatter + Markdown body

```yaml
---
name: k8s-inspection
description: K8s cluster inspection workflow. Use when user wants to inspect cluster health.
license: Apache-2.0
compatibility: k8s-agent
metadata:
  author: k8s-agent
  version: "1.0"
---

# K8s Inspection

## Workflows

### 巡检流程
1. 检查节点状态：调用 resource_list(resource="nodes")
2. 分析节点是否全部 Ready
3. 检查 Pod 状态：调用 resource_list(resource="pods")
4. 分析是否有 Error/Pending Pod
5. 检查 Events：调用 resource_list(resource="events")
6. 生成巡检报告
```

### 3. 发现机制

**Decision**: System Prompt 嵌入 `<available_skills>` XML 块

模型看到的 prompt 结构：
```
## Skills (mandatory)
Before replying: scan <available_skills> <description> entries.
- If exactly one skill clearly applies: read its SKILL.md at <location> with `Read`, then follow it.
- If multiple could apply: choose the most specific one, then read/follow it.
- If none clearly apply: do not read any SKILL.md.
Constraints: never read more than one skill up front; only read after selecting.

<available_skills>
  <skill>
    <name>k8s-inspection</name>
    <description>K8s cluster inspection workflow. Use when user wants to inspect cluster health.</description>
    <location>~/.config/k8s-agent/skills/k8s-inspection/SKILL.md</location>
  </skill>
</available_skills>
```

### 4. 执行机制

**Decision**: 通过现有 Function Calling 执行 Skill workflow

Skill workflow 中的步骤通过调用已注册的 functions 实现：
- `resource_list` - 列出 k8s 资源
- `resource_get` - 获取单个资源详情
- 其他已注册的 functions

模型按照 SKILL.md 中的 workflow 依次调用 functions。

### 5. 触发方式

**Decision**: 支持显式 + 隐式触发

- **显式**: 用户输入 `/skill k8s-inspection` 或 `用巡检skill`
- **隐式**: 用户说"巡检一下"，LLM 识别意图，读取对应 SKILL.md

### 6. Skill 加载流程

```
1. 启动时加载 ~/.config/k8s-agent/skills/ 下所有 Skill
2. 扫描每个 Skill 目录，读取 SKILL.md 提取 name/description
3. 构建 <available_skills> XML 块
4. 将 XML 块注入 System Prompt
5. 模型根据 <description> 匹配 Skill，调用 Read 读取内容
6. 模型按 SKILL.md 中的 workflow 执行
```

## Architecture

```
k8s-agent
├── cmd/cli/          # CLI 入口
├── pkg/
│   ├── agent/        # Agent loop
│   ├── llm/         # LLM integration (functions.go, auto_register.go)
│   ├── k8s/         # Executor (resource_list, resource_get)
│   ├── skill/        # NEW: Skill system
│   │   ├── loader.go    # 加载用户目录下的 Skills
│   │   ├── registry.go  # Skill 注册与管理
│   │   └── prompt.go   # 构建 <available_skills> XML
│   └── ui/           # TUI
└── ~/.config/k8s-agent/skills/
    ├── k8s-inspection/
    │   └── SKILL.md
    └── user-defined-skill/
        └── SKILL.md
```

## Files to Modify/Create

1. **pkg/skill/loader.go** - Scan skills directory, load SKILL.md
2. **pkg/skill/registry.go** - Manage registered skills, provide XML output
3. **pkg/skill/prompt.go** - Build `<available_skills>` for system prompt
4. **pkg/llm/executor.go** - Add skill loading during initialization
5. **pkg/agent/agent.go** - Inject skills into system prompt

## Risks / Trade-offs

| 风险 | 影响 | 缓解 |
|------|------|------|
| 模型不遵循 prompt 指令 | Skill 不被触发 | 良好的 description 设计 + prompt 优化 |
| Skill 数量过多导致 prompt 超长 | 模型性能下降 | 限制同时嵌入的 Skill 数量（按需加载） |
| 用户 Skill 格式错误 | 无法正常加载 | 加载时验证 YAML 格式 |
| Skill 执行结果不一致 | 巡检报告格式不统一 | 在 SKILL.md 中明确定义输出格式 |

## Open Questions

1. **Skill 数量限制**: 同一 prompt 中嵌入的 Skill 数量是否有限制？
   - 建议：限制最多 10 个，超出按使用频率筛选

2. **Skill 更新刷新**: Skill 内容变更后如何刷新？
   - 建议：启动时加载一次，或增加 `/skill reload` 命令

3. **内置 vs 用户 Skill**: 是否需要内置 Skill（如 k8s-inspection）？
   - 建议：先支持用户自定义，内置 Skill 作为示例后续添加