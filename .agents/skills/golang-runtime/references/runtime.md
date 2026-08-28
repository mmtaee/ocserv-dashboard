# Runtime Rules

## Graceful Shutdown

Servers must support graceful shutdown. Startup and bootstrap code should start Echo, listen for termination signals, stop accepting new requests, shut down using a bounded context, release resources where necessary, and surface unexpected shutdown errors.

Do not use abrupt process termination as the normal shutdown path. Avoid `os.Exit` or `log.Fatal` in reusable packages.

## Cobra Commands

CLI commands belong under `cmd/` and use Cobra. Keep command construction separate from business logic. Delegate substantial work to application or bootstrap packages, and normally return errors instead of terminating deep in application code.

## Dependency Construction

Construct dependencies explicitly in this direction:

```text
Database -> Repository -> Usecase -> Controller -> Router
```

Do not create hidden package-level mutable dependencies. Avoid global state unless the existing architecture explicitly requires it.

## Interfaces

Define interfaces when they provide a real architectural boundary, make testing possible, or match existing structure. Typical boundaries are `Usecase -> Repository` and `Controller -> Usecase`. Keep interfaces small and do not create one merely because a struct exists.

## Routing Providers

Service routes belong in `internal/service/{service-name}/router.go`. Global registration belongs in the established routing provider, normally `internal/provider/routing/routes.go` or `pkg/routing/router.go`. Do not register the same route in multiple places.
