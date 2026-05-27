## MODIFIED Requirements

### Requirement: Checks execute in configured order and short-circuit
The engine SHALL execute checks in order (malicious-db, cooldown, typo, OSV) and SHALL stop at the first check that returns BLOCK, skipping all remaining checks.

#### Scenario: Blacklist block stops pipeline
- **WHEN** the blacklist check returns BLOCK
- **THEN** cooldown, typo, and OSV checks are NOT executed
- **AND** the engine returns BLOCK with reason "blacklist"

#### Scenario: All checks pass returns pass
- **WHEN** all enabled checks return PASS
- **THEN** the engine returns PASS with no reason

### Requirement: Streaming check is phase two
The script-scan check SHALL NOT participate in the phase-one sequential pipeline. The proxy SHALL invoke it via StreamChecker interface after all phase-one checks pass and during the download stream forwarding.

#### Scenario: Script scan runs after metadata checks pass
- **WHEN** all metadata checks return PASS
- **AND** the proxy begins forwarding the download stream
- **THEN** the StreamChecker is invoked on the cloned stream
