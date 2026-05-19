# Skill System Implementation Plan

> **For agentic workers:** Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add Skill system to k8s-agent enabling standardized workflows via Agent Skills progressive disclosure mechanism.

**Architecture:** New `pkg/skill/` package handles Skill discovery, loading, and registry. Skills are exposed to the LLM via `<available_skills>` XML block in System Prompt. LLM self-selects Skills based on description matching and reads SKILL.md via standard Read tool. The `Read` tool is implemented as a registered function, allowing the LLM to read local files during conversation.

**Tech Stack:** Go (k8s-agent), YAML frontmatter parsing, file system scanning

---

## Task 1: Create pkg/skill/ Package Structure

**Files:**
- Create: `pkg/skill/skill.go` (shared Skill type definition)

- [ ] **Step 1: Create pkg/skill/ directory and skill.go with Skill struct**

```go
// pkg/skill/skill.go
package skill

import "fmt"

// Skill represents a loaded Skill with metadata and location
type Skill struct {
    Name        string
    Description string
    Location    string // absolute path to SKILL.md
}

// Frontmatter represents the YAML frontmatter of a SKILL.md
type Frontmatter struct {
    Name           string            `yaml:"name"`
    Description    string            `yaml:"description"`
    License        string            `yaml:"license,omitempty"`
    Compatibility  string            `yaml:"compatibility,omitempty"`
    Metadata       map[string]string `yaml:"metadata,omitempty"`
}

// String returns the skill name
func (s *Skill) String() string {
    return s.Name
}
```

- [ ] **Step 2: Verify skill.go compiles**

Run: `go build ./pkg/skill/...`
Expected: BUILD SUCCESS

- [ ] **Step 3: Commit**

```bash
git add pkg/skill/skill.go
git commit -m "feat(skill): add Skill type definition"
```

---

## Task 2: Implement loader.go — Scan and Load SKILL.md Files

**Files:**
- Create: `pkg/skill/loader.go`
- Create: `pkg/skill/loader_test.go`

- [ ] **Step 1: Create pkg/skill/loader.go with directory scanning**

```go
// pkg/skill/loader.go
package skill

import (
    "fmt"
    "os"
    "path/filepath"

    "gopkg.in/yaml.v3"
)

// LoadSkills loads all Skills from the user config directory
func LoadSkills(configDir string) ([]*Skill, error) {
    skillsDir := filepath.Join(configDir, "skills")

    // Check if directory exists
    if _, err := os.Stat(skillsDir); os.IsNotExist(err) {
        return []*Skill{}, nil // empty is ok
    }

    entries, err := os.ReadDir(skillsDir)
    if err != nil {
        return nil, fmt.Errorf("failed to read skills directory: %w", err)
    }

    var skills []*Skill
    for _, entry := range entries {
        if !entry.IsDir() {
            continue // skip files
        }

        skillPath := filepath.Join(skillsDir, entry.Name(), "SKILL.md")
        skill, err := loadSkillFile(skillPath)
        if err != nil {
            // log warning but continue
            fmt.Printf("Warning: failed to load skill %s: %v\n", entry.Name(), err)
            continue
        }
        skills = append(skills, skill)
    }

    return skills, nil
}

// loadSkillFile loads a single SKILL.md file
func loadSkillFile(path string) (*Skill, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, fmt.Errorf("cannot read file: %w", err)
    }

    // Split frontmatter from markdown body
    fm, _, err := parseFrontmatter(string(data))
    if err != nil {
        return nil, fmt.Errorf("invalid frontmatter: %w", err)
    }

    // Validate required fields
    if fm.Name == "" {
        return nil, fmt.Errorf("missing required field: name")
    }
    if fm.Description == "" {
        return nil, fmt.Errorf("missing required field: description")
    }

    // Resolve absolute path
    absPath, err := filepath.Abs(path)
    if err != nil {
        return nil, fmt.Errorf("cannot resolve absolute path: %w", err)
    }

    return &Skill{
        Name:        fm.Name,
        Description: fm.Description,
        Location:    absPath,
    }, nil
}

// parseFrontmatter extracts YAML frontmatter from markdown content
func parseFrontmatter(content string) (*Frontmatter, string, error) {
    if len(content) < 4 || content[:4] != "---\n" {
        return nil, "", fmt.Errorf("missing frontmatter separator")
    }

    endIdx := 4
    for i := 4; i < len(content)-3; i++ {
        if content[i:i+4] == "\n---" {
            endIdx = i + 4
            break
        }
    }

    yamlContent := content[4 : endIdx-4]
    body := content[endIdx:]

    var fm Frontmatter
    if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil {
        return nil, "", fmt.Errorf("invalid yaml: %w", err)
    }

    return &fm, body, nil
}
```

- [ ] **Step 2: Create pkg/skill/loader_test.go with TDD**

```go
// pkg/skill/loader_test.go
package skill

import (
    "os"
    "path/filepath"
    "testing"
)

func TestLoadSkills_ValidSkill(t *testing.T) {
    tmpDir := t.TempDir()
    skillDir := filepath.Join(tmpDir, "test-skill")
    os.Mkdir(skillDir, 0755)

    skillFile := filepath.Join(skillDir, "SKILL.md")
    content := `---
name: test-skill
description: A test skill for unit testing
license: Apache-2.0
---
# Test Skill
`
    os.WriteFile(skillFile, []byte(content), 0644)

    skills, err := LoadSkills(tmpDir)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if len(skills) != 1 {
        t.Fatalf("expected 1 skill, got %d", len(skills))
    }

    if skills[0].Name != "test-skill" {
        t.Errorf("expected name 'test-skill', got %q", skills[0].Name)
    }
}

func TestLoadSkills_MissingFrontmatter(t *testing.T) {
    tmpDir := t.TempDir()
    skillDir := filepath.Join(tmpDir, "bad-skill")
    os.Mkdir(skillDir, 0755)

    skillFile := filepath.Join(skillDir, "SKILL.md")
    content := `# No frontmatter here`
    os.WriteFile(skillFile, []byte(content), 0644)

    skills, err := LoadSkills(tmpDir)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    if len(skills) != 0 {
        t.Errorf("expected 0 skills, got %d", len(skills))
    }
}
```

- [ ] **Step 3: Verify all tests pass**

Run: `go test ./pkg/skill/... -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add pkg/skill/loader.go pkg/skill/loader_test.go
git commit -m "feat(skill): add loader for scanning SKILL.md files"
```

---

## Task 3: Implement registry.go — Skill Registration and Management

**Files:**
- Create: `pkg/skill/registry.go`
- Create: `pkg/skill/registry_test.go`

- [ ] **Step 1: Create registry.go with Registry struct**

```go
// pkg/skill/registry.go
package skill

import (
    "fmt"
    "sync"
)

// Registry manages all loaded Skills
type Registry struct {
    mu     sync.RWMutex
    skills map[string]*Skill
}

// NewRegistry creates a new empty Registry
func NewRegistry() *Registry {
    return &Registry{
        skills: make(map[string]*Skill),
    }
}

// Register adds a Skill to the registry
func (r *Registry) Register(s *Skill) error {
    r.mu.Lock()
    defer r.mu.Unlock()

    if _, exists := r.skills[s.Name]; exists {
        return fmt.Errorf("skill %q already registered", s.Name)
    }
    r.skills[s.Name] = s
    return nil
}

// Get retrieves a Skill by name
func (r *Registry) Get(name string) (*Skill, bool) {
    r.mu.RLock()
    defer r.mu.RUnlock()

    s, ok := r.skills[name]
    return s, ok
}

// List returns all registered Skills
func (r *Registry) List() []*Skill {
    r.mu.RLock()
    defer r.mu.RUnlock()

    result := make([]*Skill, 0, len(r.skills))
    for _, s := range r.skills {
        result = append(result, s)
    }
    return result
}

// Count returns the number of registered Skills
func (r *Registry) Count() int {
    r.mu.RLock()
    defer r.mu.RUnlock()
    return len(r.skills)
}
```

- [ ] **Step 2: Create registry_test.go**

```go
// pkg/skill/registry_test.go
package skill

import (
    "testing"
)

func TestRegistry_RegisterAndGet(t *testing.T) {
    r := NewRegistry()
    s := &Skill{Name: "test", Description: "desc", Location: "/path/to/SKILL.md"}

    err := r.Register(s)
    if err != nil {
        t.Fatalf("unexpected error: %v", err)
    }

    retrieved, ok := r.Get("test")
    if !ok {
        t.Fatal("skill not found")
    }
    if retrieved.Name != "test" {
        t.Errorf("expected name 'test', got %q", retrieved.Name)
    }
}

func TestRegistry_DuplicateRegistration(t *testing.T) {
    r := NewRegistry()
    s := &Skill{Name: "dup", Description: "desc", Location: "/path"}

    r.Register(s)
    err := r.Register(s)

    if err == nil {
        t.Fatal("expected error for duplicate registration")
    }
}

func TestRegistry_List(t *testing.T) {
    r := NewRegistry()
    r.Register(&Skill{Name: "s1", Description: "d1", Location: "/p1"})
    r.Register(&Skill{Name: "s2", Description: "d2", Location: "/p2"})

    list := r.List()
    if len(list) != 2 {
        t.Errorf("expected 2 skills, got %d", len(list))
    }
}
```

- [ ] **Step 3: Verify tests pass**

Run: `go test ./pkg/skill/... -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add pkg/skill/registry.go pkg/skill/registry_test.go
git commit -m "feat(skill): add Registry for skill management"
```

---

## Task 4: Implement prompt.go — Generate `<available_skills>` XML Block

**Files:**
- Create: `pkg/skill/prompt.go`
- Create: `pkg/skill/prompt_test.go`

- [ ] **Step 1: Create prompt.go with XML generation**

```go
// pkg/skill/prompt.go
package skill

import (
    "bytes"
    "fmt"
    "text/template"
)

// AvailableSkillsXML generates the <available_skills> XML block for System Prompt
func AvailableSkillsXML(skills []*Skill) string {
    if len(skills) == 0 {
        return ""
    }

    tmpl := `<available_skills>
{{range .}}<skill>
  <name>{{.Name}}</name>
  <description>{{.Description}}</description>
  <location>{{.Location}}</location>
</skill>
{{end}}</available_skills>`

    t := template.Must(template.New("available_skills").Parse(tmpl))

    var buf bytes.Buffer
    if err := t.Execute(&buf, skills); err != nil {
        return fmt.Sprintf("<!-- error generating skills: %v -->", err)
    }

    return buf.String()
}

// ProgressiveDisclosurePrompt returns the instruction block for LLM skill discovery
func ProgressiveDisclosurePrompt() string {
    return `## Skills (mandatory)
Before replying: scan <available_skills> <description> entries.
- If exactly one skill clearly applies: read its SKILL.md at <location> with ` + "`Read`" + `, then follow it.
- If multiple could apply: choose the most specific one, then read/follow it.
- If none clearly apply: do not read any SKILL.md.
Constraints: never read more than one skill up front; only read after selecting.`
}
```

- [ ] **Step 2: Create prompt_test.go**

```go
// pkg/skill/prompt_test.go
package skill

import (
    "strings"
    "testing"
)

func TestAvailableSkillsXML_SingleSkill(t *testing.T) {
    skills := []*Skill{
        {Name: "k8s-inspection", Description: "K8s cluster inspection", Location: "/home/user/.config/k8s-agent/skills/k8s-inspection/SKILL.md"},
    }

    xml := AvailableSkillsXML(skills)

    if !strings.Contains(xml, "<available_skills>") {
        t.Error("missing <available_skills> tag")
    }
    if !strings.Contains(xml, "<name>k8s-inspection</name>") {
        t.Error("missing skill name")
    }
}

func TestAvailableSkillsXML_Empty(t *testing.T) {
    xml := AvailableSkillsXML([]*Skill{})
    if xml != "" {
        t.Errorf("expected empty string for no skills, got %q", xml)
    }
}

func TestProgressiveDisclosurePrompt(t *testing.T) {
    prompt := ProgressiveDisclosurePrompt()

    if !strings.Contains(prompt, "scan <available_skills>") {
        t.Error("missing scan instruction")
    }
    if !strings.Contains(prompt, "Read") {
        t.Error("missing Read tool instruction")
    }
    if !strings.Contains(prompt, "never read more than one skill") {
        t.Error("missing constraint")
    }
}
```

- [ ] **Step 3: Verify tests pass**

Run: `go test ./pkg/skill/... -v`
Expected: PASS

- [ ] **Step 4: Commit**

```bash
git add pkg/skill/prompt.go pkg/skill/prompt_test.go
git commit -m "feat(skill): add prompt.go for available_skills XML generation"
```

---

## Task 5: Implement Read Tool for SKILL.md Discovery

**Files:**
- Modify: `pkg/llm/functions.go`
- Modify: `pkg/llm/auto_register.go`

**Context:** The LLM needs a `Read` tool to read SKILL.md files when it decides a Skill matches the user's request. This is fundamental to the progressive disclosure mechanism.

- [ ] **Step 1: Read current functions.go to understand the function definition pattern**

Run: `cat pkg/llm/functions.go`

- [ ] **Step 2: Add Read function definition to functions.go**

Add this function definition alongside existing functions:

```go
// ReadFunctionDefinition returns the definition for the Read tool
func ReadFunctionDefinition() FunctionDefinition {
    return FunctionDefinition{
        Name:        "Read",
        Description: "Read the contents of a file from the local filesystem. Use this to read SKILL.md files when you need to follow a skill's workflow.",
        Parameters: map[string]any{
            "type": "object",
            "properties": map[string]any{
                "file_path": map[string]any{
                    "type":        "string",
                    "description": "The absolute path to the file to read",
                },
            },
            "required": []string{"file_path"},
        },
    }
}
```

- [ ] **Step 3: Implement the Read function handler**

In the same file or in a new `read.go`, add:

```go
// ReadFile handles the Read function call
func ReadFile(args json.RawMessage) (string, error) {
    var input struct {
        FilePath string `json:"file_path"`
    }
    if err := json.Unmarshal(args, &input); err != nil {
        return "", fmt.Errorf("invalid arguments: %w", err)
    }

    content, err := os.ReadFile(input.FilePath)
    if err != nil {
        return "", fmt.Errorf("failed to read file: %w", err)
    }

    return string(content), nil
}
```

- [ ] **Step 4: Register the Read function in auto_register.go**

Find where functions are registered and add:

```go
// In the function registration section
funcs = append(funcs, FunctionWithHandler{
    Definition: ReadFunctionDefinition(),
    Handler:    ReadFile,
})
```

- [ ] **Step 5: Verify build**

Run: `go build ./pkg/llm/...`
Expected: BUILD SUCCESS

- [ ] **Step 6: Commit**

```bash
git add pkg/llm/functions.go pkg/llm/auto_register.go
git commit -m "feat(skill): add Read tool for SKILL.md discovery"
```

---

## Task 6: Integrate Skills into LLM Executor

**Files:**
- Modify: `pkg/llm/executor.go`

- [ ] **Step 1: Read current executor.go to understand the initialization pattern**

Run: `cat pkg/llm/executor.go`

- [ ] **Step 2: Add skill loading in executor initialization**

Add to imports:
```go
import "k8s-agent/pkg/skill"
```

Add to Executor struct:
```go
type Executor struct {
    // ... existing fields ...
    skillRegistry *skill.Registry
}
```

In NewExecutor function, add after cluster init:
```go
// Load skills
homeDir, err := os.UserHomeDir()
var skillRegistry *skill.Registry
if err == nil {
    skills, err := skill.LoadSkills(homeDir + "/.config/k8s-agent")
    if err == nil {
        skillRegistry = skill.NewRegistry()
        for _, s := range skills {
            skillRegistry.Register(s)
        }
    }
}
```

- [ ] **Step 3: Add method to get skills XML for prompt injection**

Add to Executor:
```go
// GetAvailableSkillsXML returns the <available_skills> XML block
func (e *Executor) GetAvailableSkillsXML() string {
    if e.skillRegistry == nil {
        return ""
    }
    skills := e.skillRegistry.List()
    return skill.AvailableSkillsXML(skills)
}

// GetProgressiveDisclosurePrompt returns the skill instruction block
func (e *Executor) GetProgressiveDisclosurePrompt() string {
    return skill.ProgressiveDisclosurePrompt()
}
```

- [ ] **Step 4: Verify build**

Run: `go build ./pkg/llm/...`
Expected: BUILD SUCCESS

- [ ] **Step 5: Commit**

```bash
git add pkg/llm/executor.go
git commit -m "feat(skill): integrate skill loading in executor initialization"
```

---

## Task 7: Integrate Skills into Agent System Prompt

**Files:**
- Modify: `pkg/agent/agent.go`

- [ ] **Step 1: Read current agent.go to find System Prompt construction**

Run: `grep -n "systemPrompt\|SystemPrompt\|system_prompt" pkg/agent/agent.go`

- [ ] **Step 2: Add skill prompt to system prompt**

Find where system prompt is built, add:
```go
// Append skill prompt to system prompt
skillXML := executor.GetAvailableSkillsXML()
skillInstruction := executor.GetProgressiveDisclosurePrompt()
if skillXML != "" {
    systemPrompt += "\n\n" + skillInstruction + "\n\n" + skillXML
}
```

- [ ] **Step 3: Verify build**

Run: `go build ./pkg/agent/...`
Expected: BUILD SUCCESS

- [ ] **Step 4: Commit**

```bash
git add pkg/agent/agent.go
git commit -m "feat(skill): inject available_skills into agent system prompt"
```

---

## Task 8: Create Example k8s-inspection Skill

**Files:**
- Create: `~/.config/k8s-agent/skills/k8s-inspection/SKILL.md` (in user home, not repo)

- [ ] **Step 1: Create example SKILL.md**

```markdown
---
name: k8s-inspection
description: K8s cluster inspection workflow. Use when user wants to inspect cluster health, check node status, or perform routine maintenance checks.
license: Apache-2.0
compatibility: k8s-agent
metadata:
  author: k8s-agent
  version: "1.0"
---

# K8s Inspection

## Workflows

### 巡检流程 (Inspection Workflow)

1. **检查节点状态**：调用 `resource_list(resource="nodes")`
2. **分析节点状态**：确认所有节点都是 Ready 状态
3. **检查 Pod 状态**：调用 `resource_list(resource="pods")`
4. **分析 Pod 状态**：识别 Error/Pending/CrashLoopBackOff 状态的 Pod
5. **检查 Events**：调用 `resource_list(resource="events")`
6. **生成巡检报告**：汇总节点、Pod、Events 的状态，输出巡检结论

## 输出格式

巡检完成后，输出以下格式的摘要：

```
## 巡检报告

### 节点状态
- 总节点数: X
- Ready: Y
- NotReady: Z

### Pod 状态
- 总 Pods: X
- Running: Y
- Error: Z
- Pending: W

### 关键事件
- [列出最近的关键事件]

### 结论
[给出巡检结论和建议]
```

## 触发方式

- 显式: `/skill k8s-inspection` 或 `用巡检skill`
- 隐式: "巡检一下", "检查集群健康", "cluster inspection"
```

- [ ] **Step 2: Test skill loading manually**

Run: `go run ./cmd/k8s-agent chat` and try "巡检一下" to verify skill discovery works

---

## Task 9: Add Documentation to README

**Files:**
- Modify: `README.md`

- [ ] **Step 1: Add Skills section to README**

Add before Configuration section:

```markdown
## Skills

k8s-agent supports Skills for standardized workflows. Skills are stored in `~/.config/k8s-agent/skills/`.

### Using Skills

Skills are discovered automatically through the System Prompt. When your request matches a Skill's description, the LLM will read the Skill's SKILL.md and follow its workflow.

**Explicit trigger:**
```
> /skill k8s-inspection
> 用巡检skill
```

**Implicit trigger:**
```
> 巡检一下
> 检查集群健康
```

### Creating Custom Skills

Create a directory at `~/.config/k8s-agent/skills/<skill-name>/SKILL.md`:

```yaml
---
name: my-skill
description: Description of when this skill applies
license: Apache-2.0
---

# My Skill

## Workflows

1. Step 1: call resource_list(resource="...")
2. Step 2: analyze results
3. Step 3: generate output
```

### Available Skills

- `k8s-inspection` - Cluster health inspection workflow (nodes → pods → events → report)
```

- [ ] **Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add Skills section to README"
```

---

## Verification

After implementation, verify:

1. **Unit tests pass:** `go test ./pkg/skill/... -v`
2. **Build succeeds:** `go build ./...`
3. **Skill discovery works:** Run `k8s-agent chat` and say "巡检一下"
4. **Empty skills handled:** No crash when `~/.config/k8s-agent/skills/` doesn't exist
5. **Read tool works:** LLM can read SKILL.md files via the Read function