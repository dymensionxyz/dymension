# Test File Refactoring Example

## Current State: dym_name_test.go (3,505 lines)

The `x/dymns/keeper/dym_name_test.go` file is currently a monolithic test file that should be split into focused, parallelizable test files.

## Proposed Split Structure

### 1. Basic Operations: `dym_name_basic_test.go`
**Focus**: Core CRUD operations
**Estimated lines**: ~800
**Tests include**:
- `TestKeeper_GetSetDeleteDymName`
- `TestKeeper_GetDymNameWithExpirationCheck`
- `TestKeeper_GetAllDymNamesAndNonExpiredDymNames`
- `TestKeeper_PruneDymName`

```go
//go:build unit
// +build unit

package keeper_test

import (
    "testing"
    "time"
    
    dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

func TestDymNameBasicOperations(t *testing.T) {
    suite := setupKeeperTestSuite(t)
    
    t.Run("CRUD Operations", func(t *testing.T) {
        t.Parallel()
        suite.TestKeeper_GetSetDeleteDymName()
    })
    
    t.Run("Expiration Check", func(t *testing.T) {
        t.Parallel() 
        suite.TestKeeper_GetDymNameWithExpirationCheck()
    })
    
    // ... other tests
}
```

### 2. Ownership Operations: `dym_name_ownership_test.go`
**Focus**: Owner and controller changes
**Estimated lines**: ~600
**Tests include**:
- `TestKeeper_BeforeAfterDymNameOwnerChanged`
- `TestKeeper_GetDymNamesOwnedBy`
- Owner validation tests

```go
//go:build unit
// +build unit

package keeper_test

func TestDymNameOwnership(t *testing.T) {
    suite := setupKeeperTestSuite(t)
    
    t.Run("Owner Changes", func(t *testing.T) {
        t.Parallel()
        suite.TestKeeper_BeforeAfterDymNameOwnerChanged()
    })
    
    t.Run("Ownership Queries", func(t *testing.T) {
        t.Parallel()
        suite.TestKeeper_GetDymNamesOwnedBy()
    })
}
```

### 3. Configuration Management: `dym_name_config_test.go`
**Focus**: DymName configuration changes
**Estimated lines**: ~700
**Tests include**:
- `TestKeeper_BeforeAfterDymNameConfigChanged`
- Configuration validation tests

### 4. Address Resolution: `dym_name_resolution_test.go`
**Focus**: Address resolution functionality  
**Estimated lines**: ~1,200
**Tests include**:
- `TestKeeper_ResolveByDymNameAddress`
- `TestKeeper_ReverseResolveDymNameAddress`
- `Test_ParseDymNameAddress`

### 5. Integration Tests: `dym_name_integration_test.go`
**Focus**: Cross-module interactions
**Estimated lines**: ~400
**Tests include**:
- Complex scenarios involving multiple modules
- Chain alias integration

```go
//go:build integration
// +build integration

package keeper_test

func TestDymNameIntegration(t *testing.T) {
    if testing.Short() {
        t.Skip("Skipping integration tests in short mode")
    }
    
    suite := setupFullKeeperTestSuite(t)
    
    t.Run("Cross Module Integration", func(t *testing.T) {
        // Complex integration scenarios
    })
}
```

## Implementation Benefits

### 1. Parallel Execution
```bash
# Before: Sequential execution of 3,505 lines
go test ./x/dymns/keeper/ -v

# After: Parallel execution of 5 smaller files
go test ./x/dymns/keeper/ -v -parallel 5
```

### 2. Selective Testing
```bash
# Test only basic operations during development
go test ./x/dymns/keeper/ -run=TestDymNameBasicOperations

# Skip integration tests for fast feedback
go test ./x/dymns/keeper/ -short
```

### 3. Better Organization
- Each file has a clear, single responsibility
- Easier to navigate and maintain
- Reduced cognitive load for developers

### 4. Improved Performance
- **Compilation**: Smaller files compile faster
- **Parallelization**: Tests can run simultaneously
- **Memory**: Lower memory footprint per test file
- **Caching**: More granular test result caching

## Shared Test Utilities

Create `x/dymns/keeper/testutil.go` for common setup:

```go
package keeper_test

import (
    "testing"
    "time"
    
    dymnstypes "github.com/dymensionxyz/dymension/v3/x/dymns/types"
)

// Shared test data
var (
    DefaultTestOwner = testAddr(1).bech32()
    DefaultExpiry    = time.Now().Add(time.Hour).Unix()
)

// Factory for creating test DymNames
func NewTestDymName(name string) dymnstypes.DymName {
    return dymnstypes.DymName{
        Name:       name,
        Owner:      DefaultTestOwner,
        Controller: DefaultTestOwner,
        ExpireAt:   DefaultExpiry,
        Configs: []dymnstypes.DymNameConfig{{
            Type:  dymnstypes.DymNameConfigType_DCT_NAME,
            Path:  "www",
            Value: DefaultTestOwner,
        }},
    }
}

// Quick setup for unit tests
func setupKeeperTestSuite(t *testing.T) *KeeperTestSuite {
    suite := &KeeperTestSuite{}
    suite.SetT(t)
    suite.SetupTest()
    return suite
}

// Full setup for integration tests
func setupFullKeeperTestSuite(t *testing.T) *KeeperTestSuite {
    suite := &KeeperTestSuite{}
    suite.SetT(t)
    suite.SetupSuite()
    suite.SetupTest()
    return suite
}
```

## Migration Plan

### Step 1: Create New Files
1. Create the 5 new test files with build tags
2. Move appropriate test functions to each file
3. Add parallel execution markers

### Step 2: Update Shared Code
1. Extract common setup to `testutil.go`
2. Replace repetitive setup with factories
3. Optimize test data creation

### Step 3: Optimize CI/CD
1. Update GitHub Actions to use new Makefile targets
2. Add fast feedback loop for PR validation
3. Implement test sharding for parallel CI execution

### Step 4: Validate Performance
1. Measure before/after performance
2. Ensure test coverage is maintained
3. Verify all tests pass independently

## Expected Performance Impact

**Before**:
```bash
# Single large file: ~3-4 minutes
go test ./x/dymns/keeper/dym_name_test.go -v
```

**After**:
```bash
# 5 parallel files: ~45-60 seconds
go test ./x/dymns/keeper/ -v -parallel 5

# Unit tests only: ~20-30 seconds  
go test ./x/dymns/keeper/ -short -parallel 5

# Quick validation: ~10-15 seconds
go test ./x/dymns/keeper/ -run=TestDymNameBasicOperations
```

This refactoring approach should be applied to all test files > 1,000 lines:
- `x/dymns/keeper/grpc_query_test.go` (3,501 lines)
- `x/dymns/keeper/msg_server_update_resolve_address_test.go` (2,066 lines)  
- `x/streamer/keeper/abci_test.go` (1,958 lines)
- `x/dymns/keeper/msg_server_place_buy_order_test.go` (1,600+ lines)