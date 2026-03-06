# CLI Authentication

`spritz` supports interactive, headless, and fully non-interactive authentication.

## Recommended Flows

### Humans

```bash
spritz login
spritz whoami
```

Default `spritz login` opens a browser-based device flow and stores the resulting API key locally.

### Agents and Headless Environments

```bash
spritz login --device-start
spritz login --device-complete --json
```

`--device-complete` uses the pending local device session created by `--device-start` in the active config directory.

`--device-start` prints JSON to stdout:

```json
{
  "mode": "device_start",
  "envVarActive": false,
  "userCode": "ABCD1234",
  "verificationUri": "https://app.spritz.finance/device",
  "verificationUriComplete": "https://app.spritz.finance/device?code=ABCD1234",
  "expiresAt": "2026-03-06T12:10:00Z"
}
```

`--device-complete --json` prints structured success output:

```json
{
  "mode": "stored_credentials",
  "email": "user@example.com",
  "firstName": "Test",
  "storage": "system keychain",
  "envVarActive": false
}
```

### CI and Secret Managers

```bash
export SPRITZ_API_KEY=ak_...
spritz whoami -o json
```

This is the preferred pattern when a secret manager can inject environment variables directly.

## Security Guidance

- Prefer `SPRITZ_API_KEY` or secure stdin over `--api-key`.
- Treat `--api-key ak_...` as a last resort; command-line arguments may end up in shell history and process inspection tools.
- Credentials are stored in the system keychain by default.
- `--allow-file-storage` falls back to a machine-encrypted file when keychain storage is unavailable.
- `SPRITZ_API_KEY` always overrides stored credentials.

Safer direct-key example:

```bash
printf '%s' "$SPRITZ_API_KEY" | spritz login
```

## CLI Behavior

- `spritz login` is interactive and requires a TTY.
- `spritz login --device-start` is machine-readable and writes JSON to stdout.
- `spritz login --json` and `spritz logout --json` write structured JSON to stdout for automation.
- Human-oriented status and warnings are written to stderr where possible.
- `spritz whoami` shows the active user plus the credential source.

Example `spritz whoami -o json` output:

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

## Logout

```bash
spritz logout
spritz logout --json
```

Logout removes locally stored credentials only. It does not revoke the server-side API key. To revoke it, visit:

`https://app.spritz.finance/settings/api-keys`

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
