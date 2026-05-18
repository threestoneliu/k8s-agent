## ADDED Requirements

### Requirement: ResponseParser Interface

The `llm` package SHALL provide a `ResponseParser` interface that parses LLM output text into structured parts containing think content and regular text content.

### Requirement: TextPart Structure

The parser SHALL return a slice of `TextPart` structs, each containing:
- `IsThink`: boolean indicating if this is think content
- `Content`: the text content of the part

### Requirement: OpenAI Response Parser

The `llm` package SHALL provide an `OpenAIResponseParser` implementation that parses OpenAI's `<think>` `</think>` XML tags.

### Requirement: Service Parser Access

The `Service` type SHALL provide a `ResponseParser() ResponseParser` method that returns the default parser for the service.

---

## MODIFIED Requirements

None.

---

## REMOVED Requirements

None.

---

## RENAMED Requirements

None.