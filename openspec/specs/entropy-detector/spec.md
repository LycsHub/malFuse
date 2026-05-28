## ADDED Requirements

### Requirement: Shannon entropy calculation
The system SHALL calculate the Shannon entropy of a byte stream using `H = -Σ p(x)·log₂(p(x))` where p(x) is the frequency of each byte value 0-255.

#### Scenario: Normal text has low entropy
- **WHEN** scanning a plain ASCII text file
- **THEN** the entropy value is less than 4.5

#### Scenario: Encrypted content has high entropy
- **WHEN** scanning a base64-encoded random payload
- **THEN** the entropy value exceeds 5.0

#### Scenario: Empty input
- **WHEN** scanning an empty byte slice
- **THEN** the entropy value is 0.0

### Requirement: Configurable entropy threshold
The system SHALL compare the calculated entropy against a configurable threshold and SHALL block if the entropy exceeds the threshold.

#### Scenario: Entropy below threshold passes
- **WHEN** entropy threshold is 5.0
- **AND** file entropy is 4.2
- **THEN** the entropy detector returns PASS

#### Scenario: Entropy above threshold blocks
- **WHEN** entropy threshold is 4.5
- **AND** file entropy is 5.3
- **THEN** the entropy detector returns BLOCK with reason "entropy"
