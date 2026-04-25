local p = import 'pkg/main.libsonnet';

p.pkg({
  source: 'https://github.com/marcbran/jsonnet-plugin-http',
  repo: 'https://github.com/marcbran/jsonnet.git',
  branch: 'plugin-http',
  path: 'plugin/http',
  target: 'http',
}, |||
  Read-only HTTP GET requests against a REST API. Base URL and default headers are configured when the plugin is started or embedded in Go.

  Generated operation functions take `args` with optional `query` and `headers` objects (Http `in: query` / `in: header`); path parameters are separate function arguments on the nested path API.
|||, {
  request: p.desc(|||
    Sends a GET request. `input` is an object with `method` (`GET` only), `path`, optional `headers`, and optional `query` (query string map).

    On success returns parsed JSON. On failure returns a `Status` object (`kind: "Status"`).
  |||),
})
