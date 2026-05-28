## ADDED Requirements

### Requirement: Base64 encoded payload detection
The system SHALL detect base64-encoded strings of length greater than or equal to the configured minimum using the pattern `[A-Za-z0-9+/=]{N,}`.

#### Scenario: Long base64 string triggers detection
- **WHEN** scanning a file containing at least one base64 string of 100+ characters
- **THEN** the obfuscation detector returns BLOCK with reason "obfuscation"

#### Scenario: Short base64 strings do not trigger
- **WHEN** scanning a file containing only base64 strings shorter than the configured minimum
- **THEN** the obfuscation detector returns PASS

### Requirement: Hex encoded payload detection
The system SHALL detect hex-encoded escape sequences (`\\xNN`) appearing consecutively more than the configured minimum count.

#### Scenario: Many hex escapes trigger detection
- **WHEN** scanning a file containing 30+ consecutive `\\xNN` hex escape sequences
- **AND** `hex_min_length` is 20
- **THEN** the obfuscation detector returns BLOCK with reason "obfuscation"

### Requirement: Eval/exec call chain detection
The system SHALL detect use of `eval(`, `exec(`, `Function(`, and `__import__` with large string arguments.

#### Scenario: eval with encoded string triggers detection
- **WHEN** scanning a file containing `eval(atob("..."))` or `eval(base64.b64decode("..."))`
- **THEN** the obfuscation detector returns BLOCK with reason "obfuscation"
