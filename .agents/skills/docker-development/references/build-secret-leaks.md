# Where a build credential actually leaks

`docker history` is the check everyone runs, and for a multi-stage build it is
the wrong one. A credential passed as `ARG` into a builder stage never reaches
the shipped layers — that stage is not in the final image — so the history is
clean while the credential is public anyway.

buildx attaches an SLSA provenance attestation to every pushed image and records
the build arguments in it **verbatim**:

```bash
docker buildx imagetools inspect <ref> --format '{{json .Provenance}}' \
  | jq -r '.[].SLSA.buildDefinition.externalParameters.request.args | keys[]'
# build-arg:COMPOSER_AUTH   <- the value sits right there, in cleartext
```

Measured case (2026-08-12): a public GHCR package carried a working GitLab
`glpat-` token in the attestation of **1461** published versions. The `ARG` was
in a `composer-builder` stage that is never shipped.

## Verify it, with a check that can fail

Grep for the **credential pattern**, never for the argument name. The
attestation embeds the push's commit messages and the `RUN` command lines, so
the name matches your own commit text and reports a leak that is not there —
that false positive cost a round of "still broken" in the session that found
this.

```bash
for ref in "$OLD_TAG" "$NEW_TAG"; do
  n=$(docker buildx imagetools inspect "$ref" --format '{{json .Provenance}}' \
      | grep -oE 'glpat-[A-Za-z0-9_-]{10,}|ghp_[A-Za-z0-9]{20,}' | sort -u | wc -l)
  echo "$ref: $n credential values"
done
```

A pre-fix tag returning `1` and a post-fix tag returning `0` is the proof. A
single post-fix `0` on its own proves nothing about the check.

## The fix

```dockerfile
RUN --mount=type=secret,id=composer_auth \
    COMPOSER_AUTH="$(cat /run/secrets/composer_auth 2>/dev/null || true)" \
    composer install --no-dev
```

```hcl
target "app" {
  secret = ["type=env,id=composer_auth,env=COMPOSER_AUTH"]
}
```

`type=env` reads the variable the CI already exports, so the calling workflow
usually needs no change at all. Tolerate a missing secret (`|| true`): forks and
local builds have no credential and should still resolve public dependencies
rather than fail.

## After a leak

Rotation is the only remedy that acts on what is already published — the
attestations of existing versions keep their copy of the value forever, or until
someone deletes those package versions. Rotate first, fix the build second.

## The other leak: your own debugging session

A leak does not need buildx provenance to happen. `docker build --progress=plain`
(or any verbose/raw build log) echoes the literal `RUN` command line with every
`ARG`/`ENV` already interpolated — so a credential-bearing build arg shows up in
cleartext the moment that output reaches a terminal, a piped log file, or a
`tail`. This is a transient leak (nothing gets published), but it lands directly
in whatever you're capturing the session with — chat transcript, screen
recording, CI job log — which is exactly the audience `--mount=type=secret` was
meant to keep it from. It happens even when the project's real Dockerfile is
already fixed with `--mount=type=secret`, because the leak comes from an ad-hoc
debug invocation outside that path, not from the build definition.

Treat it the same as a provenance leak once it happens: rotate the credential
first, then clean up (delete the log file, prune the build cache — the layer
with the interpolated value may still be cached locally even though it was
never pushed). Avoid it by not running debug builds with secret-bearing `ARG`s
through `--progress=plain`/verbose output that you then cat, tail, or pipe
into something you'll read — redact the known secret value first if you must.
