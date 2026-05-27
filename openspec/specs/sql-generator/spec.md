## ADDED Requirements

### Requirement: Generate SQL insert statements
The system SHALL generate `INSERT OR REPLACE INTO malicious_packages (name, version, ecosystem, published, source) VALUES (...)` statements for each parsed malicious package.

#### Scenario: SQL file with inserts
- **WHEN** output mode is "sql"
- **AND** 3 malicious packages are parsed
- **THEN** the output file contains 3 INSERT OR REPLACE statements

### Requirement: Generate SQL delete statements for removed entries
The system SHALL generate `DELETE FROM malicious_packages WHERE name=? AND ecosystem=? AND (version=? OR (version IS NULL AND ? IS NULL))` statements for packages whose OSV files were deleted from the repo.

#### Scenario: SQL file with deletes
- **WHEN** a package was previously in the database
- **AND** its OSV file was removed in the latest commit
- **THEN** the output file contains a DELETE statement for that package

### Requirement: Output to specified file path
The system SHALL write generated SQL to a user-specified file path when in SQL output mode.

#### Scenario: Output file written
- **WHEN** output mode is "sql" with path `updates.sql`
- **THEN** the file `updates.sql` is created with valid SQL
