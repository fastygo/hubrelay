FROM golang:1.25.5-alpine AS build
WORKDIR /src

ARG BOT_PROFILE_ID=tunnel-email-openai
ARG BOT_DISPLAY_NAME="Tunnel chat + Yandex mail + OpenAI"
ARG BOT_HTTP_BIND=0.0.0.0:5500
ARG BOT_EMAIL_ENABLED=true
ARG BOT_EMAIL_PROVIDER=yandex
ARG BOT_EMAIL_MODE=scaffold
ARG BOT_OPENAI_ENABLED=true
ARG AI_PROVIDER=openai
ARG AI_API_KEY=
ARG AI_BASE_URL=
ARG AI_MODEL=gpt-4.1-mini
ARG AI_API_MODE=chat_completions
ARG CHAT_HISTORY=false
ARG PROXY_SESSION_ENABLED=true
ARG PROXY_SESSION_FORCE=true
ARG PRIVATE_EGRESS_REQUIRED=false
ARG PRIVATE_EGRESS_INTERFACE=
ARG PRIVATE_EGRESS_TEST_HOST=
ARG PRIVATE_EGRESS_FAIL_CLOSED=false

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -ldflags "-s -w \
        -X 'sshbot/internal/buildprofile.currentProfileID=${BOT_PROFILE_ID}' \
        -X 'sshbot/internal/buildprofile.currentDisplayName=${BOT_DISPLAY_NAME}' \
        -X 'sshbot/internal/buildprofile.currentHTTPBind=${BOT_HTTP_BIND}' \
        -X 'sshbot/internal/buildprofile.currentEmailEnabled=${BOT_EMAIL_ENABLED}' \
        -X 'sshbot/internal/buildprofile.currentEmailProvider=${BOT_EMAIL_PROVIDER}' \
        -X 'sshbot/internal/buildprofile.currentEmailMode=${BOT_EMAIL_MODE}' \
        -X 'sshbot/internal/buildprofile.currentOpenAIEnabled=${BOT_OPENAI_ENABLED}' \
        -X 'sshbot/internal/buildprofile.currentAIProvider=${AI_PROVIDER}' \
        -X 'sshbot/internal/buildprofile.currentAIAPIKey=${AI_API_KEY}' \
        -X 'sshbot/internal/buildprofile.currentAIBaseURL=${AI_BASE_URL}' \
        -X 'sshbot/internal/buildprofile.currentAIModel=${AI_MODEL}' \
        -X 'sshbot/internal/buildprofile.currentAIAPIMode=${AI_API_MODE}' \
        -X 'sshbot/internal/buildprofile.currentChatHistory=${CHAT_HISTORY}' \
        -X 'sshbot/internal/buildprofile.currentProxySession=${PROXY_SESSION_ENABLED}' \
        -X 'sshbot/internal/buildprofile.currentProxyForce=${PROXY_SESSION_FORCE}' \
        -X 'sshbot/internal/buildprofile.currentPrivateEgressRequired=${PRIVATE_EGRESS_REQUIRED}' \
        -X 'sshbot/internal/buildprofile.currentPrivateEgressInterface=${PRIVATE_EGRESS_INTERFACE}' \
        -X 'sshbot/internal/buildprofile.currentPrivateEgressTestHost=${PRIVATE_EGRESS_TEST_HOST}' \
        -X 'sshbot/internal/buildprofile.currentPrivateEgressFailClosed=${PRIVATE_EGRESS_FAIL_CLOSED}'" \
      -o /out/bot ./cmd/bot

FROM scratch
WORKDIR /app
COPY --from=build /out/bot /app/bot
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

ENTRYPOINT ["/app/bot"]
