local p = import 'pkg/main.libsonnet';

p.ex({
  request: p.ex([{
    name: 'get with query params',
    inputs: [{
      method: 'GET',
      path: '/api/v1/query',
      query: { query: 'up' },
    }],
  }, {
    name: 'get with extra headers',
    inputs: [{
      method: 'GET',
      path: '/v1/status',
      headers: { Accept: 'application/json' },
    }],
  }]),
})
