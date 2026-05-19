# Architecture Decision Records

This directory contains the long-form rationale for architectural decisions in qctx. Each ADR is immutable once accepted — superseding decisions get new ADRs that reference the prior.

| # | Title | Status |
|---|---|---|
| [0001](0001-jsonl-schema-for-ai-agents.md) | Type-discriminated JSONL schema for AI agent consumption | Accepted |
| [0002](0002-enterprise-first-networking.md) | Enterprise-first networking — no public-cloud defaults | Accepted |
| [0003](0003-partial-success-policy.md) | Partial-success policy and `--strict` opt-in | Accepted |
| [0004](0004-secret-type-for-token-redaction.md) | `Secret` type for token redaction | Accepted |
| [0005](0005-cobra-cli-framework.md) | Cobra as the CLI framework | Accepted |
| [0006](0006-jsonl-byte-stitching.md) | JSONL writer uses raw byte-stitching for the type discriminator | Accepted |
| [0007](0007-retryablehttp.md) | `hashicorp/go-retryablehttp` for HTTP retries | Accepted |

## Why ADRs

Design context decays fastest. A reviewer six months out cannot tell the difference between "this is intentional and load-bearing" and "this was an accident we never cleaned up". ADRs capture the *why* so the *what* stays maintainable.

## When to write one

- A choice will affect more than one package.
- A choice constrains future contributors (e.g., schema stability, security policy).
- A choice was non-obvious and the alternatives are worth recording.

## Template

```markdown
# NNNN — Title

**Status:** Proposed | Accepted | Superseded by [NNNN](NNNN-link.md)
**Date:** YYYY-MM-DD

## Context

What problem are we solving? What constraints apply?

## Decision

The specific choice. One paragraph.

## Consequences

What follows from this — both upside and downside. Be honest about the costs.

## Alternatives considered

What else did we look at, and why did we not pick it?
```
