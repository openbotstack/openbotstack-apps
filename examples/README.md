# OpenBotStack Skill Examples

Each example demonstrates one skill execution type.

| Skill | Mode | Description |
|-------|------|-------------|
| hello-world | wasm | Simplest Wasm skill — accepts input, returns greeting |

See `openbotstack-runtime/skills/` for system-default skills (summarize, etc.).

To use these examples:
```bash
# Copy to runtime skills directory
cp -r examples/hello-world ../openbotstack-runtime/skills/

# Or set OBS_SKILLS_PATH
OBS_SKILLS_PATH=./examples go run ./cmd/openbotstack
```
