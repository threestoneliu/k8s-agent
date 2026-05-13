# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test Commands

```bash
# Build
go build -o k8s-agent ./cmd/k8s-agent

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests with race detection
go test -race ./...

# Run single test
go test ./pkg/k8s/... -run TestParser -v
```

## Architecture

k8s-agent 是一个提供自然语言界面的 Kubernetes CLI 工具，包含意图识别和人工确认机制。

### 核心流程

```
用户输入 → Parser.Parse() → Classifier.Classify() → Executor.Execute()
                                              ↓
                                    Mutation 操作需 ConfirmationManager 确认
```

### 核心组件

- **cmd/cli/** - Cobra CLI 命令层，包括 chat 交互模式
- **pkg/k8s/** - 核心引擎：Executor（执行）
- **pkg/llm/** - LLM 集成层，支持 OpenAI，通过 Function Calling 实现意图识别
- **pkg/confirmation/** - 人工确认管理器，Mutation 操作需要确认
- **pkg/cluster/** - 多集群管理，Registry 管理 kubeconfig
- **pkg/session/** - 会话管理

### 操作分类

- **Query** (get/list/describe/watch/logs/exec) - 直接执行
- **Mutation** (create/update/patch/delete/scale/cordon/uncordon/drain) - 需要确认
- **高危资源** (nodes/persistentvolumes/namespaces/storageclasses) - 无论操作类型都需要确认

### LLM 集成

项目使用 LLM 进行自然语言理解和 Function Calling：
- 配置通过 `config.example.yaml` 或环境变量
- 支持 OpenAI (GPT-4)
- 通过 `pkg/llm/executor.go` 统一调用入口

## Agent skills

### Issue tracker

Issues live in GitHub. See `docs/agents/issue-tracker.md`.

### Triage labels

Five-role state machine: needs-triage → needs-info → ready-for-agent/ready-for-human → wontfix. See `docs/agents/triage-labels.md`.

### Domain docs

Single-context: one `CONTEXT.md` + `docs/adr/` at repo root. See `docs/agents/domain.md`.