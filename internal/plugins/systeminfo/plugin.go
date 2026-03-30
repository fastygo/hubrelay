package systeminfo

import (
	"context"
	"fmt"
	"runtime"
	"strings"

	"github.com/shirou/gopsutil/v4/disk"
	"github.com/shirou/gopsutil/v4/host"
	"github.com/shirou/gopsutil/v4/mem"

	"sshbot/internal/buildprofile"
	"sshbot/internal/core"
)

func Builtins() []core.Plugin {
	return []core.Plugin{
		CapabilitiesPlugin{},
		SystemInfoPlugin{},
		AuditPlugin{},
	}
}

type CapabilitiesPlugin struct{}

func (CapabilitiesPlugin) Descriptor() core.PluginDescriptor {
	return core.PluginDescriptor{
		Name:    "capabilities",
		Summary: "Describe immutable profile capabilities",
		RequiredCapabilities: []buildprofile.Capability{
			buildprofile.CapabilityPluginCapabilities,
		},
	}
}

func (CapabilitiesPlugin) Execute(_ context.Context, cmdCtx core.CommandContext, _ core.CommandEnvelope) (core.CommandResult, error) {
	capabilities := make([]string, 0, len(cmdCtx.Profile.Capabilities))
	for _, capability := range cmdCtx.Profile.Capabilities {
		capabilities = append(capabilities, string(capability))
	}

	return core.CommandResult{
		Status:  "ok",
		Message: fmt.Sprintf("profile %s exposes %d immutable capabilities", cmdCtx.Profile.ID, len(capabilities)),
		Data: map[string]any{
			"profile_id":     cmdCtx.Profile.ID,
			"display_name":   cmdCtx.Profile.DisplayName,
			"capabilities":   capabilities,
			"http_bind":      cmdCtx.Profile.HTTPChat.BindAddress,
			"email_enabled":  cmdCtx.Profile.Email.Enabled,
			"ai_enabled":     cmdCtx.Profile.OpenAI.Enabled,
			"ai_provider":    cmdCtx.Profile.OpenAI.Provider,
			"ai_base_url":    cmdCtx.Profile.OpenAI.BaseURL,
			"ai_model":       cmdCtx.Profile.OpenAI.Model,
			"ai_api_mode":    cmdCtx.Profile.OpenAI.APIMode,
			"chat_history":   cmdCtx.Profile.OpenAI.ChatHistory,
			"ai_has_api_key": cmdCtx.Profile.OpenAI.HasAPIKey,
			"proxy_session":  cmdCtx.Profile.ProxySession.Enabled,
			"proxy_force":    cmdCtx.Profile.ProxySession.Force,
		},
	}, nil
}

type SystemInfoPlugin struct{}

func (SystemInfoPlugin) Descriptor() core.PluginDescriptor {
	return core.PluginDescriptor{
		Name:    "system-info",
		Summary: "Return safe host inspection data",
		RequiredCapabilities: []buildprofile.Capability{
			buildprofile.CapabilityPluginSystemInfo,
		},
	}
}

func (SystemInfoPlugin) Execute(_ context.Context, _ core.CommandContext, _ core.CommandEnvelope) (core.CommandResult, error) {
	hostInfo, _ := host.Info()
	memInfo, _ := mem.VirtualMemory()
	diskInfo, _ := disk.Usage("/")

	data := map[string]any{
		"go_os":      runtime.GOOS,
		"go_arch":    runtime.GOARCH,
		"host_name":  "",
		"platform":   "",
		"kernel":     "",
		"uptime_sec": uint64(0),
	}
	if hostInfo != nil {
		data["host_name"] = hostInfo.Hostname
		data["platform"] = strings.TrimSpace(hostInfo.Platform + " " + hostInfo.PlatformVersion)
		data["kernel"] = hostInfo.KernelVersion
		data["uptime_sec"] = hostInfo.Uptime
	}
	if memInfo != nil {
		data["memory_total_mb"] = memInfo.Total / 1024 / 1024
		data["memory_used_mb"] = memInfo.Used / 1024 / 1024
		data["memory_used_percent"] = memInfo.UsedPercent
	}
	if diskInfo != nil {
		data["disk_total_gb"] = diskInfo.Total / 1024 / 1024 / 1024
		data["disk_used_gb"] = diskInfo.Used / 1024 / 1024 / 1024
		data["disk_used_percent"] = diskInfo.UsedPercent
	}

	return core.CommandResult{
		Status:  "ok",
		Message: "safe server inspection data collected",
		Data:    data,
	}, nil
}

type AuditPlugin struct{}

func (AuditPlugin) Descriptor() core.PluginDescriptor {
	return core.PluginDescriptor{
		Name:    "audit",
		Summary: "Return recent audit entries",
		RequiredCapabilities: []buildprofile.Capability{
			buildprofile.CapabilityPluginAudit,
		},
	}
}

func (AuditPlugin) Execute(_ context.Context, cmdCtx core.CommandContext, envelope core.CommandEnvelope) (core.CommandResult, error) {
	limit := 10
	if raw, ok := envelope.Args["limit"]; ok && raw != "" {
		fmt.Sscanf(raw, "%d", &limit)
	}
	entries, err := cmdCtx.Store.ListRecentAudit(limit)
	if err != nil {
		return core.CommandResult{}, err
	}

	items := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		items = append(items, map[string]any{
			"command":   entry.CommandName,
			"principal": entry.PrincipalID,
			"transport": entry.Transport,
			"outcome":   entry.Outcome,
			"message":   entry.Message,
			"at":        entry.RecordedAt,
		})
	}

	return core.CommandResult{
		Status:  "ok",
		Message: fmt.Sprintf("returned %d audit entries", len(items)),
		Data: map[string]any{
			"entries": items,
		},
	}, nil
}
