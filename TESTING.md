# Testing Documentation

This document outlines the comprehensive test coverage for the URL Shortener Service.

## 📊 Coverage Overview

**Overall Test Coverage: 85.2%**

| Package            | Coverage | Status       |
| ------------------ | -------- | ------------ |
| `internal/http`    | 76.7%    | ✅ Good      |
| `internal/id`      | 94.1%    | ✅ Excellent |
| `internal/storage` | 92.9%    | ✅ Excellent |

## 🧪 Test Categories

### 1. HTTP Handler Tests (`internal/http`)

#### **Integration Tests**

- **`TestCreateURL_Integration`** - Comprehensive URL creation testing
- **`TestRedirectURL_Integration`** - Enhanced 404 error path testing
- **`TestRedirectURL_EdgeCases`** - Additional edge case scenarios
- **`TestCreateURL_Concurrent`** - Concurrency and race condition testing
- **`TestDeleteURL_Integration`** - URL deletion and idempotency testing
- **`TestDeleteURL_Concurrent`** - Concurrent deletion safety testing

#### **Coverage Breakdown**

| Function      | Coverage | Test Coverage    |
| ------------- | -------- | ---------------- |
| `NewHandler`  | 100.0%   | ✅ Fully tested  |
| `SetupRoutes` | 100.0%   | ✅ Fully tested  |
| `CreateURL`   | 69.2%    | ⚠️ Good coverage |
| `RedirectURL` | 83.3%    | ✅ Well tested   |
| `DeleteURL`   | 100.0%   | ✅ Fully tested  |

### 2. ID Generator Tests (`internal/id`)

#### **Unit Tests**

- **`TestNewGenerator`** - Generator initialization
- **`TestGenerator_Generate`** - Key generation and uniqueness
- **`TestGenerator_ValidateKey`** - Key format validation
- **`TestGenerator_Generate_Distribution`** - Character distribution analysis
- **`TestGenerator_Generate_RandomError`** - Error handling

#### **Coverage Breakdown**

| Function       | Coverage | Test Coverage   |
| -------------- | -------- | --------------- |
| `NewGenerator` | 100.0%   | ✅ Fully tested |
| `Generate`     | 100.0%   | ✅ Fully tested |
| `ValidateKey`  | 83.3%    | ✅ Well tested  |

### 3. Storage Tests (`internal/storage`)

#### **Integration Tests**

- **`TestRedisStore_Set`** - Key storage with TTL
- **`TestRedisStore_Get`** - Key retrieval and TTL refresh
- **`TestRedisStore_Delete`** - Key deletion
- **`TestRedisStore_ConnectionFailure`** - Error handling
- **`TestRedisStore_Concurrent`** - Concurrent operations
- **`TestRedisStore_TTLExpiration`** - TTL behavior

#### **Coverage Breakdown**

| Function        | Coverage | Test Coverage        |
| --------------- | -------- | -------------------- |
| `NewRedisStore` | 100.0%   | ✅ Fully tested      |
| `Set`           | 100.0%   | ✅ Fully tested      |
| `Get`           | 87.5%    | ✅ Well tested       |
| `Delete`        | 100.0%   | ✅ Fully tested      |
| `Close`         | 100.0%   | ✅ Fully tested      |
| `FlushDB`       | 0.0%     | ⚠️ Test-only utility |

## 🎯 Detailed Test Scenarios

### HTTP Handler Tests

#### **URL Creation Tests (`TestCreateURL_Integration`)**

✅ **Valid URL handling**

- Standard HTTPS URLs
- Very long URLs (100+ path segments)
- URL validation and normalization

✅ **Error Handling**

- Malformed JSON requests
- Invalid URL formats
- Missing URL fields
- Empty URLs
- URLs without schemes

✅ **Concurrency Testing (`TestCreateURL_Concurrent`)**

- 50 concurrent requests
- Key uniqueness verification
- Race condition detection

#### **URL Redirection Tests (`TestRedirectURL_Integration`)**

✅ **Successful Redirection**

- Valid key resolution
- Proper HTTP 302 redirect
- Location header verification

✅ **404 Error Path Testing** (Enhanced Coverage)

- **Invalid Key Formats:**
  - Special characters (`!@#`)
  - Too short keys (< 8 characters)
  - Too long keys (> 8 characters)
  - Contains spaces (URL-encoded)
  - Contains underscores
  - Contains hyphens
- **Non-existent Keys:**
  - Valid format but doesn't exist
  - Multiple valid format examples

#### **Edge Case Tests (`TestRedirectURL_EdgeCases`)**

✅ **Special Scenarios**

- Empty key (root path `/`)
- URL-encoded invalid characters
- Keys with dots and other special chars

#### **URL Deletion Tests (`TestDeleteURL_Integration`)**

✅ **Successful Deletion**

- Valid key deletion
- Redis storage verification
- Proper 200 OK response

✅ **Error Handling**

- Non-existent keys (204 No Content)
- Invalid key formats (400 Bad Request):
  - Too short keys
  - Invalid characters
  - Malformed keys
- Double deletion (204 No Content)

✅ **Concurrency Testing (`TestDeleteURL_Concurrent`)**

- 50 concurrent deletion attempts
- Exactly one successful deletion (200 OK)
- All other attempts return 204 No Content
- Final storage state verification
- Race condition prevention

### ID Generator Tests

#### **Key Generation (`TestGenerator_Generate`)**

✅ **Generation Quality**

- 100 key generation tests
- 1000 key uniqueness verification
- Base62 character set compliance

✅ **Distribution Analysis (`TestGenerator_Generate_Distribution`)**

- 10,000 key generation test
- Character frequency analysis
- Statistical distribution verification

✅ **Validation (`TestGenerator_ValidateKey`)**

- Valid key formats
- Length validation (exactly 8 chars)
- Character set validation
- Empty string handling

### Storage Tests

#### **Redis Operations**

✅ **Core Operations**

- Set with TTL (3 hours)
- Get with TTL refresh
- Delete operations
- Key collision handling

✅ **Error Scenarios**

- Connection failures
- Non-existent key handling
- Empty key/value validation

✅ **Performance & Reliability**

- 100 concurrent operations
- TTL expiration behavior (2-second test)
- Database cleanup (`FlushDB`)

## 🚀 Running Tests

### Basic Test Execution

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run specific package tests
go test ./internal/http -v
```

### Coverage Analysis

```bash
# Run tests with coverage
go test -cover ./...

# Generate detailed coverage report
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out

# Generate HTML coverage report
go tool cover -html=coverage.out -o coverage.html
```

### Benchmark Tests

```bash
# Run performance benchmarks (if available)
go test -bench=. ./...
```

## 🔧 Test Configuration

### Prerequisites

- **Redis Server**: Tests require Redis running on `localhost:6379`
- **Go Version**: Go 1.22 or later
- **Test Database**: Uses Redis DB 0 (cleared before each test)

### Docker Test Environment

```bash
# Start Redis for testing
docker run -d -p 6379:6379 --name redis-test redis:6

# Run tests
go test ./...

# Cleanup
docker stop redis-test && docker rm redis-test
```

## 📈 Coverage Goals

| Component     | Current | Target | Status               |
| ------------- | ------- | ------ | -------------------- |
| HTTP Handlers | 76.7%   | 85%    | 🔶 Needs improvement |
| ID Generator  | 94.1%   | 90%    | ✅ Exceeds target    |
| Storage Layer | 92.9%   | 90%    | ✅ Exceeds target    |
| **Overall**   | 85.2%   | 85%    | ✅ Meets target      |

## 🎯 Areas for Improvement

### HTTP Handler Coverage (76.7% → 85%+)

- **`CreateURL` function (69.2%)**:
  - Add tests for retry logic edge cases
  - Test maximum retry scenarios
  - Add Redis connection failure scenarios

### Potential Additional Tests

- **Integration Tests**: Full API workflow testing
- **Load Testing**: High-concurrency scenarios
- **Error Recovery**: Network failure handling
- **Security Tests**: Input sanitization verification

## 🏆 Test Quality Highlights

✅ **Comprehensive 404 Error Path Testing**

- 8 different invalid key format scenarios
- 3 edge case scenarios
- Proper error message validation

✅ **Concurrency Safety**

- Race condition detection
- Atomic operation verification
- Unique key generation under load

✅ **Real Redis Integration**

- No mocks - tests against real Redis
- TTL behavior verification
- Connection failure handling

✅ **Input Validation Coverage**

- Malformed requests
- Edge case inputs
- Error boundary testing

---

_Last Updated: December 2024_  
_Generated from: `go test -cover ./...` and `go tool cover -func=coverage.out`_
