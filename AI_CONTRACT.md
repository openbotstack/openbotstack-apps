# AI RULES — READ FIRST

This file defines mandatory rules for AI coding tools.
All instructions in this file are authoritative for this repository.

# REPOSITORY: openbotstack-apps

## ROLE:
This repository implements the APPLICATION PLANE of OpenBotStack.

## IT MAY CONTAIN:
- Domain-specific skills (nursing, clinical, operations)
- Tool adapters (EHR, analytics, database wrappers)
- Workflows (compositions of skills and tools)
- Example applications and CLI demos
- Stub data for development and testing

## IT MUST NOT:
- Define assistant identity or persona
- Make policy or permission decisions
- Contain tenant or user configuration logic
- Implement skill execution engines
- Define long-term memory models
- Implement HTTP servers or API endpoints
- Contain control-plane or runtime logic

## DESIGN RULES:
- Skills implement the `registry/skills.Skill` interface from openbotstack-core
- Tools are self-contained, stateless adapters
- Workflows build `execution.ExecutionPlan` from openbotstack-core
- All domain logic is deterministic and testable
- No direct LLM provider calls; use core model abstractions

> This repo MUST NOT contain any executable entrypoint beyond example demos.
