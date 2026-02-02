# TuneTudo - Testing Guide

## 📋 Test Coverage

### Unit Tests
All core services have comprehensive unit tests:

1. **AuthService** - `services/auth_service_test.go`
   - User registration
   - User login
   - JWT token generation and validation
   - Password policy checking
   - User retrieval

2. **PlaylistService** - `services/playlist_service_test.go`
   - Create playlists
   - Get user playlists
   - Add/remove songs
   - Delete playlists
   - Authorization checks

3. **SearchService** - `services/search_service_test.go`
   - Full-text search
   - Category browsing
   - User upload privacy
   - Empty query handling

4. **PlaybackService** - `services/playback_service_test.go`
   - Get song by ID
   - Stream authorization
   - Recent songs
   - User upload exclusion

5. **Integration Tests** - `main_test.go`
   - Complete auth flow
   - Playlist management flow
   - Search functionality
   - API endpoint validation

## 🚀 Running Tests

### Run All Tests
```bash
go test ./...
```

### Run Tests with Coverage
```bash
go test ./... -cover
```

### Run Tests with Detailed Coverage
```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Run Specific Test File
```bash
go test ./services -run TestRegisterUser
```

### Run Tests in Verbose Mode
```bash
go test -v ./...
```

### Run Tests with Race Detection
```bash
go test -race ./...
```

## 📁 Test File Structure

```
tunetudo/
├── services/
│   ├── auth_service.go
│   ├── auth_service_test.go        ✅ Unit tests
│   ├── playlist_service.go
│   ├── playlist_service_test.go    ✅ Unit tests
│   ├── search_service.go
│   ├── search_service_test.go      ✅ Unit tests
│   ├── playback_service.go
│   ├── playback_service_test.go    ✅ Unit tests
│   └── test_helper.go              ✅ Shared test utilities
│
├── main_test.go                     ✅ Integration tests
└── go.mod
```

## 🔧 Setup Requirements

### Install Test Dependencies

Add to `go.mod`:
```bash
go get github.com/stretchr/testify/assert
go get github.com/stretchr/testify/require
```

Or run:
```bash
go mod tidy
```

### Required Packages
- `github.com/stretchr/testify` - Assertion library
- `github.com/mattn/go-sqlite3` - SQLite driver (already included)

## 📊 Test Statistics

### Coverage Goals
- **Auth Service**: >90% coverage
- **Playlist Service**: >85% coverage
- **Search Service**: >80% coverage
- **Playback Service**: >75% coverage
- **Overall**: >80% coverage

### Test Count
```
AuthService Tests:      8 tests
PlaylistService Tests:  8 tests
SearchService Tests:    5 tests
PlaybackService Tests:  4 tests
Integration Tests:      7 tests
--------------------------------
Total:                  32 tests
```

## 🧪 Test Categories

### 1. Happy Path Tests ✅
Tests that verify normal, expected behavior:
- Valid user registration
- Successful login
- Creating playlists
- Adding songs to playlists
- Searching for songs

### 2. Error Path Tests ❌
Tests that verify error handling:
- Invalid passwords
- Duplicate usernames
- Non-existent resources
- Unauthorized access
- Invalid input

### 3. Edge Case Tests 🔍
Tests for boundary conditions:
- Empty queries
- Maximum limits
- Duplicate operations
- Concurrent access

### 4. Privacy Tests 🔒
Tests for data isolation:
- User uploads not in public search
- User playlists are private
- Unauthorized access prevention

## 📝 Writing New Tests

### Test Template
```go
func TestYourFeature(t *testing.T) {
    // Setup
    service, cleanup := setupTestService(t)
    defer cleanup()

    // Test cases
    tests := []struct {
        name        string
        input       interface{}
        expectError bool
        errorMsg    string
    }{
        {
            name:        "Valid case",
            input:       validInput,
            expectError: false,
        },
        {
            name:        "Invalid case",
            input:       invalidInput,
            expectError: true,
            errorMsg:    "expected error",
        },
    }

    // Run tests
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result, err := service.YourMethod(tt.input)

            if tt.expectError {
                assert.Error(t, err)
                assert.Contains(t, err.Error(), tt.errorMsg)
            } else {
                require.NoError(t, err)
                assert.NotNil(t, result)
            }
        })
    }
}
```

## 🔍 Testing Best Practices

### 1. Use Table-Driven Tests
```go
tests := []struct {
    name     string
    input    string
    expected string
}{
    {"case 1", "input1", "output1"},
    {"case 2", "input2", "output2"},
}
```

### 2. Clean Up Resources
```go
defer cleanup()  // Always clean up
```

### 3. Use Require for Critical Checks
```go
require.NoError(t, err)  // Stops test if fails
assert.NoError(t, err)   // Continues test
```

### 4. Test Isolation
- Each test should be independent
- Use separate test databases
- Clean up after each test

### 5. Meaningful Names
```go
// Good
func TestRegisterUser_WithValidData_CreatesUser(t *testing.T)

// Bad
func TestFunc1(t *testing.T)
```

## 🐛 Debugging Failed Tests

### View Test Output
```bash
go test -v ./services -run TestRegisterUser
```

### Check Test Database
```bash
# Test databases are named: test_TestName.db
sqlite3 test_TestRegisterUser.db
.tables
SELECT * FROM users;
```

### Add Debug Logging
```go
t.Logf("Debug: %v", variable)
```

### Run Single Test
```bash
go test -run TestSpecificTest ./services
```

## 📈 Continuous Integration

### GitHub Actions Example
```yaml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2
      - uses: actions/setup-go@v2
        with:
          go-version: 1.21
      - run: go test -v -race -coverprofile=coverage.out ./...
      - run: go tool cover -func=coverage.out
```

## 🎯 Key Test Scenarios

### Authentication Tests
- ✅ Register with valid data
- ✅ Register with short password (fail)
- ✅ Register with duplicate email (fail)
- ✅ Login with correct credentials
- ✅ Login with wrong password (fail)
- ✅ Token generation and validation
- ✅ Expired token handling

### Playlist Tests
- ✅ Create playlist
- ✅ Create duplicate playlist (fail)
- ✅ Add song to playlist
- ✅ Add duplicate song (fail)
- ✅ Remove song from playlist
- ✅ Delete playlist
- ✅ Access other user's playlist (fail)

### Search Tests
- ✅ Search by song title
- ✅ Search by artist name
- ✅ Search by album name
- ✅ Empty query returns empty results
- ✅ User uploads excluded from search
- ✅ Category filtering

### Privacy Tests
- ✅ User uploads not in public search
- ✅ User uploads not in categories
- ✅ User uploads not in recent songs
- ✅ Users can only see own uploads

## 🚨 Common Test Issues

### Issue 1: Database Locked
**Solution**: Ensure cleanup functions are called
```go
defer cleanup()
```

### Issue 2: Test Files Not Cleaned Up
**Solution**: Use t.Cleanup()
```go
t.Cleanup(func() {
    os.Remove(dbPath)
})
```

### Issue 3: Race Conditions
**Solution**: Run with race detector
```bash
go test -race ./...
```

### Issue 4: Flaky Tests
**Solution**: Ensure test isolation
- Don't rely on global state
- Reset database between tests
- Don't depend on test execution order

## 📚 Additional Resources

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Documentation](https://github.com/stretchr/testify)
- [Go Testing Best Practices](https://golang.org/doc/code#Testing)

## ✅ Pre-Deployment Checklist

Before deploying, ensure:
- [ ] All tests pass: `go test ./...`
- [ ] No race conditions: `go test -race ./...`
- [ ] Coverage > 80%: `go test -cover ./...`
- [ ] Integration tests pass
- [ ] Manual testing completed
- [ ] Edge cases covered
- [ ] Error handling tested

## 🎉 Running the Full Test Suite

```bash
# Complete test run with coverage
go test -v -race -coverprofile=coverage.out ./...

# View coverage report
go tool cover -html=coverage.out

# Check coverage percentage
go tool cover -func=coverage.out | grep total
```

Expected output:
```
PASS
coverage: 85.3% of statements
ok      tunetudo/services    2.134s
```