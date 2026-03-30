package safety

import "testing"

func TestScanTextFindsCommonSensitivePatterns(t *testing.T) {
	report := ScanText("email=alice@example.com\nAuthorization: Bearer sk-abcdef1234567890")
	if !report.Blocked {
		t.Fatalf("expected report to block sensitive input")
	}
	if len(report.Findings) < 2 {
		t.Fatalf("expected multiple findings, got %d", len(report.Findings))
	}
	rules := report.RuleCodes()
	if len(rules) == 0 {
		t.Fatalf("expected rule codes to be populated")
	}
}

func TestScanTextFindsNearMissKeyword(t *testing.T) {
	report := ScanText("please review this passwrod=supersecretvalue")
	if !report.Blocked {
		t.Fatalf("expected fuzzy keyword match to block input")
	}
}

func TestScanTextAvoidsOrdinaryPrompt(t *testing.T) {
	report := ScanText("Summarize the deployment strategy for a bot running in a read-only container.")
	if report.Blocked {
		t.Fatalf("expected ordinary text to pass without findings: %+v", report.Findings)
	}
}
