## Design Summary

基于 Agent Skills 标准 + k8s-agent 现有架构的 Skill 系统设计。

**核心发现：**
- Agent Skills 使用渐进式披露（Progressive Disclosure）机制
- 通过 System Prompt 嵌入 `<available_skills>` XML 块
- 模型根据 description 自行匹配并调用 Read 工具读取 SKILL.md
- 设计哲学："nudge not force"，通过 prompt 指令诱导而非技术强制

**系统架构：**
```
~/.config/k8s-agent/skills/
├── k8s-inspection/
│   ├── SKILL.md           # YAML frontmatter + Markdown instructions
│   └── references/        # 可选：参考文档
└── ...

System Prompt 包含:
<available_skills>
  <skill>
    <name>k8s-inspection</name>
    <description>K8s cluster inspection...</description>
    <location>~/.config/k8s-agent/skills/k8s-inspection/SKILL.md</location>
  </skill>
</available_skills>
```

**工作流程：**
1. 模型扫描 `<available_skills>` 列表
2. 根据 description 匹配最合适的 skill
3. 调用标准 Read 工具读取 SKILL.md
4. 按照 SKILL.md 中的 workflow 执行（通过 Function Calling）

## Alternatives Considered

### 方案 A：Function Calling 触发（未采用）
- **做法**：Skill 作为特殊 function，`execute_skill(name)` 调用
- **優點**：技术强制，控制粒度细
- **缺點**：增加复杂度，与现有 Function Calling 职责重叠
- **為何未採用**：Agent Skills 标准使用 Read 工具更简单

### 方案 B：代码内置判断（未采用）
- **做法**：Go 代码判断用户意图，直接执行对应 Skill 逻辑
- **優點**：确定性高，性能好
- **缺點**：不灵活，难以扩展新 Skill
- **為何未採用**：不符合渐进式披露设计，绕过 LLM 判断

### 方案 C：Agent Skills 标准 + 渐进式披露（采用）
- **做法**：采用 Agent Skills 标准，Skill 存储在 ~/.config/k8s-agent/skills/，通过 System Prompt 触发
- **優點**：符合行业标准，扩展性好，模型自主判断
- **缺點**：依赖模型遵循 prompt 指令
- **為何採用**：与 Agent Skills 生态兼容，与现有架构（Function Calling）配合良好

## Agreed Approach

采用方案 C，基于 Agent Skills 标准实现 Skill 系统：

1. **Skill 存储**：`~/.config/k8s-agent/skills/<skill-name>/SKILL.md`
2. **发现机制**：System Prompt 嵌入 `<available_skills>` XML 块
3. **匹配方式**：模型根据 description 自行判断
4. **读取方式**：标准 Read 工具，一次只读一个
5. **执行方式**：通过现有 Function Calling（resource_list, resource_get 等）
6. **触发方式**：显式 (`/skill k8s-inspection`) + 隐式（LLM 意图识别）

## Key Decisions

1. **Skill 格式**：SKILL.md = YAML frontmatter（name/description/license/metadata）+ Markdown body
2. **存储位置**：`~/.config/k8s-agent/skills/`（用户级目录）
3. **触发机制**：System Prompt 中的 `<available_skills>` + 自然语言指令
4. **执行机制**：通过现有 Function Calling，Skill workflow 调用已注册的 functions
5. **判断方式**：模型自行判断（基于 description 匹配），不是代码强制
6. **限制读取**：prompt 约束 "never read more than one skill up front; only read after selecting"

## Open Questions

1. **Skill 注册**：用户自定义 Skill 如何注册到系统？（通过目录扫描还是显式命令）
2. **Skill 数量限制**：同一 prompt 中可嵌入的 Skill 数量是否有限制？
3. **Skill 更新**：Skill 内容变更后是否需要重启 agent 或刷新 context？

无重大阻塞问题，核心设计已明确。