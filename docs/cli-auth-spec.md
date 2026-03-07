# CLI Authentication

`spritz` supports interactive, headless, and fully non-interactive authentication.

The CLI has two auth shapes:

- human-friendly commands under `spritz auth login`
- agent-friendly device flows under `spritz auth device`

## Recommended Flows

### Humans

```bash
spritz auth login
spritz auth status
```

Default `spritz auth login` opens a browser-based device flow and stores the resulting API key locally.

### Agents and Headless Environments

```bash
spritz auth device start
spritz auth device complete
```

`spritz auth device start` creates a unique pending device session automatically and prints JSON to stdout:

```json
{
  "mode": "device_start",
  "envVarActive": false,
  "deviceStateFile": "/home/user/.config/spritz/device/device-20260307T120000Z-a1b2c3d4.json",
  "userCode": "ABCD1234",
  "verificationUri": "https://app.spritz.finance/device",
  "verificationUriComplete": "https://app.spritz.finance/device?code=ABCD1234",
  "expiresAt": "2026-03-07T12:10:00Z"
}
```

`spritz auth device complete` completes the only pending session by default. If multiple pending sessions exist, the CLI requires an explicit `--device-state-file`.

Example with explicit state path:

```bash
spritz auth device start --device-state-file /tmp/spritz-device.json
spritz auth device complete --device-state-file /tmp/spritz-device.json
```

Successful completion prints structured JSON to stdout:

```json
{
  "mode": "stored_credentials",
  "email": "user@example.com",
  "firstName": "Test",
  "storage": "system keychain",
  "envVarActive": false,
  "deviceStateFile": "/home/user/.config/spritz/device/device-20260307T120000Z-a1b2c3d4.json"
}
```

### CI and Secret Managers

```bash
export SPRITZ_API_KEY=ak_...
spritz auth status
```

This is the preferred pattern when a secret manager can inject environment variables directly.

## CLI Behavior

- `spritz auth login` is interactive and requires a TTY unless credentials are piped on stdin or passed via `--api-key`
- `spritz auth device start` always writes JSON to stdout
- `spritz auth device complete` always writes JSON to stdout
- `spritz auth status` uses the normal CLI output modes, with CSV as the default
- human-oriented status and warnings are written to stderr where possible
- `SPRITZ_API_KEY` always overrides stored credentials

Example `spritz auth status -o json` output:

```json
[
  {
    "email": "user@example.com",
    "firstName": "Test",
    "source": "environment variable",
    "envOverride": "true",
    "storedCredentials": "true"
  }
]
```

## Security Guidance

- prefer `SPRITZ_API_KEY` or secure stdin over `--api-key`
- treat `--api-key ak_...` as a last resort; command-line arguments may end up in shell history and process inspection tools
- credentials are stored in the system keychain by default
- `--allow-file-storage` falls back to a machine-encrypted file when keychain storage is unavailable
- only approved destinations and explicitly authorized payment flows should be used by agents

Safer direct-key example:

```bash
printf '%s' "$SPRITZ_API_KEY" | spritz auth login
```

## Logout

```bash
spritz auth logout
spritz auth logout --json
```

Logout removes locally stored credentials only. It does not revoke the server-side API key. To revoke it, visit:

`https://app.spritz.finance/settings/api-keys`

## Why Device State Exists

The device flow is split into two commands:

1. `spritz auth device start`
2. `spritz auth device complete`

The first command receives temporary device authorization state from the API. The second command needs that state in order to finish polling and redeem the approved API key.

The CLI stores this state in a file so parallel agent runs do not clobber one another.

- by default, `start` generates a unique file path automatically
- `complete` uses the only pending session when there is exactly one
- if multiple pending sessions exist, the CLI requires `--device-state-file`

This gives agents a simple default path without falling back to unsafe singleton state.

## Device Authorization Protocol

The browser flow is based on [RFC 8628](https://datatracker.ietf.org/doc/html/rfc8628). Endpoints are verified against the Spritz OpenAPI spec.

### Flow

```
CLI                              Browser                         API
 │                                                                │
 ├─ POST /v1/device/authorize ──────────────────────────────────►│
 │◄─── deviceCode, userCode, verificationUriComplete ────────────┤
 │                                                                │
 ├─ open(verificationUriComplete) ──►│                            │
 │                                   ├─ GET /v1/device/info ────►│
 │                                   │◄─── clientId, expiresIn ──┤
 │                                   │                            │
 │                                   │  user picks permissions,   │
 │                                   │  expiry, key name          │
 │                                   │                            │
 │                                   ├─ POST /v1/device/approve ►│
 │                                   │◄─── approved: true ───────┤
 │                                   │                            │
 ├─ POST /v1/device/token (poll) ──────────────────────────────►│
 │◄─── apiKey, permissions ──────────────────────────────────────┤
 │                                                                │
 ╰─ store apiKey locally                                          │
```

### Endpoints

#### POST /v1/device/authorize (unauthenticated)

Request:

```json
{ "client_id": "spritz-cli" }
```

Response 200:

```json
{
  "deviceCode": "string",
  "userCode": "ABCD1234",
  "verificationUri": "https://app.spritz.finance/device",
  "verificationUriComplete": "https://app.spritz.finance/device?code=ABCD1234",
  "expiresIn": 600,
  "interval": 5
}
```

#### GET /v1/device/info?user_code=ABCD1234 (Cognito JWT)

Response 200:

```json
{
  "clientId": "spritz-cli",
  "expiresIn": 480,
  "createdAt": "2026-03-06T12:00:00Z"
}
```

#### POST /v1/device/approve (Cognito JWT)

Request:

```json
{
  "user_code": "ABCD1234",
  "permissions": ["bank-accounts:read", "off-ramp-quotes:write"],
  "expires_at": "2027-03-06T00:00:00Z",
  "name": "my-laptop"
}
```

Valid permissions: `bank-accounts:read`, `bank-accounts:write`, `bank-accounts:delete`, `bills:read`, `bills:delete`, `off-ramp-quotes:write`.

Response 200:

```json
{
  "approved": true,
  "clientId": "spritz-cli",
  "permissions": ["bank-accounts:read", "off-ramp-quotes:write"],
  "expiresAt": "2027-03-06T00:00:00Z"
}
```

#### POST /v1/device/token (unauthenticated)

Request:

```json
{
  "device_code": "<deviceCode from step 1>",
  "grant_type": "urn:ietf:params:oauth:grant-type:device_code"
}
```

Response 200 (after approval):

```json
{
  "apiKey": "ak_...",
  "keyId": "string",
  "permissions": ["bank-accounts:read", "off-ramp-quotes:write"],
  "expiresAt": "2027-03-06T00:00:00Z",
  "keyName": "my-laptop"
}
```

Error responses use RFC 9457 problem details (400):

| `detail` | Meaning |
|----------|---------|
| `authorization_pending` | User has not approved yet; keep polling |
| `slow_down` | Polling too fast; increase interval by 5s |
| `expired_token` | Device code expired; start over |
