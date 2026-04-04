# Go SDK

`sdk/hubrelay` is the primary contract for clients talking to HubRelay.

## Transports

- HTTP
- unix socket

## Client setup

```go
import "sshbot/sdk/hubrelay"

client := hubrelay.NewHTTPClient("http://127.0.0.1:5500",
    hubrelay.WithPrincipal(hubrelay.Principal{
        ID:    "operator-local",
        Roles: []string{"operator"},
    }),
)
defer client.Close()
```

## Commands

```go
result, err := client.Health(ctx)
result, err := client.Execute(ctx, hubrelay.CommandRequest{
    Command: "ask",
    Args: map[string]string{"prompt": "hello"},
})
```

## Streaming

```go
stream, err := client.ExecuteStream(ctx, hubrelay.CommandRequest{
    Command: "ask",
    Args: map[string]string{"prompt": "hello"},
})
```

## Egress status

`client.EgressStatus(ctx, hubrelay.Principal{ID: "operator-local", Roles: []string{"operator"}})`

Response includes gateway name, interface, health level, active flag, and probe timestamps.
