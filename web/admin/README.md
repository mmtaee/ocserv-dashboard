# Ocserv Dashboard Admin UI

The admin UI is a Vue 3, TypeScript, and Vite application for managing an Ocserv Dashboard installation. It uses Pinia for state, Vue Router for navigation, Vue I18n for translations, Tailwind CSS for styling, and the generated Admin API client for backend contracts.

## Prerequisites

- A current Node.js LTS release with Corepack enabled.
- Yarn 4; the exact package manager version is pinned in `package.json`.
- A running backend only when developing against the real API. The default API base URL is `/api`.

From this directory, install dependencies:

```bash
corepack enable
yarn install
```

## Run the UI

```bash
# Development server with the real API at /api
yarn dev

# Development server with API-shaped local mock data; no network calls are made
yarn dev:test

# Type-check and create a production bundle
yarn build

# Preview a completed production build
yarn preview
```

`yarn dev:test` runs Vite in `test` mode. Mock mode can also be enabled with `VITE_USE_MOCKS=true`; it is useful when working on layouts, empty states, errors, charts, or interaction flows without a backend.

## Environment variables

Vite exposes variables prefixed with `VITE_` to browser code. Put local overrides in an uncommitted `.env.local` file:

```env
# Optional. Defaults are /api and 15000 respectively.
VITE_API_BASE_URL=http://localhost:8080/api
VITE_API_TIMEOUT_MS=15000

# Optional. Use local mocks instead of network requests.
VITE_USE_MOCKS=true
```

Never put credentials, tokens, or server-only secrets in a `VITE_` variable: they are included in the browser bundle.

## Project map

```text
src/
  api/          HTTP client, service wrappers, and generated Admin API client
  components/   Reusable feature and UI components
  composables/  Reusable Vue state and side-effect logic
  locales/      Translation messages and locale registration
  mocks/        API-shaped data used by test mode
  router/       Routes, navigation metadata, and route guards
  stores/       Pinia authentication, server, and setup state
  views/        Route-level pages
```

Use the `@/` alias for imports from `src`. Keep route pages in `views`, reusable feature UI in `components`, and API calls inside `api/services`; do not call Axios directly from a view or component.

## Add or change a UI feature

1. Add or update the route-level page in `src/views/` and register it in `src/router/`. Dashboard navigation metadata lives in `src/router/dashboard-routes.ts`.
2. Build reusable UI from the existing components in `src/components/ui/` before adding a new primitive. Use Tailwind utilities and match the established light/dark styling.
3. Add API behavior in `src/api/services/`. Keep authentication, error normalization, and request configuration in `src/api/http.ts`.
4. Add representative mock data and mutations in `src/mocks/` so the feature works with `yarn dev:test`.
5. Add all visible strings to the locales. English is the fallback; preserve the same message shape in every supported locale and keep RTL behavior correct for Persian and Arabic.
6. Check loading, empty, error, narrow-screen, keyboard, and dark-mode states before opening a pull request.

## API client generation

The Admin API client in `src/api/generated/` is generated from `../../backend/docs/swagger.json`. Do not hand-edit generated files. When the backend OpenAPI contract changes, regenerate it from this directory:

```bash
yarn codegen
```

Then update service wrappers, mocks, and UI code to match the generated types.

## Quality checks

Run these before submitting a UI change:

```bash
yarn format
yarn type-check
yarn build
yarn build:test
```

The project currently has no separate frontend test script. Exercise changed flows in both real-API mode when available and mock mode. Keep changes focused, avoid committing `.env` files or generated build output, and describe UI-visible changes with screenshots in the pull request when helpful.

## Contributing translations

Locale registration and message composition are in `src/locales/`. Add new text to English first, then provide matching keys for every supported locale. For a new language, add its messages, register it in the locale options, and update RTL handling only if the language is right-to-left. Do not use raw user-facing strings in templates when an existing translation key can be used.
