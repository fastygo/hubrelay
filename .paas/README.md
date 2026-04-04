# HubRelay PAAS Deployment (Bot + Dashboard)

Этот `.paas`-пакет сейчас поддерживает **2 основных сценария**:

- `deploy-hostrun` — деплой **бота** (runtime `/opt/hubrelay`, service `hubrelay.service`)
- `deploy-app` — деплой **админки** (dashboard, runtime `/opt/hubrelay-dashboard`, service `hubrelay-dashboard.service`)

> Важно: обе команды используют `parse_config.sh`, поэтому файл
> `.paas/parse_config.sh` должен быть в репозитории.

---

## Текущий набор ключевых файлов `.paas`

- `config.yml` (дефолтные `INPUT_*`)
- `extensions/deploy-hostrun.yml`
- `extensions/deploy-app.yml`
- `deploy-hostrun-clean.sh`
- `deploy-app-clean.sh`
- `README.md` (этот файл)

У вас должен быть:

- `.paas/parse_config.sh` — **обязательный** (без него `deploy-*.sh` не выполняется)
- пары ключей SSH в `~/.ssh` + пароль/выписки доступа к серверу

---

## Сценарий 1: деплой только бота (`hubrelay.service`)

Используйте, если нужен **только bot API**:

```bash
export HUBRELAY_HOST='176.124.209.3'
export HUBRELAY_USER='root'
export HUBRELAY_SSH_KEY='C:/Users/alexe/.ssh/appserv'

export INPUT_AI_API_KEY='<OPENAI_API_KEY>'
export INPUT_AI_BASE_URL='https://api.cerebras.ai/v1'
export INPUT_AI_MODEL='gpt-oss-120b'

bash ./.paas/deploy-hostrun-clean.sh
```

Что делает скрипт:

- рендерит env для расширения `deploy-hostrun`
- билдит и деплоит только `./cmd/bot` как `/opt/hubrelay/bot`
- пересобирает `systemd` юнит `hubrelay.service`
- перезапускает сервис и делает health/capabilities smoke

Проверка после деплоя:

```bash
ssh -i "$HUBRELAY_SSH_KEY" "${HUBRELAY_USER}@${HUBRELAY_HOST}"
systemctl status hubrelay.service --no-pager
curl http://127.0.0.1:5500/healthz
curl -X POST http://127.0.0.1:5500/api/command \
  -H "Content-Type: application/json" \
  -d '{"principal_id":"operator-local","roles":["operator"],"command":"capabilities"}'
```

SSH доступ к API с вашей машины:

```bash
ssh -N -L 5500:127.0.0.1:5500 -i "$HUBRELAY_SSH_KEY" "${HUBRELAY_USER}@${HUBRELAY_HOST}"
# в браузере: http://127.0.0.1:5500
```

---

## Сценарий 2: деплой админки (`hubrelay-dashboard.service`)

Нужно, чтобы работала веб-админка. Требует, чтобы бот уже был поднят и доступен на `127.0.0.1:5500`.

```bash
export HUBRELAY_HOST='176.124.209.3'
export HUBRELAY_USER='root'
export HUBRELAY_SSH_KEY='C:/Users/alexe/.ssh/appserv'

# переменные для dashboard-сценария
export APP_HOST="$HUBRELAY_HOST"
export APP_USER="$HUBRELAY_USER"
export APP_SSH_KEY="$HUBRELAY_SSH_KEY"

export INPUT_APP_ADMIN_PASS='<CHANGE_ME>'

bash ./.paas/deploy-app-clean.sh
```

Что делает скрипт:

- рендерит env для расширения `deploy-app`
- собирает `apps/dashboard/cmd/server` локально (`dist/hubrelay-dashboard`)
- загружает бинарник и static-ассеты на `SERVER_HOST`
- рендерит и перезапускает `hubrelay-dashboard.service`
- делает smoke-проверки `/login`, `/`, статических файлов и auth-эндпоинтов

Проверка после деплоя:

```bash
ssh -i "$APP_SSH_KEY" "${APP_USER}@${APP_HOST}"
systemctl status hubrelay-dashboard.service --no-pager
curl -I http://127.0.0.1:8080/login
curl -sS -u "${INPUT_APP_ADMIN_USER:-admin}:${INPUT_APP_ADMIN_PASS}" \
  http://127.0.0.1:8080/capabilities
```

SSH-доступ в админку:

```bash
ssh -N -L 18080:127.0.0.1:8080 -i "$APP_SSH_KEY" "${APP_USER}@${APP_HOST}"
# в браузере: http://127.0.0.1:18080/login
```

---

## Что важно помнить

- `deploy-hostrun` и `deploy-app` — независимые процессы и разные артефакты.
- Убеждайтесь, что пароль `INPUT_APP_ADMIN_PASS` всегда задаётся через env (не хранить в git).
- Если у вас есть `WG`, можно оставить расширение `deploy-hostrun` с WG-настройками; если WG не нужен, оставьте `INPUT_BOT_APP_WG_ENABLED=false` (по умолчанию в скрипте).
