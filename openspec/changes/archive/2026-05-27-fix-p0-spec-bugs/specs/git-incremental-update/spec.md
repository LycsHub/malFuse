## MODIFIED Requirements

### Requirement: Update commit hash after successful update
The system SHALL store the latest `origin/main` commit hash in `update_state.last_commit` and the current time in `update_state.last_updated` after a successful update.

#### Scenario: Commit hash and timestamp saved
- **WHEN** an update completes successfully
- **THEN** `update_state.last_commit` is set to the current HEAD of origin/main
- **AND** `update_state.last_updated` is set to the current time in RFC3339 format
