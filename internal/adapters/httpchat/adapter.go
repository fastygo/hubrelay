package httpchat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"sshbot/internal/buildprofile"
	"sshbot/internal/core"
	proxymgr "sshbot/internal/proxy"
)

const maxRequestBodyBytes = 1 << 20 // 1 MiB

type Adapter struct {
	bindAddress  string
	service      *core.Service
	proxy        *proxymgr.Manager
	server       *http.Server
	indexTmpl    *template.Template
	indexTmplErr error
	tmplOnce     sync.Once
}

func New(bindAddress string, service *core.Service, proxyManager *proxymgr.Manager) *Adapter {
	return &Adapter{
		bindAddress: bindAddress,
		service:     service,
		proxy:       proxyManager,
	}
}

func (a *Adapter) Name() string {
	return "http_chat"
}

func (a *Adapter) Start(ctx context.Context) error {
	mux := a.buildMux()

	a.server = &http.Server{
		Addr:              a.bindAddress,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := a.server.Shutdown(shutdownCtx); err != nil {
			log.Printf("http chat shutdown failed: %v", err)
		}
	}()

	err := a.server.ListenAndServe()
	if err != nil && err != http.ErrServerClosed {
		return err
	}
	return nil
}

func (a *Adapter) buildMux() *http.ServeMux {
	mux := http.NewServeMux()

	// Use exact and method-specific patterns to avoid ServeMux conflicts.
	mux.HandleFunc("GET /{$}", a.handleIndex)
	mux.HandleFunc("GET /healthz", a.handleHealth)
	mux.HandleFunc("POST /api/command", a.handleCommand)
	mux.HandleFunc("POST /api/command/stream", a.handleCommandStream)
	mux.HandleFunc("POST /api/proxy/session", a.handleProxyCollection)
	mux.HandleFunc("GET /api/proxy/session/", a.handleProxySession)
	mux.HandleFunc("POST /api/proxy/session/", a.handleProxySession)

	return mux
}

func (a *Adapter) handleHealth(writer http.ResponseWriter, _ *http.Request) {
	payload := map[string]any{
		"status":  "ok",
		"adapter": a.Name(),
		"profile": a.service.Profile().ID,
	}
	writeJSON(writer, http.StatusOK, payload)
}

func (a *Adapter) handleIndex(writer http.ResponseWriter, _ *http.Request) {
	type pageData struct {
		ProfileID    string
		AIEnabled    bool
		ChatHistory  bool
		ProxyEnabled bool
		ProxyForced  bool
		DefaultModel string
	}

	page := a.getIndexTemplate(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <title>Autonomous Bot</title>
  <style>
    body { font-family: Arial, sans-serif; margin: 2rem auto; max-width: 860px; padding: 0 1rem; }
    textarea, input { width: 100%; font-family: monospace; box-sizing: border-box; }
    textarea { min-height: 120px; }
    button { margin-top: 1rem; padding: 0.6rem 1rem; }
    pre { background: #f4f4f4; padding: 1rem; overflow: auto; }
    .panel { border: 1px solid #ddd; border-radius: 8px; padding: 1rem; margin-top: 1rem; }
    .chat-log { min-height: 220px; max-height: 420px; overflow: auto; background: #fafafa; border: 1px solid #ddd; padding: 1rem; border-radius: 8px; }
    .chat-message { margin-bottom: 1rem; padding-bottom: 1rem; border-bottom: 1px solid #eee; }
    .chat-role { font-weight: bold; margin-bottom: 0.35rem; }
    .inline-actions { display: flex; gap: 0.75rem; flex-wrap: wrap; }
    .muted { color: #666; font-size: 0.9rem; }
    .proxy-results { margin-top: 1rem; }
    .proxy-item { border: 1px solid #ddd; border-radius: 8px; padding: 0.75rem; margin-bottom: 0.75rem; }
    .proxy-item.active { border-color: #0b7285; background: #f1fbfd; }
    dialog { width: min(760px, calc(100vw - 2rem)); border: 1px solid #ddd; border-radius: 10px; padding: 1rem 1.25rem; }
    dialog::backdrop { background: rgba(0, 0, 0, 0.25); }
  </style>
</head>
<body>
  <h1>Autonomous Bot</h1>
  <p>Loopback HTTP adapter intended for SSH tunneling.</p>
  <p class="muted">Profile: {{.ProfileID}} | AI enabled: {{.AIEnabled}} | Browser chat history: {{.ChatHistory}} | Proxy session: {{.ProxyEnabled}} | Proxy forced: {{.ProxyForced}}</p>

  <div class="panel">
    <h2>Command Block</h2>
    <form id="command-form">
      <label>Principal ID</label>
      <input name="principal" value="operator-local">
      <label>Roles (comma-separated)</label>
      <input name="roles" value="operator">
      <label>Command</label>
      <input name="command" value="capabilities">
      <button type="submit">Run command</button>
    </form>
    <h3>Command Result</h3>
    <pre id="output">{}</pre>
  </div>

  {{if .ProxyEnabled}}
  <div class="panel">
    <h2>Proxy Session</h2>
    <p class="muted">Proxy pools stay only in browser sessionStorage and server memory. They are not stored in chat history or BoltDB.</p>
    <div class="inline-actions">
      <button type="button" id="open-proxy">Configure Proxy Session</button>
      <button type="button" id="refresh-proxy">Refresh lease</button>
      <button type="button" id="clear-proxy">Clear browser proxy state</button>
    </div>
    <pre id="proxy-summary">{"status":"idle"}</pre>
  </div>

  <dialog id="proxy-dialog">
    <h3>Configure Proxy Session</h3>
    <label>SOCKS proxy pool (host:port per line)</label>
    <textarea id="proxy-list" placeholder="127.0.0.1:1080"></textarea>
    <div class="inline-actions">
      <button type="button" id="create-proxy-session">Create session</button>
      <button type="button" id="check-proxy-session">Check proxies</button>
      <button type="button" id="select-fastest-proxy">Select fastest</button>
      <button type="button" id="close-proxy">Close</button>
    </div>
    <div id="proxy-results" class="proxy-results"></div>
  </dialog>
  {{end}}

  <div class="panel">
    <h2>MVP Chat</h2>
    <p class="muted">This chat stores messages only in the browser. Server-side transcript persistence is disabled in MVP.</p>
    <div id="chat-log" class="chat-log"></div>
    <form id="ai-form">
      <label>Prompt</label>
      <textarea name="prompt" placeholder="Ask something..."></textarea>
      <label>Model override (optional)</label>
      <input name="model" value="{{.DefaultModel}}">
      <button type="submit" {{if not .AIEnabled}}disabled{{end}}>Send to AI</button>
      <button type="button" id="export-chat">Export chat JSON</button>
    </form>
  </div>

  <dialog id="guard-dialog">
    <h3>Review Sensitive Data</h3>
    <p class="muted">The text looks like it may contain sensitive data. Review these findings before sending anything to the AI provider.</p>
    <div id="guard-findings"></div>
    <div class="inline-actions">
      <button type="button" id="guard-edit">Back and edit</button>
      <button type="button" id="guard-send">Send anyway</button>
    </div>
  </dialog>

  <script>
    const commandForm = document.getElementById('command-form');
    const aiForm = document.getElementById('ai-form');
    const output = document.getElementById('output');
    const chatLog = document.getElementById('chat-log');
    const exportButton = document.getElementById('export-chat');
    const guardDialog = document.getElementById('guard-dialog');
    const guardFindings = document.getElementById('guard-findings');
    const guardEditButton = document.getElementById('guard-edit');
    const guardSendButton = document.getElementById('guard-send');
    const openProxyButton = document.getElementById('open-proxy');
    const refreshProxyButton = document.getElementById('refresh-proxy');
    const clearProxyButton = document.getElementById('clear-proxy');
    const proxyDialog = document.getElementById('proxy-dialog');
    const proxyList = document.getElementById('proxy-list');
    const proxySummary = document.getElementById('proxy-summary');
    const proxyResults = document.getElementById('proxy-results');
    const createProxySessionButton = document.getElementById('create-proxy-session');
    const checkProxySessionButton = document.getElementById('check-proxy-session');
    const selectFastestProxyButton = document.getElementById('select-fastest-proxy');
    const closeProxyButton = document.getElementById('close-proxy');
    const pageConfig = {
      profileId: '{{.ProfileID}}',
      aiEnabled: {{if .AIEnabled}}true{{else}}false{{end}},
      chatHistory: {{if .ChatHistory}}true{{else}}false{{end}},
      proxyEnabled: {{if .ProxyEnabled}}true{{else}}false{{end}},
      proxyForced: {{if .ProxyForced}}true{{else}}false{{end}}
    };
    const storageKey = 'sshbot-chat-history:' + pageConfig.profileId;
    const proxyStorageKey = 'sshbot-proxy-session:' + pageConfig.profileId;
    let memoryMessages = [];
    let guardDecisionResolver = null;

    function escapeHTML(value) {
      const text = String(value || '');
      return text.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;').replace(/'/g, '&#39;');
    }

    function getRoles(form) {
      const roles = form.roles.value.trim();
      if (!roles) return [];
      return roles.split(',').map(item => item.trim()).filter(Boolean);
    }

    function loadMessages() {
      if (!pageConfig.chatHistory) return memoryMessages;
      try {
        const raw = localStorage.getItem(storageKey);
        return raw ? JSON.parse(raw) : [];
      } catch (_) {
        return [];
      }
    }

    function saveMessages(messages) {
      memoryMessages = messages;
      if (!pageConfig.chatHistory) return;
      localStorage.setItem(storageKey, JSON.stringify(messages));
    }

    function renderMessages() {
      const messages = loadMessages();
      if (!messages.length) {
        chatLog.innerHTML = '<div class="muted">No chat messages yet.</div>';
        return;
      }
      chatLog.innerHTML = messages.map(item =>
        '<div class="chat-message">' +
          '<div class="chat-role">' + escapeHTML(item.role) + '</div>' +
          '<div>' + escapeHTML(item.content).replaceAll('\n', '<br>') + '</div>' +
          '<div class="muted">' + escapeHTML(item.at) + '</div>' +
        '</div>'
      ).join('');
      chatLog.scrollTop = chatLog.scrollHeight;
    }

    async function sendCommand(payload) {
      const response = await fetch('/api/command', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(payload)
      });
      return response.json();
    }

    function normalizeSensitiveToken(value) {
      return String(value || '').toLowerCase().replace(/[^a-z0-9]/g, '');
    }

    function levenshtein(left, right) {
      if (left === right) return 0;
      if (!left) return right.length;
      if (!right) return left.length;
      const prev = Array.from({ length: right.length + 1 }, (_, index) => index);
      const curr = new Array(right.length + 1).fill(0);
      for (let i = 1; i <= left.length; i += 1) {
        curr[0] = i;
        for (let j = 1; j <= right.length; j += 1) {
          const cost = left[i - 1] === right[j - 1] ? 0 : 1;
          curr[j] = Math.min(
            prev[j] + 1,
            curr[j - 1] + 1,
            prev[j - 1] + cost
          );
        }
        for (let j = 0; j < prev.length; j += 1) prev[j] = curr[j];
      }
      return prev[right.length];
    }

    function maskExcerpt(value) {
      const text = String(value || '').trim();
      if (!text) return '';
      if (text.length <= 8) {
        return text.slice(0, 1) + '*'.repeat(Math.max(text.length - 2, 1)) + text.slice(-1);
      }
      return text.slice(0, 4) + '...' + text.slice(-4);
    }

    function pushFinding(findings, finding) {
      const key = finding.rule_code + ':' + finding.excerpt;
      if (findings.some(item => item.rule_code + ':' + item.excerpt === key)) return;
      findings.push(finding);
    }

    function scanSensitiveInput(text) {
      const value = String(text || '').trim();
      if (!value) return { blocked: false, findings: [] };

      const findings = [];
      const regexRules = [
        {
          rule_code: 'email',
          type: 'email',
          severity: 'medium',
          hint: 'Replace personal or work email addresses with example placeholders.',
          pattern: /\b[A-Z0-9._%+-]+@[A-Z0-9.-]+\.[A-Z]{2,}\b/gi
        },
        {
          rule_code: 'authorization_bearer',
          type: 'bearer_token',
          severity: 'high',
          hint: 'Remove bearer tokens and replace them with <ACCESS_TOKEN>.',
          pattern: /(authorization\s*:\s*bearer\s+[A-Z0-9._-]+|\bbearer\s+[A-Z0-9._-]{16,})/gi
        },
        {
          rule_code: 'jwt',
          type: 'jwt',
          severity: 'high',
          hint: 'Do not paste real JWTs. Replace them with <JWT_TOKEN>.',
          pattern: /\beyJ[A-Za-z0-9_-]{5,}\.[A-Za-z0-9._-]{10,}\.[A-Za-z0-9._-]{10,}\b/g
        },
        {
          rule_code: 'private_key_block',
          type: 'private_key',
          severity: 'high',
          hint: 'Do not send private key material. Replace it with a short description instead.',
          pattern: /-----BEGIN [A-Z ]+PRIVATE KEY-----/g
        },
        {
          rule_code: 'openai_key',
          type: 'api_key',
          severity: 'high',
          hint: 'Replace real API keys with placeholders such as <OPENAI_API_KEY>.',
          pattern: /\bsk-[A-Za-z0-9]{12,}\b/g
        },
        {
          rule_code: 'secret_assignment',
          type: 'secret_assignment',
          severity: 'high',
          hint: 'Remove real password, secret, token, cookie, or api_key values before sending.',
          pattern: /\b(password|passwd|secret|token|api[_-]?key|cookie|authorization)\b\s*[:=]\s*([^\s,;]+)/gi
        },
        {
          rule_code: 'credential_url',
          type: 'credential_url',
          severity: 'high',
          hint: 'Do not send URLs with embedded credentials. Replace user:pass with placeholders.',
          pattern: /https?:\/\/[^\s/@:]+:[^\s/@]+@/gi
        }
      ];

      regexRules.forEach(rule => {
        const matches = value.match(rule.pattern) || [];
        matches.slice(0, 8).forEach(match => {
          pushFinding(findings, {
            rule_code: rule.rule_code,
            type: rule.type,
            severity: rule.severity,
            excerpt: maskExcerpt(match),
            hint: rule.hint
          });
        });
      });

      const riskyKeywords = ['password', 'passwd', 'secret', 'token', 'apikey', 'api_key', 'bearer', 'cookie', 'privatekey', 'sessionid', 'authorization'];
      const words = value.match(/[A-Za-z_][A-Za-z0-9_-]{3,}/g) || [];
      words.forEach(word => {
        const normalized = normalizeSensitiveToken(word);
        if (!normalized) return;
        if (riskyKeywords.some(keyword => normalized === keyword || levenshtein(normalized, keyword) <= 2)) {
          pushFinding(findings, {
            rule_code: 'risky_keyword',
            type: 'keyword',
            severity: 'medium',
            excerpt: maskExcerpt(word),
            hint: 'Replace risky words and surrounding values with placeholders before sending.'
          });
        }
      });

      const longTokens = value.match(/[A-Za-z0-9_+=\/-]{24,}/g) || [];
      longTokens.forEach(token => {
        if (token.includes('://') || token.includes('@')) return;
        const hasLower = /[a-z]/.test(token);
        const hasUpper = /[A-Z]/.test(token);
        const hasDigit = /[0-9]/.test(token);
        const hasSpecial = /[_+=\/-]/.test(token);
        const categories = [hasLower, hasUpper, hasDigit, hasSpecial].filter(Boolean).length;
        if (categories >= 3) {
          pushFinding(findings, {
            rule_code: 'long_random_token',
            type: 'high_entropy_secret',
            severity: 'medium',
            excerpt: maskExcerpt(token),
            hint: 'Mask long random-looking tokens unless they are synthetic examples.'
          });
        }
      });

      findings.sort((left, right) => {
        const rank = value => value === 'high' ? 0 : value === 'medium' ? 1 : 2;
        if (rank(left.severity) !== rank(right.severity)) return rank(left.severity) - rank(right.severity);
        return left.rule_code.localeCompare(right.rule_code);
      });
      return {
        blocked: findings.length > 0,
        findings: findings
      };
    }

    function renderGuardFindings(report) {
      if (!guardFindings) return;
      if (!report.findings.length) {
        guardFindings.innerHTML = '<div class="muted">No findings.</div>';
        return;
      }
      guardFindings.innerHTML = report.findings.map(item =>
        '<div class="proxy-item">' +
          '<div><strong>' + escapeHTML(item.type) + '</strong> (' + escapeHTML(item.severity) + ')</div>' +
          '<div class="muted">' + escapeHTML(item.excerpt) + '</div>' +
          '<div>' + escapeHTML(item.hint) + '</div>' +
        '</div>'
      ).join('');
    }

    function confirmSensitiveSend(report) {
      if (!guardDialog) return Promise.resolve(true);
      renderGuardFindings(report);
      guardDialog.showModal();
      return new Promise(resolve => {
        guardDecisionResolver = resolve;
      });
    }

    function parseAskCommand(raw) {
      const trimmed = String(raw || '').trim();
      if (!trimmed) return { name: '', prompt: '' };
      const fields = trimmed.split(/\s+/);
      const name = (fields[0] || '').toLowerCase();
      if (name !== 'ask') return { name: name, prompt: '' };
      if (fields.length > 1 && !fields[1].includes('=')) {
        return { name: name, prompt: trimmed.slice(fields[0].length).trim() };
      }
      const promptMatch = trimmed.match(/\bprompt=([^\s]+)/i);
      return { name: name, prompt: promptMatch ? promptMatch[1] : '' };
    }

    function loadProxyState() {
      if (!pageConfig.proxyEnabled) {
        return { rawList: '', sessionId: '', results: [], selectedProxy: '', lease: null };
      }
      try {
        const raw = sessionStorage.getItem(proxyStorageKey);
        if (!raw) return { rawList: '', sessionId: '', results: [], selectedProxy: '', lease: null };
        const parsed = JSON.parse(raw);
        return {
          rawList: parsed.rawList || '',
          sessionId: parsed.sessionId || '',
          results: Array.isArray(parsed.results) ? parsed.results : [],
          selectedProxy: parsed.selectedProxy || '',
          lease: parsed.lease || null
        };
      } catch (_) {
        return { rawList: '', sessionId: '', results: [], selectedProxy: '', lease: null };
      }
    }

    function saveProxyState(state) {
      if (!pageConfig.proxyEnabled) return;
      sessionStorage.setItem(proxyStorageKey, JSON.stringify(state));
    }

    function setProxyState(partial) {
      const next = Object.assign(loadProxyState(), partial);
      saveProxyState(next);
      renderProxySummary();
      renderProxyResults();
    }

    function syncProxyTextarea() {
      if (proxyList) {
        proxyList.value = loadProxyState().rawList;
      }
    }

    function renderProxySummary() {
      if (!pageConfig.proxyEnabled || !proxySummary) return;
      const state = loadProxyState();
      proxySummary.textContent = JSON.stringify({
        session_id: state.sessionId || null,
        selected_proxy: state.selectedProxy || null,
        lease: state.lease,
        checked: state.results.length
      }, null, 2);
    }

    function renderProxyResults() {
      if (!pageConfig.proxyEnabled || !proxyResults) return;
      const state = loadProxyState();
      if (!state.results.length) {
        proxyResults.innerHTML = '<div class="muted">No checked proxies yet.</div>';
        return;
      }
      proxyResults.innerHTML = state.results.map(item => {
        const isActive = item.address === state.selectedProxy;
        const latency = typeof item.latency_ms === 'number' ? item.latency_ms + ' ms' : 'n/a';
        const note = item.last_error ? ' | ' + escapeHTML(item.last_error) : '';
        return '<div class="proxy-item' + (isActive ? ' active' : '') + '">' +
          '<div><strong>' + escapeHTML(item.address) + '</strong></div>' +
          '<div class="muted">status=' + escapeHTML(item.status) + ' | latency=' + escapeHTML(latency) + note + '</div>' +
          '<button type="button" class="proxy-select" data-address="' + escapeHTML(item.address) + '"' + (item.status !== 'healthy' ? ' disabled' : '') + '>Use this proxy</button>' +
        '</div>';
      }).join('');
      proxyResults.querySelectorAll('.proxy-select').forEach(button => {
        button.addEventListener('click', async () => {
          await selectProxy(button.getAttribute('data-address'));
        });
      });
    }

    async function createProxySessionIfNeeded(force) {
      if (!pageConfig.proxyEnabled) return loadProxyState();
      const state = loadProxyState();
      const rawList = proxyList ? proxyList.value.trim() : state.rawList;
      if (!rawList) throw new Error('Proxy list is empty');
      if (state.sessionId && !force) {
        setProxyState({ rawList: rawList });
        return loadProxyState();
      }
      const response = await fetch('/api/proxy/session', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          principal_id: commandForm.principal.value,
          proxies: rawList
        })
      });
      const payload = await response.json();
      if (!response.ok) throw new Error(payload.message || 'Failed to create proxy session');
      setProxyState({
        rawList: rawList,
        sessionId: payload.session.id,
        results: payload.session.candidates || [],
        selectedProxy: payload.session.selected_proxy || '',
        lease: payload.lease || null
      });
      return loadProxyState();
    }

    async function fetchProxySession() {
      if (!pageConfig.proxyEnabled) return;
      const state = loadProxyState();
      if (!state.sessionId) {
        renderProxySummary();
        return;
      }
      const response = await fetch('/api/proxy/session/' + encodeURIComponent(state.sessionId));
      if (response.status === 404) {
        saveProxyState({ rawList: state.rawList, sessionId: '', results: [], selectedProxy: '', lease: null });
        renderProxySummary();
        renderProxyResults();
        return;
      }
      const payload = await response.json();
      if (!response.ok) throw new Error(payload.message || 'Failed to load proxy session');
      setProxyState({
        results: payload.session.candidates || [],
        selectedProxy: payload.session.selected_proxy || '',
        lease: payload.lease || null
      });
    }

    async function checkProxySession() {
      const state = await createProxySessionIfNeeded(false);
      const response = await fetch('/api/proxy/session/' + encodeURIComponent(state.sessionId) + '/check', {
        method: 'POST'
      });
      const payload = await response.json();
      if (!response.ok) throw new Error(payload.message || 'Failed to check proxy session');
      setProxyState({
        results: payload.session.candidates || [],
        selectedProxy: payload.session.selected_proxy || '',
        lease: payload.lease || null
      });
    }

    async function selectProxy(proxyAddress) {
      const state = await createProxySessionIfNeeded(false);
      const response = await fetch('/api/proxy/session/' + encodeURIComponent(state.sessionId) + '/select', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ proxy: proxyAddress })
      });
      const payload = await response.json();
      if (!response.ok) throw new Error(payload.message || 'Failed to select proxy');
      setProxyState({
        results: payload.session.candidates || [],
        selectedProxy: payload.session.selected_proxy || '',
        lease: payload.lease || null
      });
    }

    commandForm.addEventListener('submit', async (event) => {
      event.preventDefault();
      const parsed = parseAskCommand(commandForm.command.value);
      if (parsed.name === 'ask') {
        const report = scanSensitiveInput(parsed.prompt);
        if (report.blocked) {
          const shouldSend = await confirmSensitiveSend(report);
          if (!shouldSend) return;
        }
        const proxyState = loadProxyState();
        if (pageConfig.proxyForced && !(proxyState.sessionId && proxyState.selectedProxy)) {
          output.textContent = JSON.stringify({
            status: 'error',
            message: 'proxy session is required by outbound policy'
          }, null, 2);
          return;
        }
      }
      const payload = {
        principal_id: commandForm.principal.value,
        command: commandForm.command.value,
        roles: getRoles(commandForm)
      };
      if (parsed.name === 'ask') {
        const proxyState = loadProxyState();
        if (proxyState.sessionId && proxyState.selectedProxy) {
          payload.args = { proxy_session_id: proxyState.sessionId };
        }
      }
      output.textContent = JSON.stringify(await sendCommand(payload), null, 2);
    });

    aiForm.addEventListener('submit', async (event) => {
      event.preventDefault();
      if (!pageConfig.aiEnabled) return;

      const prompt = aiForm.prompt.value.trim();
      if (!prompt) return;
      const report = scanSensitiveInput(prompt);
      if (report.blocked) {
        const shouldSend = await confirmSensitiveSend(report);
        if (!shouldSend) return;
      }

      const payload = {
        principal_id: commandForm.principal.value,
        roles: getRoles(commandForm),
        command: 'ask',
        args: {
          prompt: prompt
        }
      };
      const model = aiForm.model.value.trim();
      if (model) payload.args.model = model;
      const proxyState = loadProxyState();
      if (pageConfig.proxyForced && !(proxyState.sessionId && proxyState.selectedProxy)) {
        output.textContent = JSON.stringify({
          status: 'error',
          message: 'proxy session is required by outbound policy'
        }, null, 2);
        return;
      }
      if (proxyState.sessionId && proxyState.selectedProxy) {
        payload.args.proxy_session_id = proxyState.sessionId;
      }

      const response = await sendCommand(payload);
      output.textContent = JSON.stringify(response, null, 2);
      if (response?.status !== 'ok') {
        return;
      }

      const messages = loadMessages().slice();
      messages.push({ role: 'user', content: prompt, at: new Date().toISOString() });
      messages.push({ role: 'assistant', content: response?.data?.answer || 'Empty AI response', at: new Date().toISOString() });
      saveMessages(messages);
      renderMessages();
      aiForm.prompt.value = '';
    });

    exportButton.addEventListener('click', () => {
      const messages = loadMessages();
      const payload = {
        profile_id: pageConfig.profileId,
        exported_at: new Date().toISOString(),
        messages
      };
      const blob = new Blob([JSON.stringify(payload, null, 2)], { type: 'application/json' });
      const url = URL.createObjectURL(blob);
      const link = document.createElement('a');
      link.href = url;
      link.download = 'chat-export.json';
      link.click();
      URL.revokeObjectURL(url);
    });

    if (pageConfig.proxyEnabled) {
      syncProxyTextarea();
      renderProxySummary();
      renderProxyResults();

      openProxyButton.addEventListener('click', async () => {
        syncProxyTextarea();
        proxyDialog.showModal();
        try {
          await fetchProxySession();
        } catch (error) {
          output.textContent = JSON.stringify({ status: 'error', message: error.message }, null, 2);
        }
      });

      refreshProxyButton.addEventListener('click', async () => {
        try {
          await fetchProxySession();
        } catch (error) {
          output.textContent = JSON.stringify({ status: 'error', message: error.message }, null, 2);
        }
      });

      clearProxyButton.addEventListener('click', () => {
        saveProxyState({ rawList: '', sessionId: '', results: [], selectedProxy: '', lease: null });
        syncProxyTextarea();
        renderProxySummary();
        renderProxyResults();
      });

      createProxySessionButton.addEventListener('click', async () => {
        try {
          await createProxySessionIfNeeded(true);
        } catch (error) {
          output.textContent = JSON.stringify({ status: 'error', message: error.message }, null, 2);
        }
      });

      checkProxySessionButton.addEventListener('click', async () => {
        try {
          await checkProxySession();
        } catch (error) {
          output.textContent = JSON.stringify({ status: 'error', message: error.message }, null, 2);
        }
      });

      selectFastestProxyButton.addEventListener('click', async () => {
        try {
          await selectProxy('fastest');
        } catch (error) {
          output.textContent = JSON.stringify({ status: 'error', message: error.message }, null, 2);
        }
      });

      closeProxyButton.addEventListener('click', () => proxyDialog.close());
    }

    guardEditButton.addEventListener('click', () => {
      guardDialog.close();
      if (guardDecisionResolver) guardDecisionResolver(false);
      guardDecisionResolver = null;
    });

    guardSendButton.addEventListener('click', () => {
      guardDialog.close();
      if (guardDecisionResolver) guardDecisionResolver(true);
      guardDecisionResolver = null;
    });

    renderMessages();
  </script>
</body>
</html>`)
	if page == nil {
		http.Error(writer, "template error", http.StatusInternalServerError)
		return
	}
	data := pageData{
		ProfileID:    a.service.Profile().ID,
		AIEnabled:    a.service.Profile().Has(buildprofile.CapabilityAIChat),
		ChatHistory:  a.service.Profile().OpenAI.ChatHistory,
		ProxyEnabled: a.service.Profile().ProxySession.Enabled && a.proxy != nil,
		ProxyForced:  a.service.Profile().ProxySession.Force,
		DefaultModel: a.service.Profile().OpenAI.Model,
	}
	if err := page.Execute(writer, data); err != nil {
		http.Error(writer, err.Error(), http.StatusInternalServerError)
	}
}

type commandRequest struct {
	PrincipalID string            `json:"principal_id"`
	Roles       []string          `json:"roles"`
	Command     string            `json:"command"`
	Args        map[string]string `json:"args,omitempty"`
}

type proxyCreateRequest struct {
	PrincipalID string `json:"principal_id"`
	Proxies     string `json:"proxies"`
}

type proxySelectRequest struct {
	Proxy string `json:"proxy"`
}

func (a *Adapter) handleCommand(writer http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(writer, request.Body, maxRequestBodyBytes)
	payload, ok := decodeCommandRequest(writer, request)
	if !ok {
		return
	}

	result, err := a.service.Execute(request.Context(), a.commandEnvelope(*payload))
	if err != nil && result.Message == "" {
		result.Message = err.Error()
	}

	statusCode := http.StatusOK
	if result.Status == "error" {
		statusCode = http.StatusBadRequest
	}
	writeJSON(writer, statusCode, result)
}

func (a *Adapter) handleCommandStream(writer http.ResponseWriter, request *http.Request) {
	request.Body = http.MaxBytesReader(writer, request.Body, maxRequestBodyBytes)
	payload, ok := decodeCommandRequest(writer, request)
	if !ok {
		return
	}

	writer.Header().Set("Content-Type", "text/event-stream")
	writer.Header().Set("Cache-Control", "no-cache")
	writer.Header().Set("Connection", "keep-alive")

	streamWriter, err := newSSEStreamWriter(request.Context(), writer)
	if err != nil {
		writeJSON(writer, http.StatusInternalServerError, map[string]any{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	if err := a.service.ExecuteStream(request.Context(), a.commandEnvelope(*payload), streamWriter); err != nil {
		log.Printf("command stream failed: %v", err)
	}
}

func decodeCommandRequest(writer http.ResponseWriter, request *http.Request) (*commandRequest, bool) {
	var payload commandRequest
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]any{
			"status":  "error",
			"message": "invalid JSON payload",
		})
		return nil, false
	}
	return &payload, true
}

func (a *Adapter) commandEnvelope(payload commandRequest) core.CommandEnvelope {
	name, args := parseCommand(payload.Command)
	if len(payload.Args) > 0 {
		if args == nil {
			args = make(map[string]string, len(payload.Args))
		}
		for key, value := range payload.Args {
			args[key] = value
		}
	}
	return core.CommandEnvelope{
		ID:        fmt.Sprintf("http-%d", time.Now().UTC().UnixNano()),
		Transport: "http_chat",
		Name:      name,
		Args:      args,
		RawText:   payload.Command,
		Principal: core.Principal{
			ID:        payload.PrincipalID,
			Display:   payload.PrincipalID,
			Transport: "http_chat",
			Roles:     payload.Roles,
		},
		RequestedAt: time.Now().UTC(),
	}
}

func (a *Adapter) handleProxyCollection(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		http.NotFound(writer, request)
		return
	}
	if !a.proxyEnabled() {
		http.NotFound(writer, request)
		return
	}

	request.Body = http.MaxBytesReader(writer, request.Body, maxRequestBodyBytes)
	var payload proxyCreateRequest
	if err := json.NewDecoder(request.Body).Decode(&payload); err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]any{
			"status":  "error",
			"message": "invalid proxy session payload",
		})
		return
	}

	session, err := a.proxy.CreateSession(payload.PrincipalID, []string{payload.Proxies})
	if err != nil {
		writeJSON(writer, http.StatusBadRequest, map[string]any{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	writeJSON(writer, http.StatusCreated, map[string]any{
		"status":  "ok",
		"message": "proxy session created",
		"session": session,
		"lease":   proxyLeaseFromSession(session),
	})
}

func (a *Adapter) handleProxySession(writer http.ResponseWriter, request *http.Request) {
	if !a.proxyEnabled() {
		http.NotFound(writer, request)
		return
	}
	request.Body = http.MaxBytesReader(writer, request.Body, maxRequestBodyBytes)

	sessionID, action, ok := parseProxyPath(request.URL.Path)
	if !ok {
		http.NotFound(writer, request)
		return
	}

	var (
		session proxymgr.Session
		err     error
	)

	switch {
	case request.Method == http.MethodGet && action == "":
		session, err = a.proxy.GetSession(sessionID)
	case request.Method == http.MethodPost && action == "check":
		session, err = a.proxy.CheckSession(request.Context(), sessionID)
	case request.Method == http.MethodPost && action == "select":
		var payload proxySelectRequest
		if decodeErr := json.NewDecoder(request.Body).Decode(&payload); decodeErr != nil {
			writeJSON(writer, http.StatusBadRequest, map[string]any{
				"status":  "error",
				"message": "invalid proxy selection payload",
			})
			return
		}
		session, err = a.proxy.SelectProxy(sessionID, payload.Proxy)
	default:
		http.NotFound(writer, request)
		return
	}

	if err != nil {
		statusCode := http.StatusBadRequest
		if errors.Is(err, proxymgr.ErrSessionNotFound) {
			statusCode = http.StatusNotFound
		}
		writeJSON(writer, statusCode, map[string]any{
			"status":  "error",
			"message": err.Error(),
		})
		return
	}

	writeJSON(writer, http.StatusOK, map[string]any{
		"status":  "ok",
		"message": "proxy session loaded",
		"session": session,
		"lease":   proxyLeaseFromSession(session),
	})
}

func parseCommand(raw string) (string, map[string]string) {
	trimmed := strings.TrimSpace(raw)
	fields := strings.Fields(trimmed)
	if len(fields) == 0 {
		return "", nil
	}

	args := make(map[string]string)
	if strings.EqualFold(fields[0], "ask") && len(fields) > 1 && !strings.Contains(fields[1], "=") {
		args["prompt"] = strings.TrimSpace(trimmed[len(fields[0]):])
		return strings.ToLower(fields[0]), args
	}
	for _, field := range fields[1:] {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) == 2 {
			args[parts[0]] = parts[1]
		}
	}
	return strings.ToLower(fields[0]), args
}

func writeJSON(writer http.ResponseWriter, status int, payload any) {
	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(status)
	_ = json.NewEncoder(writer).Encode(payload)
}

func (a *Adapter) getIndexTemplate(raw string) *template.Template {
	a.tmplOnce.Do(func() {
		a.indexTmpl, a.indexTmplErr = template.New("index").Parse(raw)
		if a.indexTmplErr != nil {
			log.Printf("failed to parse index template: %v", a.indexTmplErr)
		}
	})
	return a.indexTmpl
}

func (a *Adapter) proxyEnabled() bool {
	return a.proxy != nil && a.service.Profile().ProxySession.Enabled
}

func parseProxyPath(path string) (string, string, bool) {
	trimmed := strings.TrimPrefix(path, "/api/proxy/session/")
	if trimmed == path || trimmed == "" {
		return "", "", false
	}
	parts := strings.Split(strings.Trim(trimmed, "/"), "/")
	if len(parts) == 1 {
		return parts[0], "", true
	}
	if len(parts) == 2 {
		return parts[0], parts[1], true
	}
	return "", "", false
}

func proxyLeaseFromSession(session proxymgr.Session) *proxymgr.Lease {
	if session.SelectedProxy == "" {
		return nil
	}
	for _, candidate := range session.Candidates {
		if candidate.Address != session.SelectedProxy || candidate.Status != proxymgr.StatusHealthy {
			continue
		}
		return &proxymgr.Lease{
			SessionID: session.ID,
			Address:   candidate.Address,
			LatencyMS: candidate.LatencyMS,
			UpdatedAt: session.UpdatedAt,
		}
	}
	return nil
}
