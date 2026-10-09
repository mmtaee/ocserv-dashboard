# Frontend Project Structure

This document describes the current `web/admin` frontend architecture. It is a reference for future AI agents, not a redesign proposal.

## Mandatory rules for future agents

1. Always inspect this `PROJECT_STRUCTURE.md` before making frontend changes.
2. Preserve the existing project architecture.
3. Do not create new top-level folders unless strictly required by the existing architecture.
4. Do not move or reorganize existing files without an explicit task requirement.
5. Do not introduce a new architectural pattern when an existing project pattern already solves the problem.
6. Reuse existing components, composables, services, utilities, types, and patterns whenever possible.
7. Place new files beside equivalent existing files.
8. Follow the existing naming and import conventions.
9. Keep feature changes scoped and avoid unrelated refactors.

## Current stack and entry flow

- Vue 3 single-file components use Composition API and `<script setup lang="ts">`.
- Vite builds the application; TypeScript is checked with `vue-tsc`.
- Vue Router provides client-side routing.
- Pinia holds application-wide state.
- Axios and the generated OpenAPI client provide HTTP access.
- Tailwind CSS v4, shadcn-vue source components, and Reka UI provide styling and UI primitives.
- vue-i18n provides localization; Lucide Vue provides icons.

`src/main.ts` initializes theme handling, i18n, Pinia, system availability, restored authentication, router guards, and the global unauthorized handler before mounting `App.vue`. `App.vue` provides the reactive LTR/RTL `ConfigProvider`, renders `RouterView`, and includes the application footer.

## Main directory tree

```text
web/admin/
├── .agents/
│   ├── PROJECT_STRUCTURE.md
│   └── skills/                    # Local frontend instructions for AI agents
├── components.json               # shadcn-vue aliases, theme, icon, and RTL settings
├── package.json                  # Yarn scripts and frontend dependencies
├── tsconfig.app.json             # TypeScript application config and @ alias
├── vite.config.ts                # Vue/Tailwind Vite plugins and @ alias
└── src/
    ├── App.vue                   # Root provider and router outlet
    ├── main.ts                   # Application bootstrap
    ├── style.css                 # Tailwind imports and global design tokens
    ├── vite-env.d.ts
    ├── api/
    │   ├── generated/            # OpenAPI-generated client, models, and configuration
    │   ├── services/             # Feature-facing API functions and mock-mode branches
    │   ├── auth-token.ts         # Access-token persistence/header helpers
    │   ├── client.ts             # Generated API instance registry
    │   ├── environment.ts        # Test/mock-mode detection
    │   ├── http.ts               # Shared Axios instance, errors, and interceptors
    │   └── index.ts              # API exports
    ├── assets/                   # Static imported images and SVGs
    ├── components/
    │   ├── atoms/                # Small application-specific reusable components
    │   ├── blocks/               # Imported/composed UI blocks
    │   ├── dashboard/            # Dashboard feature components
    │   ├── occtl/                # OCCTL feature components and helpers
    │   ├── ocserv-groups/        # Ocserv group feature components and helpers
    │   ├── ocserv-sync/          # Ocserv sync feature components
    │   ├── ocserv-users/         # Ocserv user feature components
    │   ├── ui/                   # Shared shadcn-vue/Reka UI primitives
    │   └── *.vue                 # Application shell/navigation components
    ├── composables/              # Reusable Vue state and side-effect orchestration
    ├── lib/                      # Framework-independent shared utilities
    ├── locales/                  # Existing locale catalogs and i18n setup
    ├── mocks/                    # API-shaped test-mode data and behavior
    ├── router/                   # Route metadata, mapping, and guards
    ├── stores/                   # Cross-route Pinia stores
    └── views/                    # Route-level composition components
```

The local `.agents/skills` directory currently contains guidance for Vue composition, routing, Pinia, i18n, templates, debugging, pagination, composables, shadcn-vue, and dashboard authentication/bootstrap. Read the applicable skill before changing that area.

## Pages and feature modules

Route-level files live in `src/views` and use the `*View.vue` suffix. Dashboard routes normally render `DashboardLayout` and compose one feature page component, for example:

```text
views/OcservGroupsView.vue
  -> components/DashboardLayout.vue
  -> components/ocserv-groups/OcservGroupsPage.vue
```

Feature UI belongs in `src/components/<feature>/`. Existing feature folders use PascalCase component names prefixed by the feature (`OcservGroupsTable.vue`, `OcservUserEditorSheet.vue`, `OcctlResultViewer.vue`). Pure feature helpers stay beside those components as kebab-case `.ts` files such as `group-config.ts`, `default-group-fields.ts`, and `occtl-commands.ts`.

Current placement rules:

- Keep views thin; they compose layouts and feature page components.
- Put substantial feature markup in `components/<feature>/`, not in a route view.
- Put reusable application-shell components directly under `components/`.
- Put generic UI primitives in `components/ui/<primitive>/` and export them from that primitive's `index.ts`.
- Put small application-specific primitives in `components/atoms/`.
- Reuse existing dialogs, sheets, tables, fields, alerts, empty states, skeletons, spinners, cards, and other `components/ui` primitives before adding another implementation.
- Add a new feature file beside the equivalent files for that feature.

## Routing conventions

- `src/router/dashboard-routes.ts` is the source of dashboard navigation metadata: `path`, route `name`, translation `titleKey`, navigation `sectionKey`, Lucide `icon`, and `adminVisible`.
- `src/router/index.ts` creates browser-history routing, lazy-loads views, maps dashboard route names to implemented views, and falls back to `EmptyRouteView` for declared routes without an implementation.
- Login, setup, server-unavailable, and catch-all routes are declared directly in `router/index.ts`.
- `adminVisible: false` becomes `meta.superadminOnly: true`.
- Global guards enforce server availability, authentication, initial setup, and superadmin access.
- New dashboard pages should add route metadata and a lazy view mapping using the existing route-name conventions.

## API and service conventions

- `src/api/generated/` is generated by `yarn codegen` from `openapi-config.json`. Do not hand-edit generated files.
- `src/api/client.ts` constructs configured generated API classes with the shared Axios instance and exposes them through the `api` object.
- `src/api/http.ts` owns the base URL, timeout, `ApiError`, `normalizeApiError`, bearer-token interceptor, 401 handling, and the shared `httpClient`.
- `src/api/auth-token.ts` owns token storage and authorization-header helpers.
- Components and composables call functions in `src/api/services/*.ts`; they do not construct Axios clients or generated API classes themselves.
- Service files expose domain-friendly type aliases/interfaces and small async functions. Generated request/response models are reused when accurate; narrow local adapters document and normalize known backend/generated-client mismatches.
- Services use `requireAuthorizationHeader()` for generated endpoints that require an explicit authorization parameter. The shared Axios interceptor supplies bearer headers for direct `httpClient` calls.
- API failures are converted with `normalizeApiError` at the state/composable or component orchestration boundary.
- Existing test mode is detected by `src/api/environment.ts`. Service functions keep the same public contract in production and test mode; feature mock implementation remains under `src/mocks`.

## Composables and state management

There is no `hooks` directory. Vue hooks are named composables and live in `src/composables`:

- `use-theme.ts` manages global theme behavior.
- `useDashboardStats.ts` orchestrates dashboard requests.
- `useOcservGroups.ts` and `useOcservUsers.ts` own feature request state, loading/mutation flags, feedback, pagination, and actions.

Composables use `useXxx` naming, Composition API refs/computed values, explicit async actions, and read-only exposed state where appropriate. Keep local page-only form/dialog state inside the feature component. Extract a composable when stateful behavior is reused or when it matches an existing feature orchestration pattern.

Pinia stores live in `src/stores` and are reserved for state shared across routes or required during bootstrap. The current stores are `auth.ts` and `system-init.ts`; both use setup-store syntax with `defineStore`, refs/computed values, and explicit actions. Do not add a global store for state that belongs to one page.

## Types and interfaces

- Prefer generated request, response, enum, and model types from `@/api/generated`.
- Service modules export domain aliases and adapter interfaces used by components/composables.
- Component props, emits, and models are typed locally with `defineProps`, `defineEmits`, and `defineModel`.
- Shared route metadata interfaces stay beside route metadata (`DashboardRoute` in `router/dashboard-routes.ts`).
- Small helper-only interfaces stay in their owning `.ts` or `.vue` file; the project has no separate global `types` folder.
- Mock objects use TypeScript `satisfies` against real API/domain types.

## Styling and UI conventions

- `src/style.css` is the only global stylesheet. It imports Tailwind CSS and `tw-animate-css`, defines light/dark semantic color tokens, radii, fonts, chart colors, sidebar colors, and base element rules.
- `components.json` configures shadcn-vue with the `new-york` style, Inter font, Zinc base, CSS variables, Lucide icons, and RTL support.
- Use semantic Tailwind utilities such as `bg-background`, `text-muted-foreground`, `border-border`, and component variants instead of hard-coded palette colors.
- Use Tailwind utilities for layout and responsive behavior. Existing components favor flex/grid plus `gap-*`, logical direction-safe utilities, and shared component variants.
- Icons come from `@lucide/vue`; icons inside buttons use the existing `data-icon` placement convention.
- Forms reuse `components/ui/field`, inputs/selects/checkboxes, and the project's validation attributes.
- Preserve accessibility labels, dialog/sheet titles, loading states, and empty/error feedback.
- RTL direction comes from `App.vue`; `DashboardLayout.vue` places the sidebar on the physical side matching the active locale.

## i18n structure and rules

All user-facing text must use the existing vue-i18n structure under `src/locales`. Supported locales are:

- `en`
- `it`
- `zh-cn`
- `zh-tw`
- `ru`
- `fa`
- `ar`

The base catalogs are `en.ts`, `it.ts`, `zh-cn.ts`, `zh-tw.ts`, `ru.ts`, `fa.ts`, and `ar.ts`. English defines `MessageSchema`; the other base catalogs use `satisfies MessageSchema` to enforce matching structure. Existing larger feature catalogs (`group-defaults.ts`, `occtl.ts`, `ocserv-groups.ts`, and `ocserv-users.ts`) contain all seven locale variants and are merged by `locales/index.ts`.

- Use the global Composition API composer: `useI18n({ useScope: "global" })`.
- Add each new key with real translated text for every supported locale.
- Keep keys structurally identical across locales.
- Put text into the existing base catalog or the existing feature catalog that owns it; do not create another locale structure or translation system.
- Do not translate unstable backend error payloads without a stable error-code contract.
- `locales/index.ts` controls allowed language configuration, persisted selection, document language/title, and RTL detection. Persian and Arabic are RTL.
- After locale changes, run `yarn type-check` and `yarn build`.

## Mock-data structure and rules

Frontend mock data and mock API behavior belong under `src/mocks`. Current files mirror service areas: `auth.ts`, `dashboard.ts`, `occtl.ts`, `ocserv-groups.ts`, `ocserv-sync.ts`, `ocserv-users.ts`, and `system.ts`; `utils.ts` supplies `cloneMock`, and `index.ts` exports shared mock modules.

- Mocks must follow real generated or service request/response schemas and use `satisfies` where possible.
- Keep mock objects and scenario behavior out of Vue components.
- Components must consume the same API service functions in test and production modes.
- Keep mutable mock state inside the owning service/mock module and clone returned data so UI edits do not mutate fixtures accidentally.
- Reuse shared mock fixtures/factories instead of duplicating the same domain object in components.
- Mocks may cover loading outcomes, realistic populated data, configured/unconfigured values, successful mutations, backend errors, and empty states, but must not change production application architecture.

## Naming and import conventions

- Vue components and views: PascalCase (`DashboardLayout.vue`, `OcservSyncView.vue`).
- Composables: `useXxx.ts` or the existing kebab-case `use-theme.ts`.
- Stores, services, mocks, and helper modules: lowercase kebab-case (`ocserv-users.ts`, `auth-token.ts`, `group-config.ts`).
- Feature components use a shared PascalCase feature prefix.
- Route names and feature folder names use kebab-case.
- Constants use `UPPER_SNAKE_CASE`; functions and variables use camelCase; interfaces/types use PascalCase.
- Prefer the `@/` alias for `src` imports. It is configured in both `vite.config.ts` and `tsconfig.app.json`.
- Existing shadcn aliases are `@/components`, `@/components/ui`, `@/lib`, `@/lib/utils`, and `@/composables`.
- Relative imports are used for immediate SFC assets/styles where already established (for example, `./App.vue` and `./style.css` in `main.ts`); application modules otherwise use `@/`.

## Where to add feature-specific code

For an existing feature, add code next to its current equivalents:

- Route composition: `src/views/<Feature>View.vue`
- Dashboard navigation/route mapping: `src/router/dashboard-routes.ts` and `src/router/index.ts`
- Feature UI: `src/components/<feature>/`
- API operations and domain adapters: `src/api/services/<feature>.ts`
- Reusable feature request state: `src/composables/use<Feature>.ts`
- Cross-route state only: `src/stores/<feature>.ts`
- API-shaped test data/behavior: `src/mocks/<feature>.ts`
- User-facing text: the existing files in `src/locales`
- Generic reusable UI primitive: the matching existing directory in `src/components/ui/`
- Pure shared utility: `src/lib/`

Before creating a file, find the closest equivalent and place the new file beside it. If no equivalent exists, prefer extending the nearest established feature boundary over adding a new top-level directory.

## Verification commands

Run from `web/admin` as appropriate:

```sh
yarn type-check
yarn build
yarn build:test
yarn format
```

Use targeted formatting while working; avoid unrelated formatting changes.
