## MODIFIED Requirements

### Requirement: Database opened read-only
The malFuse proxy SHALL open the SQLite database in read-only mode using `sql.Open("sqlite", path+"?mode=ro")` without executing WAL pragma or migration DDL. The database file SHALL already exist and be created by `malfuse-db`.

#### Scenario: Read-only connection
- **WHEN** the proxy opens `malfuse.db`
- **THEN** the connection is opened with `mode=ro`
- **AND** no WAL pragma is executed
- **AND** no table migration DDL is executed

#### Scenario: Missing database file
- **WHEN** `malfuse.db` does not exist at the configured path
- **THEN** the open returns an error
- **AND** the proxy logs a warning and continues without the malicious-db check
