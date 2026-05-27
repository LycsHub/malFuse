## ADDED Requirements

### Requirement: SQLite schema with malicious_packages table
The database SHALL contain a `malicious_packages` table with columns: `name` (TEXT NOT NULL), `version` (TEXT), `ecosystem` (TEXT NOT NULL), `published` (TEXT), `source` (TEXT), and a composite index on `(name, ecosystem, version)`.

#### Scenario: Table creation
- **WHEN** the database is first initialized
- **THEN** the `malicious_packages` table exists with all required columns
- **AND** the composite index on `(name, ecosystem, version)` exists

### Requirement: SQLite schema with update_state table
The database SHALL contain an `update_state` table with columns: `ecosystem` (TEXT NOT NULL UNIQUE), `last_commit` (TEXT), `last_updated` (TEXT).

#### Scenario: Update state tracking
- **WHEN** an update runs for ecosystem "pypi"
- **THEN** the `update_state` table records the new commit hash and timestamp for that ecosystem

### Requirement: WAL journal mode enabled
The database SHALL be opened in WAL journal mode to allow concurrent reads and writes.

#### Scenario: WAL mode active
- **WHEN** the database is opened
- **THEN** the journal mode is set to "wal"
