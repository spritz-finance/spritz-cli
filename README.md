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
spritz auth login
spritz auth status
spritz bank-accounts list
```

### For Agents

```bash
spritz auth device start
spritz auth device complete
spritz auth status
spritz bank-accounts list
```

`spritz auth device start` returns JSON to stdout, including the auto-generated
device state file path. `spritz auth device complete` uses the only pending
session by default.

### For CI

```bash
export SPRITZ_API_KEY=ak_...
spritz auth status
```

## Agent Notes

If you are calling `spritz` from an agent or script:

- keep the default CSV output for flat resource reads; it is compact and token-efficient
- use JSON for auth handshakes and other object-shaped command results
- prefer `SPRITZ_API_KEY` or `spritz auth device` for non-interactive auth
- let `spritz auth device start` generate a unique state file unless your orchestrator needs an explicit path
- expect human-oriented warnings and status text on stderr where possible
- use `spritz auth status` to confirm which credential source is active

## Authentication

`spritz` supports three auth patterns:

1. Interactive browser login
2. Two-step device auth for agents and headless environments
3. `SPRITZ_API_KEY` for CI and secret managers

### Interactive Login

```bash
spritz auth login
```

This is the human-friendly login flow. It opens the browser-based device flow by
default and stores credentials locally.

### Two-Step Device Auth

```bash
spritz auth device start
spritz auth device complete
```

This is the preferred headless auth flow for agents.

- `start` creates a unique pending device session automatically and returns JSON
- `complete` finishes the only pending session by default
- if multiple pending sessions exist, rerun `complete` with `--device-state-file`

Example with explicit state path:

```bash
spritz auth device start --device-state-file /tmp/spritz-device.json
spritz auth device complete --device-state-file /tmp/spritz-device.json
```

### Environment Variable Auth

```bash
export SPRITZ_API_KEY=ak_...
spritz auth status
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
spritz auth status

# List bank accounts
spritz bank-accounts list

# Agent-friendly JSON output when needed
spritz bank-accounts list -o json

# Pipeline-friendly CSV output
spritz bank-accounts list --no-header

# Remove locally stored credentials
spritz auth logout
```

## Output Modes

Most commands support:

- `csv` by default for compact, agent-friendly tabular output
- `-o json` for programmatic consumers and richer structured payloads
- `-o table` for human-readable terminal output

`spritz auth device start` and `spritz auth device complete` always return JSON,
because they produce small structured auth handshakes rather than tabular data.

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
