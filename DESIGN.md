# maestro-runner - Design Document

## Overview

**maestro-runner** is a universal test runner that executes Maestro YAML flow files on multiple backends (executors). It's a clean-slate redesign of `maestro-appium` with a pluggable architecture.

## Core Architecture

```
┌──────────────┐       ┌──────────────┐       ┌──────────────┐
│     YAML     │──────▶│   EXECUTOR   │──────▶│    REPORT    │
│    (fixed)   │       │  (contract)  │       │    (fixed)   │
└──────────────┘       └──────┬───────┘       └──────────────┘
                              │
          ┌───────────────────┼───────────────────┐
          │         │         │         │         │
          ▼         ▼         ▼         ▼         ▼
     ┌────────┐┌────────┐┌────────┐┌────────┐┌────────┐
     │ appium ││ native ││ detox  ││playwright││selenium│
     └────────┘└────────┘└────────┘└────────┘└────────┘

     All equal. All just implement the Executor interface.
```

## Three Parts

### 1. YAML (Fixed - Input)
- Parses Maestro YAML flow files
- Validates structure
- Produces universal steps
- **Changes here do NOT affect executors**

### 2. Executor (Contract - Pluggable)
- Interface that all backends implement
- Converts steps to backend-specific calls
- Returns results
- **New executor = just implement the interface**

### 3. Report (Fixed - Output)
- Consumes execution results
- Generates reports (JSON, JUnit, HTML, Allure)
- **Changes here do NOT affect executors**

## Key Design Decisions

### Naming: "Executor" (not "driver", "runner", "mode")
- Follows GitLab CI pattern (runner + executor)
- "Runner" = the CLI application (`maestro-runner`)
- "Executor" = HOW it runs (appium, native, detox)
- Avoids confusion with Appium's internal "driver" terminology

### CLI Usage
```bash
maestro-runner run login.yaml --executor appium --server localhost:4723
maestro-runner run login.yaml --executor native
maestro-runner run login.yaml --executor detox
maestro-runner run login.yaml --executor playwright --browser chrome
```

## Executors

### appium (FIRST - already have working code)
- Via Appium server
- Works with BrowserStack, SauceLabs, AWS Device Farm, DeviceLab
- Reuse code from `maestro-appium` project

### native (LATER)
- Direct UiAutomator2 (Android) / WDA (iOS)
- NO Maestro dependency
- NO Appium dependency
- Fixes Maestro limitations:
  - Port 7001 hardcoded → dynamic ports (parallel execution)
  - No real iOS → real iOS via WDA
  - Closed ecosystem → open, pluggable

### detox (FUTURE)
- Via Detox server
- Fast, parallel, grey-box testing

### playwright (FUTURE)
- Via Playwright
- Web/PWA testing

### selenium (FUTURE)
- Via Selenium Grid
- Web testing

## Package Structure

```
maestro-runner/
├── cmd/
│   └── maestro-runner/        # CLI entry point
│       └── main.go
│
├── pkg/
│   ├── flow/                  # YAML parsing (FIXED)
│   │   ├── parser.go          # Parse YAML files
│   │   ├── flow.go            # Flow model
│   │   ├── step.go            # Step model (universal)
│   │   └── validator.go       # Validation
│   │
│   ├── executor/              # Interface contract
│   │   └── executor.go        # type Executor interface {...}
│   │
│   ├── executors/             # Implementations (add as needed)
│   │   └── appium/            # First implementation
│   │       ├── executor.go    # implements Executor
│   │       ├── session.go     # WebDriver session
│   │       ├── resolver.go    # Selector → locator
│   │       └── protocol/      # W3C WebDriver protocol
│   │
│   └── report/                # Reporting (FIXED)
│       ├── reporter.go        # Report generation
│       ├── formats/           # JSON, JUnit, HTML, Allure
│       └── events.go          # Execution events
│
├── internal/
│   └── config/                # Internal configuration
│
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Executor Interface (Draft)

```go
package executor

// Executor is the contract all backends must implement
type Executor interface {
    // Lifecycle
    Connect(config Config) (Session, error)
    Disconnect() error

    // Capabilities - what this executor supports
    Capabilities() Capabilities
}

// Session represents an active connection to device/browser
type Session interface {
    // Element operations
    FindElement(selector Selector) (Element, error)
    FindElements(selector Selector) ([]Element, error)

    // Actions
    Tap(element Element) error
    DoubleTap(element Element) error
    LongPress(element Element, duration time.Duration) error
    Input(element Element, text string) error
    Clear(element Element) error
    Swipe(from, to Point, duration time.Duration) error

    // App lifecycle
    LaunchApp(appID string) error
    TerminateApp(appID string) error
    ActivateApp(appID string) error

    // Queries
    GetText(element Element) (string, error)
    GetAttribute(element Element, attr string) (string, error)
    IsVisible(element Element) (bool, error)
    IsEnabled(element Element) (bool, error)

    // Device
    GetPageSource() (string, error)
    TakeScreenshot() ([]byte, error)
    SetLocation(lat, lon float64) error

    // Platform
    Platform() Platform // android, ios, web
}

// Capabilities defines what an executor supports
type Capabilities struct {
    SupportsLocation    bool
    SupportsBiometrics  bool
    SupportsDeepLinks   bool
    SupportsParallel    bool
    SupportedPlatforms  []Platform
}

// Platform enum
type Platform string
const (
    PlatformAndroid Platform = "android"
    PlatformIOS     Platform = "ios"
    PlatformWeb     Platform = "web"
)
```

## Impact Matrix

| Change | YAML | Executor | Report |
|--------|------|----------|--------|
| Add new executor (e.g., detox) | No change | New implementation | No change |
| Change YAML syntax | Change | No change | No change |
| Add report format (e.g., Allure) | No change | No change | Change |
| New command (e.g., `doubleTap`) | Parse it | All implement | No change |
| Command option for reporting only | Parse it | No change | Use it |

## Implementation Order

1. **Phase 1: Foundation**
   - [ ] Set up Go module
   - [ ] Define Executor interface (`pkg/executor/executor.go`)
   - [ ] Define Flow/Step models (`pkg/flow/`)
   - [ ] Define Report interface (`pkg/report/`)

2. **Phase 2: Appium Executor**
   - [ ] Create `pkg/executors/appium/`
   - [ ] Port WebDriver protocol from maestro-appium
   - [ ] Port selector resolver from maestro-appium
   - [ ] Implement Executor interface

3. **Phase 3: Flow Parser**
   - [ ] Port YAML parser from maestro-appium
   - [ ] Adapt to produce universal Steps
   - [ ] Add validation

4. **Phase 4: Reporter**
   - [ ] Port reporter from maestro-appium
   - [ ] Adapt to consume executor Results
   - [ ] Support JSON, JUnit, HTML, Allure

5. **Phase 5: CLI**
   - [ ] Create `cmd/maestro-runner/main.go`
   - [ ] Implement `run` command
   - [ ] Implement `validate` command
   - [ ] Add `--executor` flag

6. **Phase 6: Native Executor (Later)**
   - [ ] Create `pkg/executors/native/`
   - [ ] Implement UiAutomator2 client (Android)
   - [ ] Implement WDA client (iOS)
   - [ ] Implement Executor interface

## Reference Code

The `maestro-appium` project contains working code to reference:

```
/Users/omnarayan/work/go/src/maestro-appium/
├── pkg/parser/          → reference for pkg/flow/
├── pkg/engine/          → reference for step execution
├── pkg/adapter/         → reference for pkg/executors/appium/
├── pkg/selector/        → reference for selector resolution
├── pkg/reporter/        → reference for pkg/report/
└── pkg/jsengine/        → reference for JS execution
```

## Key Files to Port (from maestro-appium)

| From | To | Notes |
|------|-----|-------|
| `pkg/parser/parser.go` | `pkg/flow/parser.go` | YAML parsing |
| `pkg/parser/commands.go` | `pkg/flow/step.go` | Step definitions |
| `pkg/adapter/webdriver.go` | `pkg/executors/appium/protocol/` | WebDriver protocol |
| `pkg/selector/resolver.go` | `pkg/executors/appium/resolver.go` | Appium-specific |
| `pkg/reporter/reporter.go` | `pkg/report/reporter.go` | Report generation |

## Open Questions

1. Should Step contain all command types or use composition?
2. How to handle executor-specific options in YAML?
3. Event system design - push vs pull?
4. Plugin system for custom executors?

---

## Research: Maestro Issues Analysis

### Issue Statistics (as of 2024-12-21)

| Category | Count |
|----------|-------|
| Open Issues | 477 |
| Closed (not_planned) | 100 |
| Closed (duplicate) | 7 |
| Closed (completed) | 257 |

### Top Pain Points from GitHub Issues

| Category | Issues | Root Cause | Our Solution |
|----------|--------|------------|--------------|
| **Input Flakiness** | #2718, #2382, #2005, #1667 | Character-by-character input | Appium Unicode IME |
| **Animation Wait** | #2843, #2734, #1477 | Unreliable animation detection | Disable animations via Appium Settings |
| **Parallel Execution** | #2556, #1853, #2104 | Port 7001 hardcoded | Multi-session support |
| **Tap Failures** | #2448, #2326, #2062 | Hierarchy vs screenshot tap mismatch | Consistent tap strategy |
| **Assertions** | #2298, #2236 | Element visible but not found | Better wait strategies |
| **Timeout Ignored** | #2843, #684, #423 | Hardcoded timeouts | Configurable at all levels |
| **WebView** | #2293, #2064, #2585 | No ChromeDriver integration | Appium hybrid support |

### Maestro Architecture Limitations

```
MAESTRO DESIGN ISSUES
─────────────────────

1. PORT HARDCODING
   Android: gRPC port 7001 (hardcoded)
   iOS: HTTP port 22087
   → Cannot run parallel tests on same machine

2. MINIMAL DRIVER
   - Custom UiAutomator wrapper (not full UIAutomator2)
   - Missing: Unicode IME, Animation control, Notifications
   - Reinvents what Appium Settings already provides

3. FLAKY INPUT
   - Character-by-character via pressKeyCode()
   - No Unicode support
   - Drops/mangles characters

4. INCONSISTENT TAP
   - iOS: hierarchy-based
   - Android: screenshot-based (swapped in code!)
   - PR #2326 shows the bug

5. FEATURE REQUESTS IGNORED
   - #423 (configurable timeout) - open since 2022
   - #684 (custom timeout for conditions) - open since 2023
   - #1252 (test timeout) - open since 2023
```

### Design Implications

Based on Maestro issues research, problems to avoid in our design:

1. **Hardcoded values** - Ports, timeouts should be configurable
2. **Flaky operations** - Executor contract must define reliable behavior
3. **No parallel support** - Design must support multiple sessions
4. **Tight coupling** - Keep YAML, Executor, Report independent
5. **God classes** - Keep components small and focused

---

## Research: Maestro Source Code Analysis

### Key Files Analyzed

| File | Purpose | Flaw Found |
|------|---------|------------|
| `Orchestra.kt` | Command orchestrator | Hardcoded timeouts, 1500-line god class |
| `MaestroCommand.kt` | Command wrapper | Nullable field anti-pattern |
| `AndroidDriver.kt` | Host-side client | Port 7001 hardcoded |
| `MaestroDriverService.kt` | Device-side server | Character-by-character input |

### Flaws to Avoid

#### 1. Hardcoded Timeouts (Orchestra.kt)

```kotlin
class Orchestra(
    private val lookupTimeoutMs: Long = 17000L,        // Hardcoded!
    private val optionalLookupTimeoutMs: Long = 7000L, // Hardcoded!
)
```

**Lesson:** Make timeouts configurable at flow, step, and command level.

#### 2. Nullable Field Anti-Pattern (MaestroCommand.kt)

```kotlin
data class MaestroCommand(
    val tapOnElement: TapOnElementCommand? = null,
    val scrollCommand: ScrollCommand? = null,
    // ... 35+ nullable fields!
)
```

**Lesson:** Use proper type hierarchy (interface/sealed types), not nullable fields.

#### 3. God Class (Orchestra.kt - 1500+ lines)

```kotlin
return when (command) {
    is TapOnElementCommand -> ...
    is ScrollCommand -> ...
    // ... 50+ cases in one method!
}
```

**Lesson:** Keep components small, focused, and testable.

#### 4. Hardcoded Ports (AndroidDriver.kt)

```kotlin
private const val DefaultDriverHostPort = 7001  // No parallel execution!
```

**Lesson:** Design for parallel execution from the start.

### Summary: What We Learn from Maestro

| Maestro Flaw | Our Approach |
|--------------|--------------|
| Hardcoded timeouts | Configurable at all levels |
| Nullable command wrapper | Clean type hierarchy |
| 1500-line god class | Small, focused components |
| Hardcoded ports | Parallel-ready design |
| Tight coupling | 3 independent parts (YAML → Executor → Report) |

---

*Created: 2024-12-21*
*Last Updated: 2024-12-21*
*Status: Planning*
