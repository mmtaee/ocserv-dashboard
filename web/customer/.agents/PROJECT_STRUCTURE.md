# Customer Frontend Project Structure

## Mandatory rules

Preserve this architecture. Do not add top-level source directories, move files, introduce another API client, or copy admin-only behavior without an explicit requirement. The live Swagger Customers tag is the source of truth.

## Stack and flow

The customer project is an independent Vue 3 + TypeScript + Vite application using Composition API, Vue Router, Pinia, Axios, Tailwind CSS, Reka UI, and vue-i18n. The main entry installs i18n, Pinia, router guards, and the shared 401 handler before mounting the application.

## Structure

- src/api/generated: types derived from current Customers Swagger schemas
- src/api/services: customer feature API operations
- src/api/auth-token.ts: token and expiry persistence
- src/api/http.ts: shared Axios client and interceptors
- src/components: shared customer components
- src/composables: reusable stateful feature logic
- src/layouts: authenticated application shells
- src/lib: framework-independent utilities
- src/locales: all seven locale catalogs and setup
- src/mocks: API-shaped mock behavior
- src/router: routes and authentication guards
- src/stores: cross-route Pinia state
- src/views: thin route-level components

## Placement

All protected calls use the shared bearer interceptor. Login is the only public customer endpoint. Keep API errors normalized at the HTTP boundary and feature state at its owning view or composable. All display text belongs in locales with en, it, zh-cn, zh-tw, ru, fa, and ar.

## Verification

Run yarn type-check, yarn build, and yarn build:test from this directory.
