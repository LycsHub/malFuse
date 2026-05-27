## ADDED Requirements

### Requirement: Typo-squatting detection uses Levenshtein distance
The typo-squatting check SHALL compute the Levenshtein edit distance between the requested package name and each name in the top-2000 popular packages list. If any distance is less than or equal to the configured threshold, the package SHALL be blocked.

#### Scenario: Close match to popular package is blocked
- **WHEN** a request is made for package `requets`
- **AND** `requests` is in the top-2000 list
- **AND** the Levenshtein distance is 1 (threshold: 2)
- **THEN** the engine returns BLOCK with reason "typo-squatting"

#### Scenario: Exact match to popular package passes
- **WHEN** a request is made for package `requests`
- **AND** `requests` is in the top-2000 list
- **AND** the Levenshtein distance is 0
- **THEN** the engine returns PASS

#### Scenario: Far match to popular package passes
- **WHEN** a request is made for package `my-obscure-lib`
- **AND** no package in the top-2000 list has an edit distance <= 2
- **THEN** the engine returns PASS

### Requirement: Short package names skip typo-squatting
The typo-squatting check SHALL skip packages with a name shorter than 3 characters to avoid excessive false positives.

#### Scenario: Two-character package name is not checked
- **WHEN** a request is made for package `ab`
- **THEN** the typo-squatting check is skipped
- **AND** the engine returns PASS

### Requirement: Top-2000 list is embedded at build time
The top-2000 popular packages list SHALL be embedded in the binary using `//go:embed` from a data file included in the repository.

#### Scenario: List is available without network access
- **WHEN** the proxy starts
- **THEN** the top-2000 list is loaded from the embedded file
- **AND** no network request is made to fetch the list
