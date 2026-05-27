## ADDED Requirements

### Requirement: Checks execute in configured order and short-circuit
The engine SHALL execute checks in order (blacklist, cooldown, typo, OSV) and SHALL stop at the first check that returns BLOCK, skipping all remaining checks.

#### Scenario: Blacklist block stops pipeline
- **WHEN** the blacklist check returns BLOCK
- **THEN** cooldown, typo, and OSV checks are NOT executed
- **AND** the engine returns BLOCK with reason "blacklist"

#### Scenario: All checks pass returns pass
- **WHEN** all enabled checks return PASS
- **THEN** the engine returns PASS with no reason

### Requirement: Disabled checks are skipped
The engine SHALL skip any check whose `enabled` field is `false` in the configuration.

#### Scenario: Cooldown disabled is skipped
- **WHEN** `cooldown.enabled` is `false`
- **AND** a request is received
- **THEN** the blacklist check runs
- **AND** the cooldown check is skipped
- **AND** the typo check runs

### Requirement: Engine accepts context for cancellation
The `Check` method SHALL accept a `context.Context` and SHALL abort if the context is cancelled or times out, returning BLOCK.

#### Scenario: Context timeout forces block
- **WHEN** the context has a deadline that expires during a check
- **THEN** the engine returns BLOCK
- **AND** the reason indicates timeout

### Requirement: Engine returns the first blocking reason
When multiple checks could potentially block, the engine SHALL return the reason from the first check that fired.

#### Scenario: Blacklist fires before typo
- **WHEN** a package matches both blacklist and would match typo-squatting
- **AND** blacklist runs first
- **THEN** the engine returns BLOCK with reason "blacklist"
- **AND** typo-squatting is never executed
