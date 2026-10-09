# nginx FastCGI keepalive starves php-fpm

A php-fpm child stays bound to its FastCGI connection for as long as that
connection lives. An nginx keepalive pool to php-fpm therefore does not park
idle sockets — it parks **workers**. Size the pool at or above
`pm.max_children` and it can pin every child, leaving arriving requests to wait
for one to come free.

```nginx
upstream php-fpm {
    server 127.0.0.1:9000;
    keepalive 16;          # against pm.max_children = 10
}
location ~ \.php$ {
    fastcgi_keep_conn on;  # <- makes the pool real
}
```

## The signature — all of it at once, or it is something else

- container at **0 % CPU**, memory flat — nothing is working
- php-fpm **slowlog empty** at a low threshold
  (`request_slowlog_timeout = 5s`): the scripts were never slow, they never got
  a worker
- the stalled requests answer **HTTP 200 after ~60 s**, with TTFB equal to the
  total, and the edge proxy's access log reports the same duration, so the wait
  is behind it
- the very same URLs replayed **one at a time** in the same session take
  ~100 ms

The empty slowlog is the decisive one: it separates "slow code" from "never
scheduled", and it is the measurement most likely to be skipped.

## Why it hides from local testing

Only a client that can issue everything at once builds the queue. Over HTTP/2
through a reverse proxy the page's requests share one connection and arrive
together; a direct HTTP/1.1 client opens at most six connections per origin and
never triggers it. Reproduce **with the proxy in the path** — a stack tested by
addressing the application container directly measures a load shape that
production never sees.

## Measured

TYPO3 backend behind Caddy, `pm.max_children = 10`, opening the backend shell:

| nginx | slowest request | over 5 s |
|---|---|---|
| `keepalive 16` + `fastcgi_keep_conn on` | 61005 ms | 5–8 of 24 |
| `keepalive 8` + `fastcgi_keep_conn on` | 19614 ms | 1 of 24 |
| no pool, `fastcgi_keep_conn off` | **107 ms** | none |

The intermediate value still costs 19.6 s, so this is not a tuning question:
remove the pool. `fastcgi_keep_conn off` is the nginx default, and connection
setup to `127.0.0.1` is not worth holding a worker for.

Raising `pm.max_children` instead trades memory for the same failure one load
step further out — and where the container has a memory cap, the headroom for
`keepalive`-many concurrently pinned children is not there to give.
