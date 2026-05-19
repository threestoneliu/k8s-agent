## ADDED Requirements

### Requirement: Skill Discovery Mechanism
The system SHALL provide a mechanism for the LLM to discover available Skills through the System Prompt. A `<available_skills>` XML block SHALL be embedded in the System Prompt, containing each Skill's name, description, and file location.

#### Scenario: Skill discovery on agent startup
- **WHEN** the agent starts and loads Skills from `~/.config/k8s-agent/skills/`
- **THEN** the System Prompt SHALL contain an `<available_skills>` block listing all discovered Skills with their name, description, and location

#### Scenario: Empty skills directory
- **WHEN** no Skills exist in `~/.config/k8s-agent/skills/`
- **THEN** the `<available_skills>` block SHALL be empty or omitted from the System Prompt

### Requirement: Progressive Disclosure Integration
The system SHALL integrate with the Agent Skills standard progressive disclosure mechanism. The System Prompt SHALL include instructions directing the LLM to scan `<available_skills>` descriptions and read a Skill's SKILL.md when exactly one Skill clearly applies.

#### Scenario: LLM selects a matching Skill
- **WHEN** the LLM determines that a user's request matches exactly one Skill's description
- **THEN** the LLM SHALL read that Skill's SKILL.md file using the Read tool and follow the workflow defined within

#### Scenario: No matching Skill
- **WHEN** the LLM determines that no Skill clearly applies to the user's request
- **THEN** the LLM SHALL NOT read any SKILL.md and shall proceed with normal Function Calling behavior

#### Scenario: Multiple potential matches
- **WHEN** the LLM determines that multiple Skills could apply
- **THEN** the LLM SHALL choose the most specific one and read its SKILL.md

### Requirement: Skill File Location Convention
The system SHALL use the convention `~/.config/k8s-agent/skills/<skill-name>/SKILL.md` for Skill storage. Each Skill directory SHALL contain exactly one SKILL.md file at its root.

#### Scenario: Valid Skill structure
- **WHEN** a Skill directory exists at `~/.config/k8s-agent/skills/k8s-inspection/`
- **THEN** the system SHALL load the Skill if `~/.config/k8s-agent/skills/k8s-inspection/SKILL.md` exists

#### Scenario: Missing SKILL.md
- **WHEN** a Skill directory exists but contains no SKILL.md file
- **THEN** the system SHALL skip that directory and log a warning