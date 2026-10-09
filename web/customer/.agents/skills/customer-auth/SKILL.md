---
name: customer-auth
description: Maintain customer JWT login, restoration, expiry, guards, and unauthorized handling.
---

Use src/stores/auth.ts as the sole cross-route auth state. Persist token and expires_at only through src/api/auth-token.ts; expired tokens must be rejected locally. Attach bearer tokens centrally, validate restored sessions with customer summary, clear state on 401, and keep login public while guarding all application routes.
