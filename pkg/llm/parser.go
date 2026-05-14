package llm

import "strings"

// TextPart represents a part of LLM response text
type TextPart struct {
	IsThink  bool
	Content  string
}

// ResponseParser parses LLM output text
type ResponseParser interface {
	Parse(text string) []TextPart
}

// OpenAIResponseParser parses OpenAI's <think> XML tags
type OpenAIResponseParser struct{}

func (p *OpenAIResponseParser) Parse(text string) []TextPart {
	parts := []TextPart{}
	thinkStart := "<think>"
	thinkEnd := "</think>"

	for {
		startIdx := strings.Index(text, thinkStart)
		if startIdx == -1 {
			if len(text) > 0 {
				parts = append(parts, TextPart{IsThink: false, Content: text})
			}
			break
		}

		if startIdx > 0 {
			parts = append(parts, TextPart{IsThink: false, Content: text[:startIdx]})
		}

		endIdx := strings.Index(text[startIdx:], thinkEnd)
		if endIdx == -1 {
			parts = append(parts, TextPart{IsThink: false, Content: text[startIdx:]})
			break
		}

		endIdx += startIdx + len(thinkEnd)
		thinkContent := text[startIdx+len(thinkStart) : endIdx-len(thinkEnd)]
		parts = append(parts, TextPart{IsThink: true, Content: strings.TrimSpace(thinkContent)})

		text = text[endIdx:]
	}

	return parts
}