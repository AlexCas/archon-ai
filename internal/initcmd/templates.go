package initcmd

import (
	"bytes"
	"fmt"
	"text/template"
)

const agentsTemplate = `# ARCHON AI Orchestrator

## Phase Order
explore → propose → spec → design → tasks → apply → verify → judge → archive

## Rules
1. Check harness-workflow before any phase transition
2. Delegate each phase to sdd-* sub-agent
3. After verify, invoke harness-judge
4. On judge fail: re-apply with feedback (max 3 retries)

## Configuration
- Skills: {{.SkillCount}} (embedded via archon init)
- Config: .archon/config.yaml
- Agent: {{.Agent}}
- Harness Version: {{.HarnessVersion}}

## State Management
State tracked in: openspec/changes/{change-name}/state.yaml
Transitions validated by harness-workflow skill
`

const claudeTemplate = `# ARCHON AI Orchestrator

## Phase Order
explore → propose → spec → design → tasks → apply → verify → judge → archive

## Rules
1. Check harness-workflow before any phase transition
2. Delegate each phase to sdd-* sub-agent
3. After verify, invoke harness-judge
4. On judge fail: re-apply with feedback (max 3 retries)

## Configuration
- Skills: {{.SkillCount}} (embedded via archon init)
- Config: .archon/config.yaml
- Agent: {{.Agent}}
- Harness Version: {{.HarnessVersion}}

## State Management
State tracked in: openspec/changes/{change-name}/state.yaml
Transitions validated by harness-workflow skill
`

type TemplateData struct {
	ProjectName    string
	Agent          string
	HarnessVersion string
	SkillCount     int
}

func RenderAgentsMD(data TemplateData) (string, error) {
	return renderTemplate("AGENTS.md", agentsTemplate, data)
}

func RenderClaudeMD(data TemplateData) (string, error) {
	return renderTemplate("CLAUDE.md", claudeTemplate, data)
}

func renderTemplate(name, tmplContent string, data TemplateData) (string, error) {
	tmpl, err := template.New(name).Parse(tmplContent)
	if err != nil {
		return "", fmt.Errorf("parse template %s: %w", name, err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("execute template %s: %w", name, err)
	}

	return buf.String(), nil
}
