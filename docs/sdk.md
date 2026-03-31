# Go SDK

The primary client contract for HubRelay is `sdk/hubrelay`.

It lets another Go application talk to the HubRelay binary without embedding transport details into business logic.

Supported transports:
- HTTP
- unix socket

## Install

Inside an application in the same module:

```go
import "sshbot/sdk/hubrelay"
```

## HTTP client

Use HTTP when HubRelay is reached over loopback, private TCP, or an SSH tunnel:

```go
client := hubrelay.NewHTTPClient("http://127.0.0.1:5500",
    hubrelay.WithPrincipal(hubrelay.Principal{
        ID:    "operator-local",
        Roles: []string{"operator"},
    }),
)
defer client.Close()

health, err := client.Health(ctx)
```

## Unix socket client

Use unix sockets when the client runs on the same host as HubRelay:

```go
client := hubrelay.NewUnixClient("/run/hubrelay/hubrelay.sock",
    hubrelay.WithPrincipal(hubrelay.Principal{
        ID:    "dashboard",
        Roles: []string{"operator"},
    }),
)
defer client.Close()

result, err := client.Execute(ctx, hubrelay.CommandRequest{
    Command: "capabilities",
})
```

## Execute a command

```go
result, err := client.Execute(ctx, hubrelay.CommandRequest{
    Command: "ask",
    Args: map[string]string{
        "prompt": "hello",
    },
})
```

## Stream a command

```go
stream, err := client.ExecuteStream(ctx, hubrelay.CommandRequest{
    Command: "ask",
    Args: map[string]string{
        "prompt": "hello",
    },
})
if err != nil {
    return err
}
defer stream.Close()

for stream.Next() {
    chunk := stream.Chunk()
    fmt.Print(chunk.Delta)
}

result, err := stream.Result()
```

## Egress status

```go
status, err := client.EgressStatus(ctx, hubrelay.Principal{
    ID:    "operator-local",
    Roles: []string{"operator"},
})
```

The response includes:
- gateway name,
- interface name,
- health level,
- active flag,
- last transition time,
- per-level WG, transport, and business probe status.
