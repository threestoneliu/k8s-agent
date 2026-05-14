// Copyright 2026 threestoneliu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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