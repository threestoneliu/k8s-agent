## 1. Skill System Core

- [x] 1.1 Create `pkg/skill/` directory with skill.go, loader.go, registry.go, prompt.go
- [x] 1.2 Implement `pkg/skill/skill.go` with Skill type and Frontmatter struct
- [x] 1.3 Implement `pkg/skill/loader.go` to scan `~/.config/k8s-agent/skills/<skill-name>/SKILL.md`
- [x] 1.4 Implement `pkg/skill/registry.go` to store registered Skills with name, description, location
- [x] 1.5 Implement `pkg/skill/prompt.go` to generate `<available_skills>` XML block

## 2. Read Tool for Skill Discovery

- [x] 2.1 Add `Read` function to `pkg/llm/functions.go` for reading local files
- [x] 2.2 Add Read function registration in function auto-registration
- [x] 2.3 Add `Read` tool description to System Prompt so LLM knows it can use Read

## 3. System Prompt Integration

- [x] 3.1 Modify `pkg/llm/executor.go` to load Skills during initialization
- [x] 3.2 Modify `pkg/agent/agent.go` to inject `<available_skills>` into System Prompt
- [x] 3.3 Add progressive disclosure instructions to System Prompt

## 4. Example Skill

- [x] 4.1 Create example `~/.config/k8s-agent/skills/k8s-inspection/SKILL.md` with YAML frontmatter
- [x] 4.2 Define inspection workflow in the example SKILL.md (nodes → pods → events → report)
- [x] 4.3 Document Skill directory structure in README or docs