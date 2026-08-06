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

Spritz has two distinct agent principals:

- An **End User account agent** acts on the approving person's own Spritz
  account with a scoped Bearer credential. The current CLI, MCP server, and
  off-ramp tools use this path.
- A **Developer workspace agent** helps a business build an embedded
  integration with organization-owned HMAC credentials. That separate device
  grant is not deployed and must not borrow an End User credential.

The `--access` value names the credential owner. It is not an environment or a
pricing tier. Commands that broker a credential require the principal to be
explicit.

## Existing user-account flow

The existing user flow remains available for a person managing their own Spritz
account:

```bash
spritz auth login
spritz auth status
spritz bank-accounts list
```

For a headless host or local account agent, split approval from completion and
then launch the MCP server through the fixed broker:

```bash
spritz auth device start --access user
# A human opens the returned URL and approves the requested user-account scopes.
spritz auth device complete
spritz auth mcp --access user
```

`start` returns JSON containing a unique state-file path, user code, verification
URL, and expiration. `complete` uses the only pending session unless
`--device-state-file` is supplied.

The broker validates the active key as an End User credential, derives an
allowlisted Sandbox or Production API origin from the CLI configuration, clears
credential and Node/npm injection overrides, and passes the key only in the
fixed MCP child process environment. It never places the key on argv or writes
it to MCP configuration.

## Future Developer workspace-agent flow

The future entry points are reserved explicitly:

```bash
spritz auth device start --access developer
spritz auth device complete
spritz auth mcp --access developer
```

Both commands fail closed until the Developer Console can issue and attest a
workspace-, environment-, scope-, and expiry-bound HMAC grant. An End User
Bearer key is never a fallback.

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
`SPRITZ_API_KEY` always overrides stored credentials. Revoke End User account
keys at [app.spritz.finance/api-keys](https://app.spritz.finance/api-keys).

## Development

```bash
make test
```

If `go` is not on `PATH`, the Makefile falls back to `mise exec -- go`.

See [`docs/cli-auth-spec.md`](docs/cli-auth-spec.md) for the device protocol,
credential precedence, and platform contract.
