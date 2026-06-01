package prompts_test

import (
	"strings"
	"testing"

	"github.com/tikhomirovv/easyterms/internal/core/prompts"
	"github.com/tikhomirovv/easyterms/internal/core/ports"
)

func TestExtractMessages_includesLocaleAndText(t *testing.T) {
	sys, user := prompts.ExtractMessages(ports.ExtractRequest{
		RawText: "Terms here",
		Locale:  "ru",
	}, "v1")
	if !strings.Contains(sys, "ru") {
		t.Fatalf("system = %q", sys)
	}
	if user != "Terms here" {
		t.Fatalf("user = %q", user)
	}
}

func TestAnalyzeMessages_plainJSONMode(t *testing.T) {
	sys, user, jsonMode := prompts.AnalyzeMessages(ports.AnalyzeRequest{
		CleanText:    "doc",
		AnalysisType: "plain",
		Locale:       "en",
	}, "v1")
	if !jsonMode {
		t.Fatal("expected json mode")
	}
	if !strings.Contains(sys, "plain language") {
		t.Fatalf("system = %q", sys)
	}
	if user != "doc" {
		t.Fatalf("user = %q", user)
	}
}
