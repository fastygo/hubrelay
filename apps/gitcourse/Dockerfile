FROM node:22-alpine AS css
WORKDIR /src/apps/gitcourse
COPY apps/gitcourse/package.json ./
COPY apps/gitcourse/static ./static
COPY apps/gitcourse/views ./views
COPY apps/gitcourse/internal ./internal
COPY apps/gitcourse/cmd ./cmd
RUN npm install
RUN npm run sync:ui8kit
RUN npm run build:css

FROM golang:1.25.5-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
COPY hubcore ./hubcore
COPY sdk ./sdk
COPY apps/dashboard/ui8kit ./apps/dashboard/ui8kit
COPY apps/gitcourse ./apps/gitcourse
WORKDIR /src/apps/gitcourse
RUN go mod download
COPY --from=css /src/apps/gitcourse/static/css/app.css /src/apps/gitcourse/static/css/app.css
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1001
RUN templ generate ./...
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /out/gitcourse ./cmd/server

FROM scratch
WORKDIR /app
COPY --from=build /out/gitcourse /app/gitcourse
COPY --from=build /src/apps/gitcourse/static /app/static
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/app/gitcourse"]
