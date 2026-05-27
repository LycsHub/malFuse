## ADDED Requirements

### Requirement: Clone repo on first run
The system SHALL clone the `ossf/malicious-packages` repository with `--depth 1` on the first run if the local clone directory does not exist.

#### Scenario: First run clones the repo
- **WHEN** the local repo directory does not exist
- **AND** `malfuse-db update` is run
- **THEN** a shallow clone of ossf/malicious-packages is created

### Requirement: Parse OSV JSON files
The system SHALL parse OSV JSON files from the cloned repository and extract `id`, `summary`, `affected[].package.name`, `affected[].package.ecosystem`, `affected[].versions[]`, and `published` fields.

#### Scenario: Parse a valid OSV JSON file
- **WHEN** a valid OSV JSON file is read from `osv/malicious/pypi/<pkg>/MAL-YYYY-NNNN.json`
- **THEN** the parser extracts package name, ecosystem ("PyPI" → "pypi"), versions, and published timestamp

### Requirement: Map ecosystem names to lowercase
The system SHALL normalize ecosystem names from OSV format ("PyPI", "npm") to lowercase ("pypi", "npm") for database consistency.

#### Scenario: PyPI ecosystem normalized
- **WHEN** an OSV file has ecosystem "PyPI"
- **THEN** the normalized ecosystem is "pypi"

### Requirement: Skip withdrawn and unmergable entries
The system SHALL only process files under `osv/malicious/` and SHALL skip files under `osv/withdrawn/` and `osv/unmergable/`.

#### Scenario: Withdrawn file is skipped
- **WHEN** scanning the repo directory
- **AND** a file path starts with `osv/withdrawn/`
- **THEN** the file is not parsed
