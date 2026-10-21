local p = import 'pkg/main.libsonnet';

p.pkg({
  source: 'https://github.com/marcbran/jsonnet-plugin-http',
  repo: 'https://github.com/marcbran/jsonnet.git',
  branch: 'plugin/http',
  path: 'plugin/http',
  target: 'http',
}, |||
  HTTP requests against a REST API. Base URL and default headers are configured when the plugin is started or embedded in Go.

  Generated operation functions take `args` with optional `query` and `headers` objects (Http `in: query` / `in: header`); path parameters are separate function arguments on the nested path API.

  Mutating requests are exposed as a plugin action rather than a jsonnet function, so they are not callable during evaluation.
  The host invokes the `request` action with `{ method, path, baseURL, headers, query, context, body }`, where `method` is
  `POST`, `PUT`, `PATCH`, or `DELETE` and all fields but `method` and `path` are optional. It returns `{ status, headers, body }`
  as JSON and fails on a non-2xx response.
|||, {
  request: p.desc(|||
    Sends a request. `input` is an object with `method`, `path`, optional `headers`, and optional `query` (query string map).
    `method` is `GET`, or `POST` with `readonly: true` and a `body` for read-only queries such as search or GraphQL.

    On success returns parsed JSON. On failure returns a `Status` object (`kind: "Status"`).
  |||),
})
