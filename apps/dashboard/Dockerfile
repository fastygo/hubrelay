FROM node:22-alpine AS css
WORKDIR /src/app
COPY app/package.json ./
COPY app/static ./static
COPY app/views ./views
COPY app/ui8kit ./ui8kit
COPY app/internal ./internal
COPY app/cmd ./cmd
RUN npm install
RUN npm run build:css

FROM golang:1.25.5-alpine AS build
WORKDIR /src
RUN apk add --no-cache git
COPY go.mod go.sum ./
COPY sdk ./sdk
COPY app ./app
WORKDIR /src/app
RUN git clone https://github.com/fastygo/ui8kit.git /deps/ui8kit
RUN go mod edit -replace github.com/fastygo/ui8kit=/deps/ui8kit
RUN go mod download
COPY --from=css /src/app/static/css/app.css /src/app/static/css/app.css
RUN go install github.com/a-h/templ/cmd/templ@v0.3.1001
RUN templ generate ./...
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o /out/dashboard ./cmd/server

FROM scratch
WORKDIR /app
COPY --from=build /out/dashboard /app/dashboard
COPY --from=build /src/app/static /app/static
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
ENTRYPOINT ["/app/dashboard"]
