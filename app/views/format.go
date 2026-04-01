package views

import (
	"fmt"
	"strings"
	"time"

	"sshbot/sdk/hubrelay"
)

func formatTime(value time.Time) string {
	if value.IsZero() {
		return "-"
	}
	return value.UTC().Format(time.RFC3339)
}

func boolLabel(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}

func healthVariant(value bool) string {
	if value {
		return "primary"
	}
	return "outline"
}

func gatewaySummary(gateway hubrelay.GatewayStatus) string {
	parts := []string{
		"WG=" + boolLabel(gateway.Levels.WG.OK),
		"transport=" + boolLabel(gateway.Levels.Transport.OK),
		"business=" + boolLabel(gateway.Levels.Business.OK),
	}
	return strings.Join(parts, " | ")
}

func stringify(value any) string {
	return fmt.Sprintf("%v", value)
}
