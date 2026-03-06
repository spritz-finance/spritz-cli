# spritz-cli

Agent-optimized CLI for interacting with the Spritz API.

## Install

```bash
curl --proto '=https' --tlsv1.2 --fail --silent --show-error --location https://spritz.finance/install | bash
```

The installer verifies Sigstore-signed release checksums before installing the binary.

## Authentication

`spritz` supports three auth patterns:

1. Interactive login for humans
2. Device flow for agents and headless environments
3. `SPRITZ_API_KEY` for CI and secret managers

### Quickstart

```bash
# Human-friendly browser flow
spritz login

# Agent/headless flow
spritz login --device-start
spritz login --device-complete --json

# Secret-manager or CI flow
export SPRITZ_API_KEY=ak_...
spritz whoami -o json
```

### Security Notes

- Prefer `SPRITZ_API_KEY` from a secrets manager for CI and automation.
- Prefer stdin over `--api-key` when passing a key directly, since argv may be captured in shell history or process lists.
- Stored credentials live in the system keychain by default. Use `--allow-file-storage` only when keychain access is unavailable.
- `SPRITZ_API_KEY` always takes precedence over stored credentials.

See `docs/cli-auth-spec.md` for the device-flow contract and CLI auth behavior.
