## REMOVED Requirements

### Requirement: Blacklist blocks packages by name
**Reason**: Replaced by SQLite-backed `malicious-db` check that queries the `malicious_packages` table.
**Migration**: Run `malfuse-db update` to populate the SQLite database with OSV data. The proxy will use the new `malicious-db` check instead.

### Requirement: Blacklist blocks packages by name and version
**Reason**: Replaced by SQLite-backed `malicious-db` check with version-aware query.
**Migration**: Same as above.

### Requirement: Blacklist entries with invalid syntax cause startup failure
**Reason**: Blacklist entries in config.json are no longer parsed. Malformed OSV JSON is handled at ingest time, not proxy startup.
**Migration**: Config validation for blacklist is removed. OSV parse errors are logged during ingestion.

## MODIFIED Requirements

### Requirement: Engine pipeline includes malicious-db check
The engine SHALL include the `malicious-db` check as the first check in the pipeline (before cooldown, typo, and OSV checks). This check SHALL query the SQLite database as described in the `malicious-db-check` spec.

#### Scenario: Malicious-db check runs first
- **WHEN** a request is processed by the engine
- **THEN** the malicious-db check is executed before all other checks
- **AND** if it returns BLOCK, remaining checks are skipped

#### Scenario: Malicious-db check disabled when db unavailable
- **WHEN** the SQLite database file is missing
- **THEN** the malicious-db check returns PASS
- **AND** the check pipeline continues to the next check
