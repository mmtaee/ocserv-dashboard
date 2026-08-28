---
name: dashboard-auth-bootstrap
description: Maintain the pre-mount system initialization and authentication flow in this Vue admin dashboard. Use when changing main.ts startup, system-init Pinia state, auth/setup routes, generated system API wrappers, login/setup forms, or captcha gating.
---

# Dashboard Auth Bootstrap

Preserve these startup invariants:

1. Call `GET /system/init` before mounting Vue and store its raw response in the Pinia system-init store.
2. Treat a request failure as unavailable, a null response or `first_init !== true` as setup-required, and `first_init === true` as initialized.
3. Route setup-required users to the `signup-02` system configuration view.
4. For initialized systems, validate an existing token with the generated profile API. Route authenticated users home and everyone else to `login-02`.
5. Render the captcha atom only when `google_captcha_site_key` is non-empty and submit its token through the login schema's `token` field.

Keep generated files in `src/api/generated` read-only. Put API behavior in `src/api/services`, state in `src/stores`, routed screens in `src/views`, and reusable UI in `src/components`. Type forms from generated request schemas and validate changes with `yarn type-check` and `yarn build`.
