## ADDED Requirements

### Requirement: Query malicious packages by name and ecosystem
The malFuse proxy SHALL query the SQLite database for each incoming package request using `SELECT COUNT(*) FROM malicious_packages WHERE name=? AND ecosystem=? AND (version=? OR version IS NULL)` to determine if the package is known malicious.

#### Scenario: Package found with version match
- **WHEN** a request arrives for `bad-lib` version `2.0.0` ecosystem `pypi`
- **AND** the database contains `bad-lib` with version `2.0.0`
- **THEN** the engine returns BLOCK with reason "malicious-db"

#### Scenario: Package found with no version constraint
- **WHEN** a request arrives for `evil-pkg` version `1.0` ecosystem `pypi`
- **AND** the database contains `evil-pkg` with version NULL
- **THEN** the engine returns BLOCK with reason "malicious-db"

#### Scenario: Package not found passes
- **WHEN** a request arrives for `safe-pkg` ecosystem `pypi`
- **AND** the database has no entry for `safe-pkg`
- **THEN** the engine returns PASS

### Requirement: Graceful fallback when database is missing
The malFuse proxy SHALL log a warning and skip the malicious-db check if the SQLite database file does not exist or cannot be opened.

#### Scenario: Missing database file
- **WHEN** `malfuse.db` does not exist
- **THEN** a warning is logged
- **AND** the malicious-db check returns PASS
- **AND** the proxy continues processing with remaining checks

### Requirement: Database opened read-only
The malFuse proxy SHALL open the SQLite database in read-only mode.

#### Scenario: Read-only connection
- **WHEN** the proxy opens `malfuse.db`
- **THEN** the connection is opened with `mode=ro`
