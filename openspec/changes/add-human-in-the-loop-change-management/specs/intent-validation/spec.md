# Intent Validation Specification

## ADDED Requirements

### Requirement: Valid Action Types

The ParsedIntent SHALL have an Action that is one of: CREATE, UPDATE, DELETE, INSPECT

#### Scenario: Valid Action Types Accepted

Given a ParsedIntent with Action=CREATE
When ValidateIntent is called
Then the result SHALL be nil (no ClarifyQuestion)

### Requirement: Required Target Fields

The ParsedIntent SHALL have required Target fields:
- Target.Kind is always required
- Target.Name is required for UPDATE/DELETE/INSPECT
- Target.Namespace is required for namespaced resource kinds

#### Scenario: Missing Name Returns ClarifyQuestion

Given a ParsedIntent with Action=UPDATE, Target.Kind="Deployment" but no Target.Name
When ValidateIntent is called
Then a ClarifyQuestion SHALL be returned with Field="target.name"

### Requirement: Risk Level Defaults

The system SHALL assign default RiskLevel based on action:
- CREATE operations SHALL default to LOW risk
- UPDATE operations SHALL default to MEDIUM risk
- DELETE operations SHALL default to HIGH risk
- INSPECT operations SHALL default to LOW risk

#### Scenario: DELETE Defaults to HIGH Risk

Given a ParsedIntent with Action=DELETE and no RiskLevel set
When DefaultRiskLevel is called
Then the result SHALL be HIGH

### Requirement: Reason Required for High Risk

When RiskLevel >= HIGH, the ParsedIntent SHALL have a non-empty Reason field.

#### Scenario: High Risk Without Reason Returns ClarifyQuestion

Given a ParsedIntent with RiskLevel=HIGH but empty Reason
When ValidateIntent is called
Then a ClarifyQuestion SHALL be returned with Field="reason"

---

## Overview

Core validates ParsedIntent structure completeness and generates ClarifyQuestion when intent is ambiguous.

## ParsedIntent Validation

### Required Fields

| Field | Validation Rule |
|-------|------------------|
| Action | Must be one of: CREATE, UPDATE, DELETE, INSPECT |
| Target | Must have Kind specified |
| Target.Name | Required for UPDATE/DELETE/INSPECT |
| Target.Namespace | Required if Kind is namespaced |
| RiskLevel | Must be one of: LOW, MEDIUM, HIGH, CRITICAL |
| Reason | Required if RiskLevel >= HIGH |

### Risk Level Defaults

| Operation | Default RiskLevel |
|-----------|-------------------|
| CREATE with standard params | LOW |
| UPDATE to status/scale | MEDIUM |
| UPDATE to spec/label/annotation | MEDIUM |
| DELETE any resource | HIGH |
| CRITICAL operation without reason | HIGH |

## ClarifyQuestion

### Structure

```go
type ClarifyQuestion struct {
    Field         string   // dot-notation path to missing field
    Question      string   // natural language question
    Options       []string // multiple choice options (optional)
    Required      bool     // true = blocking, false = optional
}
```

### Question Generation Rules

| Condition | Question |
|-----------|----------|
| Target.Name empty | "目标资源的名称是什么？" |
| Target.Namespace empty | "目标 namespace 是哪个？" |
| Reason empty for HIGH risk | "这个操作的原因是什么？（高风险操作需要明确原因）" |
| Params empty for UPDATE | "需要修改哪些参数？" |

## Clarify Flow

1. Core receives ParsedIntent from Agent
2. Core validates structure completeness
3. If incomplete → return ClarifyQuestion(s)
4. Agent presents question(s) to user
5. User provides answer(s)
6. Agent updates ParsedIntent with answers
7. Core re-validates → loop until complete

## Blocking vs Optional Questions

- **Blocking (Required: true)**: State machine cannot advance until answered
- **Optional (Required: false)**: Can be skipped with user confirmation