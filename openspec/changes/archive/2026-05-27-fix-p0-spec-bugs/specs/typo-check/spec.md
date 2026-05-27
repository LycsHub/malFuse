## MODIFIED Requirements

### Requirement: Top-2000 list is embedded at build time
The top-2000 popular packages list SHALL be embedded in the binary using `//go:embed` from `packages.txt` in the engine package directory. The file SHALL contain approximately 2000 popular package names, one per line.

#### Scenario: List is available without file I/O
- **WHEN** the proxy starts
- **THEN** the top-2000 list is loaded from the embedded data
- **AND** no `os.ReadFile` call is made to load the file

#### Scenario: Binary is self-contained
- **WHEN** the proxy binary is copied to a different directory
- **AND** `packages.txt` is not present in the working directory
- **THEN** the typo-squatting check still has access to the full list
