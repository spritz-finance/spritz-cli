# CLI Auth Spec — Device Authorization Flow

Based on [RFC 8628](https://datatracker.ietf.org/doc/html/rfc8628).

## Flow

```
CLI                              Browser                         API
 │                                                                │
 ├─ POST /device/authorize (client_id only) ────────────────────►│
 │◄─── deviceCode, userCode, verificationUriComplete ────────────┤
 │                                                                │
 ├─ open(verificationUriComplete) ──►│                            │
 │                                   ├─ GET /device/info ────────►│
 │                                   │◄─── clientId, expiresIn ──┤
 │                                   │                            │
 │                                   │  user picks permissions,   │
 │                                   │  expiry, key name          │
 │                                   │                            │
 │                                   ├─ POST /device/approve ───►│
 │                                   │◄─── approved: true ───────┤
 │                                   │                            │
 ├─ POST /device/token (poll) ──────────────────────────────────►│
 │◄─── apiKey, permissions ──────────────────────────────────────┤
 │                                                                │
 ╰─ store apiKey locally                                          │
```

## Endpoints

### Step 1: CLI initiates authorization

**`POST /device/authorize`** (unauthenticated)

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

### Step 2: Browser fetches request info

**`GET /device/info?user_code=ABCD1234`** (Cognito JWT required)

Response 200:
```json
{
  "clientId": "spritz-cli",
  "expiresIn": 600,
  "createdAt": "2026-03-06T12:00:00Z"
}
```

### Step 3: User approves in browser

**`POST /device/approve`** (Cognito JWT required)

Request:
```json
{
  "user_code": "ABCD1234",
  "permissions": ["bank-accounts:read", "off-ramp-quotes:write"],
  "expires_at": "2027-03-06T00:00:00Z",
  "name": "my-laptop"
}
```

Response 200:
```json
{
  "approved": true,
  "clientId": "spritz-cli",
  "permissions": ["bank-accounts:read", "off-ramp-quotes:write"],
  "expiresAt": "2027-03-06T00:00:00Z"
}
```

### Step 4: CLI polls for the API key

**`POST /device/token`** (unauthenticated)

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

Error responses (400):
| `error` | Meaning |
|---------|---------|
| `authorization_pending` | User hasn't approved yet — keep polling |
| `slow_down` | Polling too fast — increase interval by 5s |
| `expired_token` | Device code expired — start over |

## CLI Behavior

- `deviceCode` is secret — never displayed to the user
- `userCode` is displayed in the terminal so the user can verify it matches the browser
- CLI auto-opens `verificationUriComplete` in the default browser
- If browser fails to open, the URL is printed to stderr
- CLI polls `/device/token` every `interval` seconds
- On `slow_down`, interval increases by 5 seconds
- On `expired_token` or timeout, CLI exits with error
