## MODIFIED Requirements

### Requirement: Proxy forwards with streaming analysis
The proxy SHALL clone the upstream response body using `io.TeeReader` when a `StreamChecker` is configured and the response Content-Type indicates a tarball/zip. The cloned stream SHALL be analyzed asynchronously in a goroutine. If the StreamChecker returns BLOCK, the proxy SHALL cancel the request context to terminate the client connection.

#### Scenario: Clean package forwarded normally
- **WHEN** the StreamChecker returns PASS for a legitimate package
- **THEN** the full response body is streamed to the client
- **AND** the scanner goroutine exits cleanly

#### Scenario: Malicious script triggers mid-stream abort
- **WHEN** the StreamChecker returns BLOCK during the download
- **THEN** the proxy cancels the context
- **AND** the client connection is terminated

### Requirement: StreamChecker configuration
The proxy SHALL only enable streaming analysis when `script_scan.enabled` is `true` in the configuration and a StreamChecker implementation is provided.

#### Scenario: script_scan disabled
- **WHEN** `script_scan.enabled` is `false`
- **THEN** the proxy forwards without TeeReader cloning
