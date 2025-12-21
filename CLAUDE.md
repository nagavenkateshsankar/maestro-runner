# Claude Instructions for maestro-runner

## Code Quality Rules

1. **KISS and DRY** - Check for simplicity and no repetition after each code change
2. **Unit tests** - Write tests only after code is reviewed and approved
3. **No Claude credits** - Do not add Claude/AI credits in files or commits

## Code Review Checklist (Run after every code change)

```
1. KISS        - Is the code simple? No over-engineering?
2. DRY         - No duplicated code/logic across files?
3. Quality     - Clean, readable, follows Go conventions?
4. Tests       - Are there tests for the changed code?
5. Unit Tests  - go test ./...
6. Build       - go build .
7. Coverage    - go test ./... -cover (report per module)
```

## Project Context

- **Architecture**: 3-part design (YAML → Executor → Report)
- **Reference code**: `/Users/omnarayan/work/go/src/maestro-appium/`
- **Maestro research**: `/Users/omnarayan/work/support-tools/Maestro/docs/`

## Design Principles

- Executor-agnostic design (Appium, Native, Detox are equal implementations)
- Configurable timeouts at flow, step, command level
- Small, focused components (avoid god classes)
- Independent parts (changes in one don't affect others)
