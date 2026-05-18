## ADDED Requirements

### Requirement: ResponseParser Interface

The `llm` package SHALL provide a `ResponseParser` interface that parses LLM output text into structured parts containing think content and regular text content.

#### Scenarios

- **Scenario: Interface contract**
  - Given a `ResponseParser` implementation
  - When `Parse(text)` is called with OpenAI response containing think tags
  - Then it returns `[]TextPart` with correct `IsThink` and `Content` fields

### Requirement: TextPart Structure

The parser SHALL return a slice of `TextPart` structs, each containing:
- `IsThink`: boolean indicating if this is think content
- `Content`: the text content of the part

#### Scenarios

- **Scenario: TextPart fields**
  - Given parsing OpenAI response `<think>thinking text`
  - When `Parse` is called
  - Then returned `TextPart[0].IsThink == true` and `TextPart[0].Content == "thinking"`
  - And `TextPart[1].IsThink == false` and `TextPart[1].Content == " text"`

### Requirement: OpenAI Response Parser

The `llm` package SHALL provide an `OpenAIResponseParser` implementation that parses OpenAI's `<think>` `` XML tags.

#### Scenarios

- **Scenario: Parse think tags**
  - Given text `Hello <think> world Goodbye`
  - When `OpenAIResponseParser.Parse` is called
  - Then it returns 3 parts: text, think, text

- **Scenario: Handle unclosed think tag**
  - Given text `Hello <think> world`
  - When `OpenAIResponseParser.Parse` is called
  - Then it returns 1 part with entire text as non-think

- **Scenario: Handle multiple think tags**
  - Given text `<think> A text <think> B more`
  - When `OpenAIResponseParser.Parse` is called
  - Then it returns 4 parts with correct IsThink values

### Requirement: Service Parser Access

The `Service` type SHALL provide a `ResponseParser() ResponseParser` method that returns the default parser for the service.

#### Scenarios

- **Scenario: Service returns parser**
  - Given a `Service` instance
  - When `ResponseParser()` is called
  - Then it returns an `OpenAIResponseParser` instance

---

## MODIFIED Requirements

None.

---

## REMOVED Requirements

None.

---

## RENAMED Requirements

None.