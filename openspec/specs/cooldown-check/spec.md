## ADDED Requirements

### Requirement: Cooldown check is disabled by default
The cooldown check SHALL be inactive when `cooldown.enabled` is `false` in the configuration.

#### Scenario: Cooldown disabled skips check
- **WHEN** `cooldown.enabled` is `false`
- **AND** a request is made for any package
- **THEN** the cooldown check is skipped
- **AND** the engine continues to the next check

### Requirement: Cooldown blocks packages published within the configured duration
When enabled, the cooldown check SHALL fetch the package's publish time from the upstream registry metadata and block the request if the package was published less than the configured duration ago.

#### Scenario: Recently published package is blocked
- **WHEN** `cooldown.enabled` is `true`
- **AND** `cooldown.duration` is `48h`
- **AND** the package was published 2 hours ago
- **THEN** the engine returns BLOCK with reason "cooldown"

#### Scenario: Old package passes cooldown
- **WHEN** `cooldown.enabled` is `true`
- **AND** `cooldown.duration` is `48h`
- **AND** the package was published 72 hours ago
- **THEN** the engine returns PASS

### Requirement: Cooldown fails closed on metadata parse error
The cooldown check SHALL return BLOCK when the upstream metadata JSON cannot be parsed or does not contain a valid publish timestamp.

#### Scenario: Corrupt metadata response blocks
- **WHEN** `cooldown.enabled` is `true`
- **AND** the upstream metadata endpoint returns invalid JSON
- **THEN** the engine returns BLOCK with reason "cooldown"

### Requirement: Cooldown fails closed on metadata fetch timeout
The cooldown check SHALL return BLOCK when the metadata fetch times out (2 second timeout).

#### Scenario: Metadata fetch timeout blocks
- **WHEN** `cooldown.enabled` is `true`
- **AND** the metadata HTTP request exceeds 2 seconds
- **THEN** the engine returns BLOCK with reason "cooldown"
