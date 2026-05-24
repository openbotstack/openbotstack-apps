# AI RULES — READ FIRST

This file defines mandatory rules for AI coding tools.
All instructions in this file are authoritative for this repository.

# REPOSITORY: openbotstack-apps

## ROLE:
This repository implements the APPLICATION PLANE of OpenBotStack.

## IT MAY CONTAIN:
- Domain-specific manifest-based skills (SKILL.md + manifest.yaml)
- MCP server implementations for external system integration
- Workflow definitions composing skills into execution plans
- Industry-specific audit event mappers (implementing core AuditEventMapper)
- Example applications with E2E test suites

## IT MUST NOT:
- Define assistant identity or persona
- Make policy or permission decisions
- Contain tenant or user configuration logic
- Implement skill execution engines
- Define long-term memory models
- Implement HTTP servers or API endpoints
- Contain control-plane or runtime logic

## DESIGN RULES:
- Skills use manifest-based definitions (manifest.yaml + SKILL.md), loaded by runtime
- MCP servers implement the MCP protocol for external system integration
- Workflows build `execution.ExecutionPlan` from openbotstack-core
- Audit mappers implement `audit.AuditEventMapper` from openbotstack-core
- All domain logic is deterministic and testable
- No direct LLM provider calls; use core model abstractions

> This repo MUST NOT contain any executable entrypoint beyond example MCP servers.
