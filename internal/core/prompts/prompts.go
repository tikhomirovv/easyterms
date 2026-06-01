// Package prompts builds LLM prompt text for ingest and analysis modes.
// Templates live in core; adapters only send the resulting messages.
package prompts

import (
	"fmt"
	"strings"

	"github.com/tikhomirovv/easyterms/internal/core/ports"
)

const defaultVersion = "v1"

// ExtractMessages returns system and user prompts for clean-text extraction.
func ExtractMessages(req ports.ExtractRequest, version string) (system, user string) {
	if version == "" {
		version = defaultVersion
	}
	system = fmt.Sprintf(`You extract and normalize legal agreement text (version %s).
Return only the cleaned plain text of the agreement in the user's language (%s).
Do not add commentary, markdown, or legal advice.`, version, req.Locale)

	var parts []string
	if req.RawText != "" {
		parts = append(parts, req.RawText)
	}
	if req.URL != "" {
		parts = append(parts, fmt.Sprintf("Source URL: %s", req.URL))
	}
	user = strings.Join(parts, "\n\n")
	return system, user
}

// AnalyzeMessages returns system and user prompts for an analysis mode.
func AnalyzeMessages(req ports.AnalyzeRequest, version string) (system, user string, jsonMode bool) {
	if version == "" {
		version = req.PromptVersion
	}
	if version == "" {
		version = defaultVersion
	}

	lang := req.Locale
	if lang == "" {
		lang = "en"
	}

	switch req.AnalysisType {
	case "plain":
		system = fmt.Sprintf(`You explain online legal agreements in plain language (version %s).
Respond in %s. Be neutral: do not tell the user to accept or reject.
Output valid JSON: {"summary": "<full plain-language explanation>"}`, version, lang)
		user = req.CleanText
		return system, user, true
	case "highlights":
		system = fmt.Sprintf(`You highlight important or risky clauses in online agreements (version %s).
Respond in %s. Be neutral. Focus on privacy, payments, liability, data, account deletion.
Output valid JSON: {"highlights": [{"title": "...", "explanation": "...", "severity": "low|medium|high"}]}`, version, lang)
		user = req.CleanText
		return system, user, true
	default:
		system = fmt.Sprintf(`Analyze the agreement (mode %q, version %s) in %s.
Output valid JSON with your analysis.`, req.AnalysisType, version, lang)
		user = req.CleanText
		return system, user, true
	}
}
