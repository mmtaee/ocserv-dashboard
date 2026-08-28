---
name: dashboard-auth-bootstrap
description: Maintain the pre-mount system initialization and authentication flow in this Vue admin dashboard. Use when changing main.ts startup, system-init Pinia state, auth/setup routes, generated system API wrappers, login/setup forms, or captcha gating.
---

# Dashboard Auth Bootstrap

Preserve these startup invariants:

1. Call `GET /system/init` before mounting Vue and store its raw response in the Pinia system-init store.
2. Treat a request failure as unavailable and route to the server-unavailable view before making authentication or initialization routing decisions.
3. When the server is available, validate the stored access token with the generated profile API. Route unauthenticated users to `login-02` regardless of `first_init`.
4. After login stores a new token, call `GET /system/users/profile` and save that canonical response in the auth store. Do not use an embedded login-response user as the stored profile.
5. When `/login` is requested with a stored token, validate it through the profile API. Redirect a valid session home; clear an invalid session and remain on login.
6. Handle every API `401` in the shared Axios response interceptor: clear the token and auth-store user, then redirect to login. Register router/store behavior through a callback so the API client does not import application routing.
7. Only after authentication succeeds, read `first_init` from the system-init store. Route `first_init !== true` to the `signup-02` system configuration view and `first_init === true` home.
8. After login, availability retry, or setup submission, preserve the same availability → authentication → first-init routing order. After a successful setup update, merge the `PATCH /system` response into the system-init store; do not call `GET /system/init` again.
9. Pass `requireAuthorizationHeader()` to generated system endpoints that require an explicit authorization parameter. The shared Axios interceptor adds the current bearer token to other system requests whenever a token exists.
10. Render the captcha atom only when `google_captcha_site_key` is non-empty and submit its token through the login schema's `token` field.

Keep generated files in `src/api/generated` read-only. Put API behavior in `src/api/services`, state in `src/stores`, routed screens in `src/views`, and reusable UI in `src/components`. Type forms from generated request schemas and validate changes with `yarn type-check` and `yarn build`.

## Naming

- Never use `OCServ` in user-facing text or project-authored identifiers. Use `Ocserv` for names and titles, and `ocserv` for lowercase identifiers, commands, paths, and configuration keys.
- Do not rename generated OpenAPI symbols; their spelling is controlled by the backend schema.
