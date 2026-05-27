## ADDED Requirements

### Requirement: Proxy routes requests by URL prefix
The proxy SHALL match incoming request paths against configured route prefixes and forward matched requests to the corresponding upstream registry over HTTPS.

#### Scenario: PyPI request is routed to pypi.org
- **WHEN** a GET request is received at `/pypi/simple/requests/`
- **THEN** the `/pypi/` prefix is stripped from the path
- **AND** the request is forwarded to `https://pypi.org/simple/requests/`

#### Scenario: NPM request is routed to registry.npmjs.org
- **WHEN** a GET request is received at `/npm/left-pad`
- **THEN** the `/npm/` prefix is stripped from the path
- **AND** the request is forwarded to `https://registry.npmjs.org/left-pad`

### Requirement: Proxy forwards upstream response unmodified on pass
The proxy SHALL stream the upstream response body and headers back to the client without modification when the engine returns PASS.

#### Scenario: Successful upstream response is passed through
- **WHEN** the engine returns PASS for a request
- **AND** upstream returns 200 with a tarball body
- **THEN** the client receives the full body unmodified
- **AND** upstream response headers are copied to the client response

### Requirement: Proxy returns 403 on blocked request
The proxy SHALL return HTTP 403 Forbidden when the engine returns BLOCK, and SHALL NOT forward the request to the upstream registry.

#### Scenario: Blocked request receives 403
- **WHEN** the engine returns BLOCK with reason "blacklist"
- **THEN** the proxy returns HTTP 403
- **AND** the response body contains the reason
- **AND** no request is made to the upstream registry

### Requirement: Proxy returns 502 on unmatched route
The proxy SHALL return HTTP 502 Bad Gateway when the request path does not match any configured route prefix.

#### Scenario: Unknown path returns 502
- **WHEN** a GET request is received at `/unknown/something`
- **AND** no route has prefix `/unknown/`
- **THEN** the proxy returns HTTP 502

### Requirement: Proxy returns 502 on upstream failure
The proxy SHALL return HTTP 502 Bad Gateway when the upstream registry is unreachable or returns a connection error.

#### Scenario: Upstream connection refused
- **WHEN** the engine returns PASS
- **AND** the upstream registry connection fails
- **THEN** the proxy returns HTTP 502

### Requirement: Proxy passes through upstream HTTP errors
The proxy SHALL pass through upstream 4xx and 5xx responses to the client without modification.

#### Scenario: Upstream returns 404
- **WHEN** the engine returns PASS
- **AND** upstream returns HTTP 404
- **THEN** the proxy returns HTTP 404 with the upstream response body
