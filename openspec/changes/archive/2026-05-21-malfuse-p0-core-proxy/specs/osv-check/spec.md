## ADDED Requirements

### Requirement: OSV check queries the OSV API for vulnerabilities
The OSV check SHALL query `GET /v1/query` with the package name and ecosystem to determine if the package has known vulnerabilities.

#### Scenario: Package with known vulnerabilities is blocked
- **WHEN** `osv.enabled` is `true`
- **AND** the OSV API returns at least one vulnerability for the package
- **THEN** the engine returns BLOCK with reason "osv"

#### Scenario: Package with no vulnerabilities passes
- **WHEN** `osv.enabled` is `true`
- **AND** the OSV API returns zero vulnerabilities for the package
- **THEN** the engine returns PASS

### Requirement: OSV results are cached with TTL
The OSV check SHALL cache query results in memory for the configured TTL duration. Subsequent requests for the same package within the TTL SHALL use the cached result.

#### Scenario: Cache hit returns saved result
- **WHEN** a previous OSV query for `requests` returned BLOCK
- **AND** the TTL has not expired
- **AND** a new request arrives for `requests`
- **THEN** the cached result is returned without a network call

#### Scenario: Cache miss triggers new query
- **WHEN** no cached result exists for `requests`
- **OR** the TTL has expired for the cached result
- **THEN** a new OSV API query is made
- **AND** the result is cached with the current time

### Requirement: OSV check fails open on API error
The OSV check SHALL return PASS when the OSV API is unreachable, times out (2 second timeout), or returns a non-200 status code.

#### Scenario: OSV API timeout returns pass
- **WHEN** `osv.enabled` is `true`
- **AND** the OSV API request exceeds 2 seconds
- **THEN** the engine returns PASS
- **AND** a warning is logged

#### Scenario: OSV API returns non-200
- **WHEN** `osv.enabled` is `true`
- **AND** the OSV API returns HTTP 500
- **THEN** the engine returns PASS
- **AND** a warning is logged

### Requirement: OSV cache is per-ecosystem
The OSV cache key SHALL include both the package name and ecosystem, so the same package name in different ecosystems (e.g., `requests` in both PyPI and npm) is cached separately.

#### Scenario: PyPI and NPM caches are independent
- **WHEN** an OSV query for `lodash` with ecosystem `npm` returns BLOCK
- **AND** a subsequent OSV query for `lodash` with ecosystem `pypi` arrives
- **THEN** the npm result is NOT reused for the pypi query
- **AND** a new API query is made for `lodash` in the `pypi` ecosystem
