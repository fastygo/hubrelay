package safety

import (
	"regexp"
	"slices"
	"strings"
	"unicode"
)

type Severity string

const (
	SeverityLow    Severity = "low"
	SeverityMedium Severity = "medium"
	SeverityHigh   Severity = "high"
)

type Finding struct {
	RuleCode string   `json:"rule_code"`
	Type     string   `json:"type"`
	Severity Severity `json:"severity"`
	Excerpt  string   `json:"excerpt"`
	Hint     string   `json:"hint"`
}

type Report struct {
	Blocked  bool      `json:"blocked"`
	Findings []Finding `json:"findings,omitempty"`
}

type regexRule struct {
	code     string
	kind     string
	severity Severity
	hint     string
	pattern  *regexp.Regexp
}

var regexRules = []regexRule{
	{
		code:     "email",
		kind:     "email",
		severity: SeverityMedium,
		hint:     "Replace personal or work email addresses with example placeholders.",
		pattern:  regexp.MustCompile(`(?i)\b[A-Z0-9._%+\-]+@[A-Z0-9.\-]+\.[A-Z]{2,}\b`),
	},
	{
		code:     "authorization_bearer",
		kind:     "bearer_token",
		severity: SeverityHigh,
		hint:     "Remove bearer tokens and replace them with <ACCESS_TOKEN>.",
		pattern:  regexp.MustCompile(`(?i)(authorization\s*:\s*bearer\s+[A-Z0-9._\-]+|\bbearer\s+[A-Z0-9._\-]{16,})`),
	},
	{
		code:     "jwt",
		kind:     "jwt",
		severity: SeverityHigh,
		hint:     "Do not paste real JWTs. Replace them with <JWT_TOKEN>.",
		pattern:  regexp.MustCompile(`\beyJ[A-Za-z0-9_-]{5,}\.[A-Za-z0-9._-]{10,}\.[A-Za-z0-9._-]{10,}\b`),
	},
	{
		code:     "private_key_block",
		kind:     "private_key",
		severity: SeverityHigh,
		hint:     "Do not send private key material. Replace it with a short description instead.",
		pattern:  regexp.MustCompile(`-----BEGIN [A-Z ]+PRIVATE KEY-----`),
	},
	{
		code:     "openai_key",
		kind:     "api_key",
		severity: SeverityHigh,
		hint:     "Replace real API keys with placeholders such as <OPENAI_API_KEY>.",
		pattern:  regexp.MustCompile(`\bsk-[A-Za-z0-9]{12,}\b`),
	},
	{
		code:     "secret_assignment",
		kind:     "secret_assignment",
		severity: SeverityHigh,
		hint:     "Remove real password, secret, token, cookie, or api_key values before sending.",
		pattern:  regexp.MustCompile(`(?i)\b(password|passwd|secret|token|api[_-]?key|cookie|authorization)\b\s*[:=]\s*([^\s,;]+)`),
	},
	{
		code:     "credential_url",
		kind:     "credential_url",
		severity: SeverityHigh,
		hint:     "Do not send URLs with embedded credentials. Replace user:pass with placeholders.",
		pattern:  regexp.MustCompile(`https?://[^\s/@:]+:[^\s/@]+@`),
	},
	{
		code:     "ssh_public_key",
		kind:     "ssh_key",
		severity: SeverityMedium,
		hint:     "Mask SSH key material unless you are sharing a synthetic example.",
		pattern:  regexp.MustCompile(`\bssh-(rsa|ed25519)\s+[A-Za-z0-9+/=]{20,}`),
	},
}

var wordPattern = regexp.MustCompile(`[A-Za-z_][A-Za-z0-9_\-]{3,}`)

var riskyKeywords = []string{
	"password",
	"passwd",
	"secret",
	"token",
	"apikey",
	"api_key",
	"bearer",
	"cookie",
	"privatekey",
	"sessionid",
	"authorization",
}

var longTokenPattern = regexp.MustCompile(`[A-Za-z0-9_\-=/+]{24,}`)

func ScanText(input string) Report {
	text := strings.TrimSpace(input)
	if text == "" {
		return Report{}
	}

	report := Report{}
	seen := make(map[string]struct{})
	addFinding := func(finding Finding) {
		key := finding.RuleCode + ":" + finding.Excerpt
		if _, ok := seen[key]; ok {
			return
		}
		seen[key] = struct{}{}
		report.Findings = append(report.Findings, finding)
	}

	for _, rule := range regexRules {
		matches := rule.pattern.FindAllString(text, 8)
		for _, match := range matches {
			addFinding(Finding{
				RuleCode: rule.code,
				Type:     rule.kind,
				Severity: rule.severity,
				Excerpt:  maskExcerpt(match),
				Hint:     rule.hint,
			})
		}
	}

	for _, token := range wordPattern.FindAllString(strings.ToLower(text), -1) {
		normalized := normalizeWord(token)
		if normalized == "" {
			continue
		}
		for _, keyword := range riskyKeywords {
			if normalized == keyword || levenshtein(normalized, keyword) <= 2 {
				addFinding(Finding{
					RuleCode: "risky_keyword",
					Type:     "keyword",
					Severity: SeverityMedium,
					Excerpt:  maskExcerpt(token),
					Hint:     "Replace risky words and surrounding values with placeholders before sending.",
				})
				break
			}
		}
	}

	for _, token := range longTokenPattern.FindAllString(text, -1) {
		if looksLikeRandomSecret(token) {
			addFinding(Finding{
				RuleCode: "long_random_token",
				Type:     "high_entropy_secret",
				Severity: SeverityMedium,
				Excerpt:  maskExcerpt(token),
				Hint:     "Mask long random-looking tokens unless they are synthetic examples.",
			})
		}
	}

	if len(report.Findings) > 0 {
		report.Blocked = true
		slices.SortStableFunc(report.Findings, func(a, b Finding) int {
			if a.Severity != b.Severity {
				return compareSeverity(a.Severity, b.Severity)
			}
			return strings.Compare(a.RuleCode, b.RuleCode)
		})
	}

	return report
}

func (r Report) RuleCodes() []string {
	if len(r.Findings) == 0 {
		return nil
	}
	rules := make([]string, 0, len(r.Findings))
	seen := make(map[string]struct{})
	for _, finding := range r.Findings {
		if _, ok := seen[finding.RuleCode]; ok {
			continue
		}
		seen[finding.RuleCode] = struct{}{}
		rules = append(rules, finding.RuleCode)
	}
	slices.Sort(rules)
	return rules
}

func compareSeverity(left, right Severity) int {
	score := func(value Severity) int {
		switch value {
		case SeverityHigh:
			return 0
		case SeverityMedium:
			return 1
		default:
			return 2
		}
	}
	return score(left) - score(right)
}

func normalizeWord(value string) string {
	var builder strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(unicode.ToLower(r))
		}
	}
	return builder.String()
}

func looksLikeRandomSecret(token string) bool {
	if strings.Contains(token, "://") || strings.Contains(token, "@") {
		return false
	}
	hasLower := false
	hasUpper := false
	hasDigit := false
	hasSpecial := false
	for _, r := range token {
		switch {
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsDigit(r):
			hasDigit = true
		case strings.ContainsRune("_-/+=", r):
			hasSpecial = true
		}
	}
	categories := 0
	for _, present := range []bool{hasLower, hasUpper, hasDigit, hasSpecial} {
		if present {
			categories++
		}
	}
	return len(token) >= 24 && categories >= 3
}

func maskExcerpt(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if len(trimmed) <= 8 {
		return trimmed[:1] + strings.Repeat("*", max(len(trimmed)-2, 1)) + trimmed[len(trimmed)-1:]
	}
	return trimmed[:4] + "..." + trimmed[len(trimmed)-4:]
}

func max(left, right int) int {
	if left > right {
		return left
	}
	return right
}

func levenshtein(left, right string) int {
	if left == right {
		return 0
	}
	if left == "" {
		return len(right)
	}
	if right == "" {
		return len(left)
	}

	prev := make([]int, len(right)+1)
	curr := make([]int, len(right)+1)
	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(left); i++ {
		curr[0] = i
		for j := 1; j <= len(right); j++ {
			cost := 0
			if left[i-1] != right[j-1] {
				cost = 1
			}
			deletion := prev[j] + 1
			insertion := curr[j-1] + 1
			substitution := prev[j-1] + cost
			curr[j] = min(deletion, insertion, substitution)
		}
		prev, curr = curr, prev
	}

	return prev[len(right)]
}

func min(values ...int) int {
	best := values[0]
	for _, value := range values[1:] {
		if value < best {
			best = value
		}
	}
	return best
}
