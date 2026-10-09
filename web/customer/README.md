# Ocserv Dashboard Customer UI

The customer UI is the self-service Vue 3 application used by VPN users. It provides account summary, active sessions, activity, traffic statistics, certificate/Cisco downloads, and password management. It uses TypeScript, Vite, Pinia, Vue Router, Vue I18n, Tailwind CSS, and API service wrappers.

## Prerequisites

- A current Node.js LTS release with Corepack enabled.
- Yarn 4; the exact version is pinned in `package.json`.
- A running Customer API only when developing against the real API. The default API base URL is `/api`.

Install dependencies from this directory:

```bash
corepack enable
yarn install
```

## Run the UI

```bash
# Development server with the real API at /api
yarn dev

# Development server with API-shaped local mock data
yarn dev:test

# Type-check and create a production bundle
yarn build

# Preview a completed production build
yarn preview
```

Use `yarn dev:test` for UI work without a backend. It sets `NODE_ENV=test` and Vite `test` mode; mock mode can also be enabled with `VITE_USE_MOCKS=true`.

## Environment variables

Create an uncommitted `.env.local` file for local overrides:

```env
# Optional. Defaults are /api and 15000 respectively.
VITE_API_BASE_URL=http://localhost:8080/api
VITE_API_TIMEOUT_MS=15000

# Optional. Use local mocks instead of live customer API requests.
VITE_USE_MOCKS=true
```

All `VITE_` values are public browser-build values. Never add credentials, API tokens, or other secrets to them.

## Project map

```text
src/
  api/          HTTP client, authentication token handling, and API services
  components/   Reusable presentation components
  layouts/      Authenticated customer layout and navigation
  locales/      Translation messages, locale options, and RTL handling
  mocks/        API-shaped data for test mode
  router/       Customer routes and authentication guards
  stores/       Pinia authentication state
  views/        Route-level customer pages
```

Use the `@/` alias for imports from `src`. Route-level work belongs in `views`, shared page structure in `layouts`, reusable UI in `components`, and backend access in `api/services`.

## Add or change a UI feature

1. Add or update a page in `src/views/` and register its route in `src/router/index.ts`. Authenticated pages are children of `CustomerLayout`.
2. Reuse existing components and Tailwind styling. Keep the small-screen experience, keyboard interactions, empty states, and error messages usable.
3. Add API behavior in `src/api/services/`; do not make HTTP calls directly from a component. The shared client in `src/api/http.ts` applies the access token and handles unauthorized responses.
4. Update `src/mocks/` with realistic response data and mutations so the feature can be developed with `yarn dev:test`.
5. Add every visible string to `src/locales/index.ts`. English is the fallback. Preserve the message structure for each language and test RTL layout for Persian and Arabic.
6. Verify the login and authenticated flows, then check summary, sessions, activity, statistics, downloads, and password-related UI if the change affects shared layout or state.

## Quality checks

Run these commands before submitting a UI contribution:

```bash
yarn format
yarn type-check
yarn build
yarn build:test
```

There is currently no separate frontend test command. Manually test changed behavior in mock mode and, when available, against the Customer API. Do not commit `.env` files or build output. Include screenshots in a pull request when they clarify a visual change.

## Contributing translations

Translations, supported locales, persisted locale selection, and RTL direction are all maintained in `src/locales/index.ts`. Add English text first, then the matching keys for all supported languages. For a new language, add its complete messages and locale option; include it in RTL detection only when appropriate. Avoid raw user-facing strings in templates.
