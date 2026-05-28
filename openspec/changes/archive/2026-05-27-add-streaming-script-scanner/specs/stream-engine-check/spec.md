## ADDED Requirements

### Requirement: StreamChecker interface defined
The engine package SHALL define a `StreamChecker` interface with a `StreamCheck(ctx context.Context, req Request, body io.Reader) ScanResult` method.

#### Scenario: StreamCheck returns BLOCK on malicious content
- **WHEN** a StreamChecker implementation detects malicious content in the stream
- **THEN** `StreamCheck` returns `ScanResult{Block: true, Reason: "entropy"}`

#### Scenario: StreamCheck returns PASS on clean content
- **WHEN** a StreamChecker implementation detects no issues in the stream
- **THEN** `StreamCheck` returns `ScanResult{Block: false}`

### Requirement: Proxy invokes StreamCheck during forward
The proxy SHALL invoke `StreamCheck` on a cloned stream via TeeReader when a StreamChecker is configured and the response is a tarball.

#### Scenario: Scanner blocks mid-stream
- **WHEN** a StreamChecker returns BLOCK during streaming
- **THEN** the proxy logs the block reason
- **AND** the client connection is terminated

#### Scenario: No StreamChecker configured
- **WHEN** the proxy has no StreamChecker set
- **THEN** the forward proceeds without streaming analysis
