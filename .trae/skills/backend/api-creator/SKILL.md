---
name: api-creator
description: Create Echo v5 APIs using Service/Repository/Usecase pattern. 
---

# Context
**CRITICAL**: You MUST read and strictly follow master-rules before starting any task. Every decision (naming, structure, errors, pagination) must align with these rules.
**CRITICAL**: Only works on dashboard/api directory.

# Strategy
- Service: `internal/service/{name}/` (controller.go, routes.go, types.go).
- Validation: Use `request.Validator` and `request.BadRequest`.
- Logic: Use `ResponseWithCode` for UNIQUE keys from `config/errors.json` ONLY—NEVER reuse existing error codes; create a new unique code and add it to errors.json first before using it in code.
- Documentation: Add Swagger comments and run `swag init --pd`.
- **Role/Ownership**: Always pass admin ID and role from controller → use case → repository layer; enforce ownership checks in repository layer for non-super admins.

# Workflow
1. Check `config/errors.json` and find the next available unique error code for any new errors your API will introduce.
2. Add the new unique error code(s) to `config/errors.json` with appropriate fa, en, it, ru, zh-cn, and zh-tw messages and correct status code. Ensure all language fields are populated for every new error code.
3. Define request structs in `types.go`. DO NOT create response structs for responses that just return a model - use the model directly.
4. Implement controller logic with validation:
   - Retrieve `id` (admin ID) and `role` from Echo context (`c.Get("id").(uint)` and `c.Get("role").(string)`)
   - Pass these values to the use case layer
5. Implement use case and repository layers with role and ownership checks:
   - Repository methods must accept `adminID` and `role` parameters
   - For non-super admins, always add `WHERE owner_admin_id = ?` condition to queries
6. Register each service routes in `internal/provider/routing/routing.go`.
7. Update Swagger docs.

# Note:
- Note:
- instead of using `// @Security ApiKeyAuth` in swagger use:
```text
// @Param        Authorization header string true "Bearer TOKEN"
// @Failure 401 {object} request.ErrorResponse
```
- for pagination use pkg/request/pagination module 
- generate in `api/` directory with: `swag init --pd`
- serve each documentation on its respective service route using `http-swagger` ONLY when `Debug` is true. 
- DONT create or define api with PUT method.
- **CRITICAL**: EVERY error must have a UNIQUE code in config/errors.json; NO REUSE of codes for different error scenarios.
- **CRITICAL**: Swagger @Router comment MUST NOT include /api/v1/ prefix. For example, use // @Router /auth/login [post], NOT /api/v1/auth/login.
- **CRITICAL**: After creating or updating any API handlers/routes, ALWAYS run `swag init --pd` from within dashboard/api directory to regenerate Swagger docs.