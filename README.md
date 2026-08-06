# spritz-cli

`spritz` is a CLI for Spritz fiat rails and human-operated account workflows.

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

For a headless host or local account agent, split approval from completion:

```bash
spritz auth device start --access user
# A human opens the returned URL and approves the requested user-account scopes.
spritz auth device complete
```

`start` returns JSON containing a unique state-file path, user code, verification
URL, and expiration. `complete` uses the only pending session unless
`--device-state-file` is supplied.

`spritz auth mcp --access user` is intentionally fail-closed. This release does
not pass a keychain credential to an MCP package launched through an ambient
Node/npm runtime: a same-user process or modified executable/cache could read
the key, so that design is not a credential-security boundary.

Until Spritz ships an integrity-verifiable packaged broker, use the read-only
MCP server only in an operator-controlled disposable Sandbox/test environment.
Treat the process as able to read its injected credential, scope the account to
non-production test data, and revoke or rotate the credential after the
session. Do not use this interim path with a Production End User account.

```bash
SPRITZ_API_KEY="${DISPOSABLE_SANDBOX_TEST_KEY}" \
SPRITZ_API_BASE_URL="https://sandbox.spritz.finance" \
npx -y @spritz-finance/mcp-server@0.3.2
```

## Developer workspace automation

The CLI does not currently broker Developer workspace credentials. Establish one
human-owned organization workspace in the
[Developer Console](https://console.spritz.finance/register), accept the current
terms as an authorized person, and follow the
[Developer Access guide](https://docs.spritz.finance/guides/developer-access) for
Sandbox HMAC credentials. An End User Bearer key is never a fallback.

Do not inject an organization HMAC credential into an agent-controlled process.
Until a scoped Developer workspace grant exists, have an operator run
credentialed integration tests outside the agent's control, or use a disposable
Sandbox credential and rotate it immediately after the test. An End User Bearer
key is never a Developer-workspace fallback.

Do not pass any key with `--api-key` in automation; argv may be captured in
shell history or process listings.

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
