## ADDED Requirements

### Requirement: Stream decompression and file extraction
The system SHALL decompress tar, gzip, and zip streams and extract individual files matching target install script patterns.

#### Scenario: setup.py extracted from tar.gz
- **WHEN** scanning a gzipped tar stream containing `package-1.0/setup.py`
- **THEN** `setup.py` content is extracted and passed to detectors

#### Scenario: preinstall.js extracted from tarball
- **WHEN** scanning a tar stream from npm containing `package/preinstall.js`
- **THEN** `preinstall.js` content is extracted and passed to detectors

### Requirement: File size limits enforced
The system SHALL skip analysis of files larger than `max_file_size` and SHALL stop reading the stream when total bytes exceed `max_total_size`.

#### Scenario: Large file skipped
- **WHEN** a file in the archive exceeds 5MB
- **THEN** the file is skipped but scanning continues for other files

#### Scenario: Total stream limit enforced
- **WHEN** total stream bytes exceed 50MB
- **THEN** the scanner stops reading and returns PASS

### Requirement: Fail-open on scan error
The system SHALL return PASS when a scan error occurs (OOM, corrupt archive, timeout) and log a warning.

#### Scenario: Corrupt archive returns pass
- **WHEN** the stream contains a corrupt tar/gzip that cannot be parsed
- **THEN** the scanner logs a warning and returns PASS
