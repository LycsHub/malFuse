## ADDED Requirements

### Requirement: Fetch latest changes
The system SHALL run `git fetch` in the cloned repository to get the latest changes from the remote.

#### Scenario: Fetch updates
- **WHEN** `malfuse-db update` is run on an existing clone
- **THEN** `git fetch origin main` is executed

### Requirement: Diff against tracked commit
The system SHALL compute changed files using `git diff --name-status <last_commit>..origin/main` where `last_commit` is retrieved from the `update_state` table.

#### Scenario: Incremental file list
- **WHEN** `last_commit` is known
- **THEN** only files added (A), modified (M), or deleted (D) since that commit are returned

### Requirement: Handle full scan on new clone
The system SHALL perform a full scan of all OSV JSON files in `osv/malicious/<ecosystem>/` when no prior commit is tracked.

#### Scenario: First-time full scan
- **WHEN** `last_commit` is NULL
- **THEN** all files in `osv/malicious/` are enumerated and parsed

### Requirement: Update commit hash after successful update
The system SHALL store the latest `origin/main` commit hash in `update_state.last_commit` after a successful update.

#### Scenario: Commit hash saved
- **WHEN** an update completes successfully
- **THEN** `update_state.last_commit` is set to the current HEAD of origin/main
- **AND** `update_state.last_updated` is set to the current time
