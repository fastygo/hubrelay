# Overview: why this bot exists

## The problem

Operators often need a **stable control surface** on a server that:

- survives **container replacement** and image updates,
- works when **browser egress** is restricted but the **server** can still reach allowed endpoints or a proxy pool,
- keeps **secrets out** of the primary runtime database,
- exposes a **single command model** over HTTP today and email (or others) tomorrow.

## What this repository provides

A **Go** service that:

1. Compiles an **immutable profile** (capabilities, bind address, AI settings) into the binary or image.
2. Persists **mutable** state (principals, sessions, audit) in **BoltDB** on a mounted volume, not in the container root filesystem.
3. Accepts work through **adapters** (e.g. loopback HTTP chat) that map everything into a **command bus** and plugins.
4. Optionally routes **outbound AI traffic** through a **SOCKS proxy lease** when policy requires it.
5. Offers a minimal **browser UI** intended for access via **SSH local port forward** to loopback.

## What it is not

- Not a multi-tenant SaaS with dynamic plugin installs at runtime.
- Not a generic shell gateway: the default profile avoids arbitrary command execution.
- Not a place to store API keys in BoltDB: keys belong in the deploy-time profile only.

## Why loopback + SSH tunnel?

**Why**: Binding HTTP to `127.0.0.1` avoids accidentally publishing an operator panel to the public internet. **SSH `-L`** reuses existing authenticated access to the host and works across many restricted networks.

## Why BoltDB outside the image?

**Why**: Containers may be read-only or recreated. State that must survive (sessions, audit) needs a **host-mounted path**. BoltDB is a single-file store that fits that model without running a separate database service.

## Why a separate outbound policy layer?

**Why**: The first consumer is the AI provider, but the same rules will apply to future HTTP integrations. Centralising **direct vs proxy vs blocked** prevents each plugin from inventing its own network rules. See [Outbound model](../../.project/outbound-model.md).

## Where to go next

- [Getting started](../getting-started/README.md) for toolchain and first run.
- [Network, tunnel, and proxy](../network-tunnel-and-proxy/README.md) for the operator path you validated (tunnel + SOCKS).
