## 1. JS Analyzer

- [x] 1.1 Implement `parsePackageJSON(data []byte) (map[string]string, error)` extracting scripts fields
- [x] 1.2 Implement `extractReferencedJSFiles(scripts map[string]string) []string` finding JS file references
- [x] 1.3 Implement `analyzeJSArchive(tr *tar.Reader, cfg ScanConfig) ScanResult` orchestrating package.json + JS files
- [x] 1.4 Unit tests: malicious preinstall script, clean scripts, node evil.js reference

## 2. Python Analyzer

- [x] 2.1 Implement `scanSetupPy(content []byte, cfg ScanConfig) ScanResult`
- [x] 2.2 Implement `scanInitPy(content []byte, cfg ScanConfig) ScanResult`
- [x] 2.3 Implement `scanPthFile(content []byte, cfg ScanConfig) ScanResult` with line-by-line import detection
- [x] 2.4 Implement `scanPyprojectToml(content []byte, cfg ScanConfig) ScanResult` extracting build-backend
- [x] 2.5 Implement `analyzePythonArchive(tr *tar.Reader, cfg ScanConfig) ScanResult` dispatching by file type
- [x] 2.6 Unit tests: malicious setup.py, __init__.py, .pth import, pyproject.toml build-backend

## 3. Scanner Refactor

- [x] 3.1 Remove `isInstallScript()` from scanner.go
- [x] 3.2 Rewrite `scanTar()` to detect ecosystem and dispatch to `analyzeJSArchive` or `analyzePythonArchive`
- [x] 3.3 Implement ecosystem detection from package.json presence
- [x] 3.4 Update existing scanner tests to use new analyzers
- [x] 3.5 Run all tests — confirm no regressions
