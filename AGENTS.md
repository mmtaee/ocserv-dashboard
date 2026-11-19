# Ocserv Dashboard agent guide

- Preserve the controller -> usecase -> repository boundaries in `backend/` and pass request contexts through the flow.
- Use Echo v5, existing request/error helpers, existing services, and the generated Swagger/client workflow. New action endpoints use `POST`, not `PUT`.
- Keep Vue work in the existing Composition API/TypeScript structure. API calls belong in `web/admin/src/api/services`, shared bootstrap state in Pinia, and all visible text in every supported locale.
- Read the applicable instructions in `.agents/skills`, `backend/.agents/skills`, and `web/admin/.agents` before editing those areas.
- Prefer the Codebase Memory graph for structural discovery, then verify operated-on paths with `check_index_coverage`; use source search for scripts, configuration, literals, and reported coverage gaps.
- Preserve user changes, avoid unrelated refactors and dependencies, format changed Go/Vue/shell files, and run focused tests plus the relevant backend/frontend build checks.
- Production runtime files live at `/opt/ocserv-dashboard`, with the Git checkout in `/opt/ocserv-dashboard/source`; never log or commit `.env`, credentials, certificates, database data, or update logs.
