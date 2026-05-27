## ADDED Requirements

### Requirement: Direct mode writes to SQLite
The system SHALL insert or replace records in the `malicious_packages` table using `INSERT OR REPLACE` semantics when in direct mode.

#### Scenario: Direct insert
- **WHEN** output mode is "direct"
- **AND** a package is parsed from an OSV file
- **THEN** a row is inserted or replaced in the SQLite database

### Requirement: Delete removed packages in direct mode
The system SHALL delete rows from `malicious_packages` when the corresponding OSV files have been removed from the repository (detected via git diff status "D").

#### Scenario: Direct delete of removed package
- **WHEN** output mode is "direct"
- **AND** git diff shows a file was deleted
- **THEN** the corresponding rows are removed from the database

### Requirement: Transactional batch updates
The system SHALL wrap batch updates in a SQL transaction for atomicity.

#### Scenario: Batch commit
- **WHEN** 100 packages are being inserted
- **THEN** all inserts are committed atomically within a single transaction
