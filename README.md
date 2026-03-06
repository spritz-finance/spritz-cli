# spritz-cli

`spritz` is the CLI that gives your agents access to your Spritz account.

With `spritz`, an agent can move beyond wallet-only actions and use real fiat rails:
authenticate, inspect approved bank destinations, create off-ramp quotes, and
coordinate crypto-to-fiat payouts through Spritz.

If your agent can control a wallet, `spritz` gives it a path to move funds into
the banking system through your Spritz account.

## Why It Matters

Most agent tooling stops at crypto rails. Agents can read data, call APIs, and
sign onchain transactions, but they usually cannot reach fiat payout systems.

`spritz` closes that gap. It gives agents a scriptable interface to your Spritz
account so they can participate in real payment workflows, not just crypto-native
ones.

## Install

```bash
curl -fsSL https://spritz.finance/install | bash
```

The installer verifies Sigstore-signed release checksums before installing the binary.

## Quickstart

### For Humans

```bash
spritz login
spritz whoami
spritz bank-accounts list
```

### For Agents

```bash
spritz login --device-start --device-state-file /tmp/spritz-device.json
spritz login --device-complete --device-state-file /tmp/spritz-device.json --json
spritz whoami -o json
spritz bank-accounts list -o json
```

### For CI

```bash
export SPRITZ_API_KEY=ak_...
spritz whoami -o json
```

## What Agents Can Do

Today `spritz` supports agent-friendly access to:

- authentication in terminals, headless runtimes, and CI
- bank account discovery for approved off-ramp destinations
- machine-readable command output for scripts and orchestration

As the CLI expands, the goal is straightforward: let agents safely operate real
Spritz payment workflows from the command line.

## Agent Notes

If you are calling `spritz` from an agent or script:

- prefer `-o json` when you need machine-readable output
- prefer `SPRITZ_API_KEY` or the two-step device flow for non-interactive auth
- prefer an explicit `--device-state-file` so parallel runs do not rely on hidden local state
- expect human-oriented warnings and status text on stderr where possible
- use `spritz whoami` to confirm which credential source is active

## Authentication

`spritz` supports three auth patterns:

1. Interactive browser login
2. Two-step device auth for agents and headless environments
3. `SPRITZ_API_KEY` for CI and secret managers

### Interactive Login

```bash
spritz login
```

This opens the browser-based device flow and stores credentials locally.

### Two-Step Device Auth

```bash
spritz login --device-start --device-state-file /tmp/spritz-device.json
spritz login --device-complete --device-state-file /tmp/spritz-device.json --json
```

`--device-start` writes machine-readable JSON to stdout. `--device-complete` uses
the same state file to finish the flow.

### Environment Variable Auth

```bash
export SPRITZ_API_KEY=ak_...
spritz whoami -o json
```

This is the preferred auth pattern for CI and secret-managed runtimes.

### Security Notes

- prefer `SPRITZ_API_KEY` from a secrets manager for CI and automation
- prefer stdin over `--api-key` when passing a key directly, since argv may be captured in shell history or process inspection tools
- stored credentials live in the system keychain by default
- use `--allow-file-storage` only when keychain access is unavailable
- `SPRITZ_API_KEY` always takes precedence over stored credentials
- only approved destinations and explicitly authorized payment flows should be used by agents

See `docs/cli-auth-spec.md` for the full auth and device-flow contract.

## Common Commands

```bash
# Show the active authenticated user
spritz whoami

# List bank accounts
spritz bank-accounts list

# Agent-friendly output
spritz bank-accounts list -o json

# Pipeline-friendly output
spritz bank-accounts list --no-header

# Remove locally stored credentials
spritz logout
```

## Output Modes

Most commands support:

- `-o json` for agents and programmatic consumers
- `-o csv` for shell pipelines and exports
- `-o table` for human-readable terminal output

## Development

Run the test suite locally with:

```bash
make test
```

If `go` is not already on your `PATH`, the Makefile falls back to `mise exec -- go`.

## Documentation

- auth and device flow: `docs/cli-auth-spec.md`
- test workflow: `.github/workflows/test.yml`
- release workflow: `.github/workflows/release.yml`
