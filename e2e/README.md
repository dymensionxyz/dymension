## Quick Start
Make sure you have Docker installed. For testing in local machine you need 2 steps:

1. Build a debug image with your code change
```bash
make docker-build-debug
```
2. Run Test-case you want to test. Example:
```bash
make e2e-test-ibc
```