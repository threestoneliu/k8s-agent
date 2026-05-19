## ADDED Requirements

### Requirement: Available Skills XML Block
The system SHALL generate an `<available_skills>` XML block for injection into the System Prompt. This block SHALL contain one `<skill>` entry per discovered Skill, with child elements for `name`, `description`, and `location`.

#### Scenario: Single Skill discovered
- **WHEN** exactly one Skill is loaded from `~/.config/k8s-agent/skills/`
- **THEN** the generated `<available_skills>` block SHALL contain one `<skill>` element with valid name, description, and location child elements

#### Scenario: Multiple Skills discovered
- **WHEN** multiple Skills are loaded from `~/.config/k8s-agent/skills/`
- **THEN** the generated `<available_skills>` block SHALL contain all Skills as separate `<skill>` elements

### Requirement: Skill Location Path Resolution
The system SHALL resolve Skill location paths to absolute paths in the `<available_skills>` XML block. The `location` element SHALL contain the absolute path to the Skill's SKILL.md file.

#### Scenario: Location path is absolute
- **WHEN** the Skill's SKILL.md path is already absolute (e.g., `/home/user/.config/...`)
- **THEN** the system SHALL use that path directly in the `<location>` element

#### Scenario: Location path needs expansion
- **WHEN** the Skill's base directory is stored as `~/.config/...`
- **THEN** the system SHALL expand `~` to the user's home directory before inserting into `<location>`

### Requirement: Skill Description for Matching
The system SHALL ensure each Skill's description in the `<available_skills>` block is sufficient for the LLM to determine if the Skill applies to a user's request. Descriptions SHALL be concise and clearly indicate the Skill's purpose.

#### Scenario: Clear, specific description
- **WHEN** a Skill's description is "K8s cluster inspection workflow. Use when user wants to inspect cluster health."
- **THEN** the LLM SHALL be able to match user phrases like "巡检一下" or "inspect the cluster" to this Skill

#### Scenario: Ambiguous description
- **WHEN** a Skill's description is too generic (e.g., "Does things")
- **THEN** the LLM MAY fail to correctly match the Skill to user requests