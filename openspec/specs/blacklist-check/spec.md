## ADDED Requirements

### Requirement: Blacklist blocks packages by name
The proxy SHALL block any package whose name matches a blacklist entry that has no version constraint.

#### Scenario: Blacklisted package by name only is blocked
- **WHEN** a request is made for package `malicious-pkg`
- **AND** the blacklist contains `{name: malicious-pkg}` with no version
- **THEN** the engine returns BLOCK with reason "blacklist"

### Requirement: Blacklist blocks packages by name and version
The proxy SHALL block any package whose name and version both match a blacklist entry with a version constraint.

#### Scenario: Blacklisted package with matching version is blocked
- **WHEN** a request is made for package `bad-lib` version `2.0.0`
- **AND** the blacklist contains `{name: bad-lib, version: "2.0.0"}`
- **THEN** the engine returns BLOCK with reason "blacklist"

#### Scenario: Blacklisted package with different version passes
- **WHEN** a request is made for package `bad-lib` version `1.5.0`
- **AND** the blacklist contains `{name: bad-lib, version: "2.0.0"}`
- **THEN** the engine returns PASS

### Requirement: Blacklist entries with invalid syntax cause startup failure
The proxy SHALL refuse to start if any blacklist entry fails validation (e.g., empty name).

#### Scenario: Invalid blacklist entry prevents startup
- **WHEN** config.yaml contains a blacklist entry with empty name
- **THEN** the program exits with a non-zero status code
- **AND** an error message is logged
