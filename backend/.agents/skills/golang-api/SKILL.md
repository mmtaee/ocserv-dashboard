---
name: golang-api
description: Build, modify, review, and test Echo v5 HTTP APIs in this repository, including controllers, routes, request validation, Swagger annotations, pagination, middleware, and application error responses.
---

# Go API

Use this skill with `golang` for HTTP endpoints, routing, controllers, request DTOs, response mapping, middleware, or Swagger.

Keep controllers limited to HTTP concerns and preserve the project's controller-usecase-repository boundaries. Use Echo v5 and existing request, response, routing, pagination, and error helpers.

Never define a new endpoint with `PUT`; use the method appropriate to the operation. Do not silently redesign existing `PUT` endpoints unless requested.

## Detailed Guidance

Read [references/api-architecture.md](references/api-architecture.md) before changing an API. It contains controller responsibilities, validation and error rules, Swagger conventions, routing locations, middleware requirements, and the endpoint checklist.
