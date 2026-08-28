---
name: vue-i18n
description: Maintain localization and bidirectional layout support in this Vue admin dashboard. Use when adding or changing component text, locale catalogs, language selection, document direction, or i18n environment configuration.
---

# Vue i18n

Keep user-facing text in `src/locales`, including placeholders, titles, and screen-reader labels. Components must use the global Composition API composer rather than hardcoded display strings. Do not translate generated API identifiers or unstable backend error payloads without a stable error-code contract.

Maintain identical message keys across all locale files. The languages exposed by the UI come from `VITE_I18N_LANGUAGES`; only codes with bundled catalogs are accepted. Change languages through `setLocale` so the selection is persisted and the document `lang`, `dir`, and title stay synchronized.

Treat Arabic and Persian as RTL. Preserve the reactive Reka `ConfigProvider`, use logical Tailwind positioning utilities for direction-agnostic layout, and place the dashboard sidebar on the matching physical side. When adding shadcn-vue components, keep `rtl: true` in `components.json` and migrate any older physical positioning classes.

After changing messages or localized UI, verify catalog key parity, run `yarn type-check`, and run `yarn build`.
