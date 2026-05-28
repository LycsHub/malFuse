## ADDED Requirements

### Requirement: URL detection in script content
The system SHALL extract and detect HTTP/HTTPS URLs embedded in install script content using regex pattern matching.

#### Scenario: Suspicious URL triggers block
- **WHEN** scanning a file containing `http://evil.com/steal?data=`
- **THEN** the network detector returns BLOCK with reason "network"

#### Scenario: Common benign URLs pass
- **WHEN** scanning a file containing `https://pypi.org/project/requests/`
- **AND** the URL matches a known registry domain
- **THEN** the network detector returns PASS

### Requirement: IP address detection
The system SHALL detect IPv4 addresses embedded in script content.

#### Scenario: Hardcoded IP triggers block
- **WHEN** scanning a file containing `connect("192.168.1.100", 4444)`
- **AND** `allow_private_ips` is false
- **THEN** the network detector returns BLOCK with reason "network"

#### Scenario: Private IP passes when allowed
- **WHEN** scanning a file containing `10.0.0.1`
- **AND** `allow_private_ips` is true
- **THEN** the network detector returns PASS
