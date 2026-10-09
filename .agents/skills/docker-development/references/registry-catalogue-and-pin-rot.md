# Registry Catalogue Probing and Pin Rot

Two questions that look settled and are not: *does this registry publish image
X?* and *is a digest pin still the careful choice?*

## A 401 is a transport answer, not an absence

Authenticated registries refuse anonymous requests for **every** repository, so
an anonymous probe cannot distinguish "not published" from "not logged in":

```bash
curl -s -o /dev/null -w '%{http_code}\n' https://dhi.io/v2/mariadb/tags/list      # 401
curl -s -o /dev/null -w '%{http_code}\n' https://dhi.io/v2/phpmyadmin/tags/list   # 401
```

Both 401. Neither says anything about the catalogue. Ask the registry which
realm it wants, then authenticate against it — the credential is already on the
machine if `docker login` has run:

```bash
curl -sI https://dhi.io/v2/mariadb/tags/list | grep -i www-authenticate
# Bearer realm="https://dhi.io/token",service="registry.docker.io",scope="repository:mariadb:pull"

AUTH=$(jq -r '.auths["dhi.io"].auth' ~/.docker/config.json)      # never echo this
tok=$(curl -s -H "Authorization: Basic $AUTH" \
  "https://dhi.io/token?service=registry.docker.io&scope=repository:mariadb:pull" \
  | jq -r '.token')
curl -s -o /dev/null -w '%{http_code}\n' \
  -H "Authorization: Bearer $tok" https://dhi.io/v2/mariadb/tags/list            # 200
```

Now 200 versus 404 discriminates, and the probe has demonstrated it can return
both — which is what makes a 404 evidence. Run the positive control (an image
you know exists) in the same loop as the question you are actually asking; a
run that returns 404 for everything is measuring your credentials.

`.auths[…].auth` is base64 `user:token`, so treat the value as the secret it
is: pass it through a variable, never into displayed output. With a
`credsStore` configured there is no `auth` field — read the credential from the
helper instead.

## A digest pin can rot into the worse choice

Pinning a third-party image to a digest freezes the application *and* the base
underneath it. That is the point while upstream is publishing; it inverts the
moment upstream resumes rebuilding the same release. Measured on one image
during a single afternoon, counting only findings that have a fix:

| reference | with a fix | CRITICAL+HIGH |
|---|---:|---:|
| digest pinned ten months earlier | 1461 | 258 |
| the floating tag it was pinned from | 293 | 59 |

Same application version in both. The tag moved with the rebuild; the pin did
not. Before writing "pinned, therefore safe" — or the opposite, "upstream never
rebuilds, so there is no update path" — read the tag's timestamp:

```bash
curl -s "https://hub.docker.com/v2/repositories/library/<image>/tags?page_size=10&ordering=last_updated" \
  | jq -r '.results[] | "\(.name)\t\(.last_updated)\t\(.digest[0:19])"'
```

Equal digests across `:latest` and `:X.Y.Z` mean the release tag *is* latest, so
waiting for an upstream rebuild is not a plan. A moved `last_updated` means any
claim resting on staleness has expired and needs re-measuring, not repeating.

## Floating tags are correct for images you rebuild

The pinning rule is about trust, not about syntax: an image your own CI rebuilds
daily should float, because a pin pins the fix out too. Two failure modes when
that exemption is written down:

- **Scoping it to a registry host.** Ownership is what earns the exemption, not
  which host serves the bytes. An organisation publishing to both a private
  registry and `ghcr.io/<org>` needs both recognised — compared as path
  components, never as substrings, or `ghcr.io/someone/<org>-lookalike` and
  `evil.com/<your-registry>/x` qualify.
- **Testing the tag against the literal string `latest`.** `latest-rolling`,
  `latest-alpine` and friends float exactly as much and sail through. Treat a
  `latest-*` or `edge-*` prefix as floating.

## A tag that "should" exist can just not

Don't assume a variant tag exists because the pattern is common elsewhere —
check the catalogue the same way as above. `composer:2-alpine` is not a real
tag: the official `composer` image is Alpine-based *at* `composer:2`, so the
`-alpine` suffix some other images use has nothing to pull. A short,
unqualified reference like `composer:2` in a `.env`/build-arg also fails
differently depending on where it resolves: buildah/podman running
non-interactively enforce short-name resolution and refuse to guess a
registry, erroring with "short-name resolution enforced but cannot prompt
without a TTY" instead of defaulting to Docker Hub. Fully qualify it
(`docker.io/library/composer:2`) rather than relying on the short name to
resolve the same way it does interactively.

## A push failing "unauthorized" is a role question, not always a credential one

`unauthorized to access repository: <repo>, action: push` from a private
registry (Harbor and similar) usually reads like a login problem, so the
instinct is to re-check the password and rebuild. Ask the registry's own RBAC
API first instead of spending a full build+push cycle per guess — Harbor
answers this in two read-only calls:

```bash
curl -s -u "$USER:$TOKEN" https://harbor.example.com/api/v2.0/users/current \
  | jq '.username'
curl -s -u "$USER:$TOKEN" "https://harbor.example.com/api/v2.0/projects/<id>/members" \
  | jq '.[] | select(.entity_name=="'"$USER"'")'
```

A valid login with no membership entry (or `role_id: null`) for the target
project means the account has zero role there — no amount of retrying the
push fixes that; someone with project-admin/Maintainer needs to grant a role
(Developer is enough to push).
