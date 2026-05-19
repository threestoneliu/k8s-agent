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