# GPG Signature Verification in Image Builds

Patterns for verifying downloaded release tarballs against GPG keys inside
multi-stage builds — and the layer pitfall that breaks the naive approach.
Distilled from building a central release-key image for PHP/nginx/Node builds.

## Pitfall: gpg import bakes a stale keybox lock into the layer

With gnupg 2.4, `gpg --import` in one `RUN` starts keyboxd/gpg-agent and leaves
`public-keys.d/*.lock` behind in the GnuPG home **inside the committed layer**
(`/root/.gnupg` when building as root, else `$GNUPGHOME`/`~/.gnupg`). The
next `RUN` that touches the keyring (e.g. `gpg --verify`) then hangs on the
stale lock and dies:

```text
gpg: Note: database_open ... waiting for lock (held by 9) ...
gpg: keydb_search failed: Operation timed out
gpg: Can't check signature: No public key
```

If you must import, clean up in the **same** `RUN` that imported:

```dockerfile
RUN gpg --no-tty --batch --import /tmp/keys.asc \
 && gpgconf --kill all \
 && rm -f "$(gpgconf --list-dirs homedir)"/public-keys.d/*.lock
```

## Prefer gpgv: verify without any keyring state

`gpgv` reads **binary** keyring files directly — no import, no `~/.gnupg`,
no agent, no locks, and the accepted signer set is exactly the key files you
pass:

```dockerfile
COPY --from=release-keys /keys/php/8.3/ /tmp/gpg-keys/
RUN set -eux; \
    for k in /tmp/gpg-keys/*.gpg; do \
        [ -s "$k" ] || exit 1; \
        set -- "$@" --keyring "$k"; \
    done; \
    gpgv "$@" php.tar.xz.asc php.tar.xz
```

Convert armored keys once with `gpg --dearmor` (or ship binary exports);
`--keyring` may be repeated per key file.

### `gpgv` is not always in the `gnupg` package

Install `gpgv` explicitly for the base you build on:

- **Debian/Ubuntu**: `gpgv` is a **separate package** — `apt-get install gnupg`
  does NOT provide it. A job that verifies with `gpgv` must
  `apt-get install ... gpgv`, or every call dies with a silent
  `gpgv: command not found` (which inside `if gpgv ...; then` reads as a
  verification *failure*, not a missing binary).
- **Alpine**: the `gnupg` package *does* bundle `gpgv`, so `apk add gnupg` is enough.

Testing on your host (which usually has `gpgv`) hides the Debian gap — run the
verification inside the actual build image before trusting it.

## A signature proves the signer, not the version

`gpgv` (and `gpg --verify`) proves *"signed by a trusted release key"* — **not**
*"is the version I asked for"*. The signature travels with the tarball, so a
**validly signed older release** placed under a newer name passes verification
unchanged. This is harmless when you download straight from the authoritative
origin (it owns the path→content binding), but the moment a **mutable store**
sits in front of the origin — a package-registry cache, a mirror, an artifact
proxy — an attacker who can write that store can serve a real, signed,
*known-vulnerable* `foo-1.2.3.tar.gz` under the `1.2.9` coordinate. It passes
`gpgv`, and the build silently compiles the downgraded version.

Bind the artifact to the requested version after the signature checks out —
assert the tarball's sole top-level directory (or the checksum-file entry)
matches the expected `name-<version>`:

```dockerfile
RUN set -eux; \
    gpgv --keyring /tmp/k.gpg php.tar.xz.asc php.tar.xz; \
    got=$(tar tf php.tar.xz | sed 's#/.*##' | sort -u); \
    [ "$got" = "php-${PHP_VERSION}" ] || { echo "version mismatch: $got" >&2; exit 1; }
```

Checksum-list formats (e.g. Node's `SHASUMS256.txt`) get this for free — the
artifact is looked up by exact filename inside the signed list, so a wrong
version yields no matching line. Detached-signature formats (php/nginx) do not —
add the assertion yourself. It also hardens any later `tar --strip-components=1`
that assumes a single predictable top-level directory.

## Ship keys as a scratch image, not from keyservers

Public keyservers (`keyserver.ubuntu.com`, `keys.openpgp.org`) time out under
parallel CI fan-out and break scheduled builds. Keep reviewed public keys in a
minimal image and consume them via `COPY --from`:

```dockerfile
FROM registry.example.com/support/gpg-keys:latest@sha256:<digest> AS release-keys
```

- Final stage `FROM scratch`: nothing to patch or trust beyond the key files.
- Tag every published digest immutably (e.g. commit SHA alongside `latest`) so
  digest pins never reference an untagged manifest.
- Key *material* belongs to authoritative origins (e.g.
  `php.net/distributions/php-keyring.gpg`, `nginx.org/keys/*.key`), never to a
  keyserver fetch at build time.

A robust shape: a central keys image that also ships `verify-release` /
`get-verified-release` helper scripts *in* the image (POSIX sh, executed by the
consumer's shell; a `scratch` image runs nothing itself).
