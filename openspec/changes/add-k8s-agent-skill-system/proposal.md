## Why

当前 k8s-agent 通过 Function Calling 让 LLM 调用 k8s API，但每次都依赖 LLM 规划执行路径。对于标准化流程（如巡检），每次结果可能不一致，且增加 LLM 上下文消耗。

引入 Skill 系统：使用 Agent Skills 标准的渐进式披露机制，让 LLM 能调用预定义的 workflow，减少规划依赖，提高结果一致性。

## What Changes

**Skill 发现与加载**
- From: LLM 直接调用 Function，不存在 Skill 概念
- To: 启动时加载 `~/.config/k8s-agent/skills/` 下所有 Skill，构建 `<available_skills>` XML 块注入 System Prompt
- Reason: 让模型能发现并调用预定义 Skill
- Impact: 非破坏性，新增 pkg/skill/ 模块

**Skill 执行机制**
- From: LLM 根据意图自由调用 functions
- To: 模型匹配 Skill 后读取 SKILL.md，按照其中定义的 workflow 依次调用 functions
- Reason: 固化流程，减少 LLM 规划，提高一致性
- Impact: 非破坏性，Skill workflow 通过现有 Function Calling 执行

**Skill 格式**
- From: 无
- To: SKILL.md = YAML frontmatter（name/description/license） + Markdown body（workflow）
- Reason: 符合 Agent Skills 标准，模型通过 description 自动匹配

**Skill 触发**
- From: 用户自然语言，LLM 自己判断调用哪些 function
- To: 支持显式（`/skill k8s-inspection`）+ 隐式（"巡检一下"→模型匹配 Skill）
- Reason: 提供明确的 Skill 调用方式
- Impact: 非破坏性

## Capabilities

### New Capabilities
- `skill-system`: 核心 Skill 系统能力，包括 Skill 加载、注册、发现、通过 System Prompt 注入模型
- `skill-execution`: Skill workflow 执行能力，模型读取 SKILL.md 后按步骤调用 functions
- `skill-discovery`: Skill 发现机制，`<available_skills>` XML 块 + 渐进式披露

### Modified Capabilities
- （无）

## Impact

**Affected:**
- `pkg/skill/` - 新增 Skill 系统模块（loader.go, registry.go, prompt.go）
- `pkg/llm/executor.go` - 初始化时加载 Skill，构建 System Prompt
- `pkg/agent/agent.go` - 注入 Skill 到 System Prompt

**New Dependencies:**
- Agent Skills 标准知识（已研究并在 design.md 中记录）

**Configuration:**
- `~/.config/k8s-agent/skills/` - Skill 存储目录（用户级）