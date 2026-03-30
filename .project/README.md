# Project design (`.project`)

Short-form architecture notes, ADRs, and profile sketches. They stay **technology-agnostic** where possible and avoid real secrets or operator-specific hosts.

## Index

| Document | Topic |
| --- | --- |
| [architecture.md](architecture.md) | Command flow, adapters, outbound policy |
| [core-contracts.md](core-contracts.md) | Envelope, principal, result shapes |
| [deploy-model.md](deploy-model.md) | Immutable image vs mutable state |
| [outbound-model.md](outbound-model.md) | Why shared egress policy exists |
| [security-model.md](security-model.md) | Capability and audit boundaries |
| [storage-model.md](storage-model.md) | `bbolt` responsibilities |
| [threat-model.md](threat-model.md) | Trust assumptions |
| [roadmap.md](roadmap.md) | Phases and invariants |
| [profiles/tunnel-email-openai.md](profiles/tunnel-email-openai.md) | Default deploy profile |
| [adr/](adr/) | Architecture decision records |

## WireGuard + HubRelay (deep dive)

| Document | Topic |
| --- | --- |
| [wg/HUBRELAY_SYSTEM_GUIDE.md](wg/HUBRELAY_SYSTEM_GUIDE.md) | Binary, `INPUT_*` profile, private egress checks, kernel routing |
| [wg/README.md](wg/README.md) | SRE runbook: Marvin/Jarvis, checks, failures |

## Canonical operator docs

End-to-end English guides (installation through deploy) live in [`../docs`](../docs/README.md).
