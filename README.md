# maestro-runner

**3.6x faster** · **14x less memory** · **single binary, no JVM** · your existing [Maestro](https://maestro.mobile.dev/) YAML flows work as-is — with the features and fixes Maestro hasn't delivered.

## Quick Start

### Install

```bash
go install github.com/devicelab-dev/maestro-runner@latest
```

Or download a pre-built binary from [releases](https://devicelab.dev/open-source/maestro-runner).

### Run

```bash
maestro-runner flow.yaml                              # Android (default)
maestro-runner flow.yaml --platform ios               # iOS
maestro-runner flows/                                 # All flows in a directory
maestro-runner --driver appium flow.yaml              # Appium (local or cloud)
```

## Why maestro-runner?

If you've hit any of these, maestro-runner fixes them:

- **`inputText` drops characters** → direct ADB input, reliable Unicode support
- **Tests are slow** → native element selectors, no polling, configurable idle timeouts
- **Can't configure timeouts** → per-command and per-flow, `--wait-for-idle-timeout 0` to disable
- **Can't run parallel** → dynamic ports, multiple instances on one machine
- **JVM eats memory in CI** → ~21 MB Go binary vs ~289 MB JVM
- **No cloud support** → BrowserStack, Sauce Labs, LambdaTest via Appium
- **Elements not found reliably** → clickable parent traversal, native `textMatches()` regex, smarter visibility checks

Addresses [78% of the top 100 most-discussed open issues](docs/maestro-issues-analysis.md) on Maestro's GitHub.

## Requirements

- Go 1.22+
- **Android:** `adb` (Android SDK Platform-Tools)
- **iOS:** Xcode command-line tools (`xcrun`)
- **Appium:** Appium server 2.x (`npm i -g appium`)

## Drivers

| Driver | Platform | Description |
|--------|----------|-------------|
| **UIAutomator2** | Android | Direct connection via UIAutomator2. Default, no external server needed. |
| **WDA** | iOS | Auto-selected with `--platform ios`. Uses WebDriverAgent. |
| **Appium** | Both | `--driver appium`. For cloud providers and custom setups. |

## Flow Config

maestro-runner extends Maestro's flow config with two additional fields:

```yaml
commandTimeout: 10000       # Default per-command timeout (ms)
waitForIdleTimeout: 3000    # Device idle wait (ms), 0 to disable
---
- launchApp: com.example.app
- tapOn: "Login"
```

## Documentation

| Document | Description |
|----------|-------------|
| **[CLI Reference](docs/cli-reference.md)** | All commands, flags, environment variables, tag filtering, parallel execution, emulator/simulator management |
| **[Flow Commands](docs/flow-commands.md)** | Complete flow YAML reference — selectors, tap/gesture, text input, assertions, app lifecycle, flow control, scripting |
| **[Technical Approach](docs/technical-approach.md)** | Driver architecture, element finding strategy, server lifecycles, report system, internals |

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

Apache License 2.0 — see [LICENSE](LICENSE).
