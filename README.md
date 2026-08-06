# spritz-cli

`spritz` is a CLI for Spritz fiat rails. It supports human-operated account
workflows and gives agent tools a credential broker that keeps API keys off
command lines and out of plaintext configuration files.

## Install

```bash
curl -fsSL https://spritz.finance/install | bash
```

The installer verifies Sigstore-signed release checksums before installing the
binary.

## Access model

Developer Access and Production are deliberately separate:

- **Developer Access** is a pre-production access path, not a pricing tier or
  commercial deal.
- **Sandbox** uses simulated funds and can begin after a human accepts the
  Developer Terms for the legal entity.
- **Live Test** uses real money only within approved limits after preliminary
  business verification and compliance approval.
- **Production** is a separate commercial service that requires full business
  verification and signed commercial agreements.

An AI agent cannot accept legal terms, complete business verification, or own a
Production credential. A human workspace administrator approves a scoped device
grant; the agent receives only the resulting capability through the local
credential broker.

The current platform device endpoint authorizes an individual Spritz user
account at `app.spritz.finance`. It is not a Developer Access workspace grant.
The CLI exposes the future mode explicitly and fails closed instead of silently
substituting the user flow:

```bash
spritz auth device start --access developer
```

Until the Developer Console endpoints are deployed, this returns a clear
not-enabled error and directs the human administrator to
`https://console.spritz.finance`.

## Existing user-account flow

The existing user flow remains available for a person managing their own Spritz
account:

```bash
spritz auth login
spritz auth status
spritz bank-accounts list
```

For a headless host, split approval from completion:

```bash
spritz auth device start
# A human opens the returned URL and approves the requested user-account scopes.
spritz auth device complete
spritz auth status
```

`start` returns JSON containing a unique state-file path, user code, verification
URL, and expiration. `complete` uses the only pending session unless
`--device-state-file` is supplied.

## Developer agent credential broker

Once the workspace grant contract is deployed, a human administrator completes
the Developer Access device flow and then starts agent tooling through the
broker:

```bash
spritz auth device start --access developer
spritz auth device complete
spritz auth mcp
```

Until the platform contract is deployed, both `--access developer` and
`auth mcp` return the typed not-enabled error before reading or launching
anything. That is intentional: locally asserted metadata cannot prove a
workspace grant and must not promote an End User key.

Once the platform supplies issuance and introspection, `auth mcp` will inject the
verified scoped credential only into a fixed Spritz MCP child process. It will
not put the key on argv, write a plaintext credential file, or accept an
arbitrary child command that could print the credential.

For CI, a secret manager may explicitly inject `SPRITZ_API_KEY`:

```bash
export SPRITZ_API_KEY="${CI_SECRET_VALUE}"
npx -y @spritz-finance/mcp-server
```

Do not pass a key with `--api-key` in automation; argv may be captured in shell
history or process listings.

## Common commands

```bash
spritz auth status
spritz bank-accounts list
spritz bank-accounts list -o json
spritz auth logout
```

Most commands default to compact CSV, with `-o json` and `-o table` available.
Device handshake commands always emit JSON.

Stored credentials use the system keychain by default. `--allow-file-storage`
enables a machine-encrypted fallback only when a keychain is unavailable.
`SPRITZ_API_KEY` always overrides stored credentials. Revoke keys in the
[Developer Console](https://console.spritz.finance/settings/api-keys).

## Development

```bash
make test
```

If `go` is not on `PATH`, the Makefile falls back to `mise exec -- go`.

See [`docs/cli-auth-spec.md`](docs/cli-auth-spec.md) for the device protocol,
credential precedence, and platform contract.
