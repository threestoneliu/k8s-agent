## ADDED Requirements

### Requirement: SKILL.md Format Validation
The system SHALL validate SKILL.md files during loading. A valid SKILL.md MUST contain YAML frontmatter with `name` and `description` fields, followed by Markdown body content.

#### Scenario: Valid SKILL.md with YAML frontmatter
- **WHEN** a SKILL.md file contains valid YAML frontmatter with `name` and `description`
- **THEN** the system SHALL load the Skill successfully and register it for discovery

#### Scenario: Missing YAML frontmatter
- **WHEN** a SKILL.md file lacks YAML frontmatter
- **THEN** the system SHALL reject the Skill and log an error indicating invalid format

#### Scenario: Missing required frontmatter fields
- **WHEN** a SKILL.md file has YAML frontmatter but lacks `name` or `description`
- **THEN** the system SHALL reject the Skill and log an error indicating missing required fields

### Requirement: Skill Workflow Execution
The system SHALL execute Skill workflows through existing Function Calling mechanisms. The LLM SHALL follow the steps defined in SKILL.md's Markdown body by calling registered functions.

#### Scenario: LLM executes inspection workflow
- **WHEN** the LLM has read a Skill's SKILL.md defining an inspection workflow
- **THEN** the LLM SHALL call `resource_list` for nodes, then for pods, then for events, following the defined sequence

#### Scenario: Workflow step references unknown function
- **WHEN** a Skill workflow step references a function that is not registered
- **THEN** the LLM SHALL handle the error gracefully and inform the user that the workflow cannot be completed

### Requirement: Workflow Step Constraints
The system SHALL constrain Skill workflow execution to registered functions only. Skills MUST NOT introduce new function types; they MUST use existing functions like `resource_list`, `resource_get`, etc.

#### Scenario: Skill uses only registered functions
- **WHEN** a Skill's workflow only references registered functions
- **THEN** the workflow SHALL execute successfully through the existing Function Calling system

#### Scenario: Skill defines custom function
- **WHEN** a Skill workflow attempts to define or reference a custom function not in the registry
- **THEN** the system SHALL reject the Skill during loading and log an error