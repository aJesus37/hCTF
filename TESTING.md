# hCTF2 - Test Suite Documentation

## Overview

The hCTF2 test suite validates that all pages render correctly with proper content and that navigation works as expected. This ensures the routing fix stays fixed and prevents regressions.

## Running Tests

### Run all tests
```bash
task test
```

### Run specific test
```bash
go test -v -run TestPageContent
```

### Run tests with coverage
```bash
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

## Test Coverage

### 1. TestPageContent
Validates that each page renders with all expected content.

**Pages tested:**
- **Home** (`/`): Welcome message, quick links, statistics
- **Login** (`/login`): Email/password fields, register link
- **Register** (`/register`): Name/email/password fields, login link
- **Challenges** (`/challenges`): Challenge list, category/difficulty filters
- **Scoreboard** (`/scoreboard`): Rankings table, user scores
- **SQL Playground** (`/sql`): SQL editor, example queries, schema

**Validation:**
- All required content elements are present
- Page structure is valid (DOCTYPE, html, body, nav tags)

### 2. TestNavigationLinks
Ensures the navigation bar contains links to all main pages.

**Links verified:**
- `/challenges` - Challenge browser
- `/scoreboard` - Rankings
- `/sql` - SQL Playground
- `/login` - User login
- `/register` - User registration

### 3. TestAPIEndpoints
Validates that API endpoints return proper HTTP responses.

**Endpoints tested:**
- `GET /api/challenges` - Returns challenge list (HTTP 200)
- `GET /api/scoreboard` - Returns rankings (HTTP 200)
- `GET /api/sql/snapshot` - Returns data snapshot (HTTP 200)

### 4. TestPageContentConsistency
Ensures the same page renders identically on repeated requests.

**Why this matters:**
- Prevents random rendering bugs
- Validates no stale state between requests
- Ensures template rendering is stable

### 5. TestNoPageCollision
Confirms that pages don't render content from other pages (regression test for the routing bug).

**Examples:**
- Login page should NOT contain challenge-specific content
- Scoreboard page should NOT contain SQL playground content
- Register page should NOT contain login/home content

**This test prevents the original navigation bug from returning.**

## Test Results

All tests pass:
```
PASS: TestPageContent (0.01s)
  ✅ Home Page
  ✅ Login Page
  ✅ Register Page
  ✅ Challenges Page
  ✅ Scoreboard Page
  ✅ SQL Playground Page

PASS: TestNavigationLinks (0.00s)
  ✅ All navigation links present

PASS: TestAPIEndpoints (0.00s)
  ✅ /api/challenges returns 200
  ✅ /api/scoreboard returns 200
  ✅ /api/sql/snapshot returns 200

PASS: TestPageContentConsistency (0.01s)
  ✅ All pages render consistently

PASS: TestNoPageCollision (0.00s)
  ✅ No cross-page content
  ✅ Login page isolated
  ✅ Register page isolated
  ✅ Challenges page isolated
```

## Adding New Tests

When adding new pages or features:

1. **Add to TestPageContent** - Verify page renders with expected content
2. **Add to TestNavigationLinks** (if page has nav link)
3. **Add to TestAPIEndpoints** (if page has API)
4. **Add to TestNoPageCollision** (to prevent regressions)

Example:
```go
{
  name: "New Page",
  method: "GET",
  path: "/new-page",
  contentMust: []string{
    "Expected heading",
    "Expected content",
  },
}
```

## Test Architecture

- **In-Memory Database**: Uses SQLite in-memory database (`:memory:`) for fast test execution
- **No External Dependencies**: Tests don't require external services
- **Isolated Tests**: Each test runs independently
- **Fast Execution**: Full test suite completes in ~30ms

## Common Issues

### Test fails with "template not found"
- Ensure all template files are in `internal/views/templates/`
- Run `task build` to embed templates in binary

### API test returns wrong status code
- Check that route is registered in router setup
- Verify middleware isn't blocking the endpoint

### Content test fails
- Verify content string is present in page HTML
- Check for typos in expected content
- Test may be too strict (whitespace, case sensitivity)

## CI/CD Integration

To run tests in CI/CD pipeline:

```bash
#!/bin/bash
set -e

# Run tests with verbose output
go test -v ./...

# Optionally generate coverage
go test -v -coverprofile=coverage.out ./...

# Check coverage threshold (optional)
# coverage=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | sed 's/%//')
# if (( $(echo "$coverage < 70" | bc -l) )); then
#   echo "Coverage too low: $coverage%"
#   exit 1
# fi
```

## Performance

- **Full test suite**: ~30ms
- **Per test**: 1-10ms
- **Memory usage**: <10MB

Tests are fast enough to run on every commit.

## Future Improvements

- [ ] Add database integration tests
- [ ] Add API authentication tests
- [ ] Add form submission tests
- [ ] Add performance regression tests
- [ ] Add E2E tests with browser automation (Selenium/Playwright)

## References

- [Testing in Go](https://golang.org/doc/effective_go#testing)
- [Table-driven tests](https://github.com/golang/go/wiki/TableDrivenTests)
- [httptest package](https://pkg.go.dev/net/http/httptest)
