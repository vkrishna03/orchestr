# Orchestr

CLI tool that automatically validates AI-generated code through functional testing with zero configuration.

## Quick Start

```bash
# Download binary (coming soon)
curl -L https://github.com/orchestr/orchestr/releases/latest/download/orchestr-darwin -o orchestr
chmod +x orchestr
sudo mv orchestr /usr/local/bin/

# Test your project
orchestr test ./my-project
```

## What It Does

Write an `Orchestrfile` spec, and Orchestr handles the rest:
- Detects service type (Spring Boot, Vite, etc.)
- Starts the service automatically
- Runs functional tests
- Reports pass/fail with details

```bash
$ orchestr test ./my-api

Starting service... done
Running 12 functional tests...
8/12 passed, 2 critical failures

Report: orchestr-report.md
```

## Orchestrfile

Define tests in a simple DSL:

```
ENTITY: User
ENDPOINT: /api/register
METHOD: POST
DATA: {email: "test@example.com", password: "test123"}
EXPECT: 201 with {id: <uuid>, email: "test@example.com"}

---

ENTITY: User
ENDPOINT: /api/login
METHOD: POST
DATA: {email: "test@example.com", password: "test123"}
EXPECT: 200 with {token: <jwt>}
SAVE: token from body.token

---

ENTITY: User
ENDPOINT: /api/me
METHOD: GET
AUTH: Bearer {{token}}
EXPECT: 200 with {email: "test@example.com"}
```

### Type Patterns

**Basic Types:**
- `<uuid>` - Valid UUID format
- `<jwt>` - Valid JWT token
- `<int>` - Integer
- `<float>` - Floating point number
- `<bool>` - Boolean (true/false)
- `<string>` - Any non-empty string
- `<email>` - Valid email format
- `<url>` - Valid URL
- `<iso8601>` - ISO 8601 datetime
- `<any>` - Matches anything

**Structural Types:**
- `<array>` - Any array
- `<object>` - Any object

**With Constraints:**
- `<array:min=1>` - Array with at least 1 element
- `<array:max=10>` - Array with at most 10 elements
- `<array:len=5>` - Array with exactly 5 elements
- `<array:nonempty>` - Non-empty array
- `<int:gte=0>` - Integer >= 0
- `<int:gt=0,lt=100>` - Integer in range (0, 100)
- `<float:gte=0.0>` - Float >= 0.0
- `<string:nonempty>` - Non-empty string
- `<string:min=3,max=50>` - String length between 3-50
- `<object:nonempty>` - Non-empty object

**Variable References:**
- `{{var}}` - Use saved value from previous test
- `{{token}}` - Common example: auth token
- `{{userId}}` - Common example: saved user ID

## UI Testing (Planned)

High-level, intent-based UI testing for frontends:

```
TYPE: UI
BASE: http://localhost:5173

FLOW: Login
DATA: {email: "test@example.com", password: "test123"}
EXPECT: /dashboard
EXPECT_TEXT: "Welcome"

---

FLOW: Signup
DATA: {name: "John", email: "john@test.com", password: "pass123"}
EXPECT: /welcome
```

The tool auto-detects forms, matches fields, and verifies outcomes. See [UI Testing Spec](docs/ui-testing.md).

## Commands

```bash
orchestr test [path]      # Run tests
orchestr init             # Generate Orchestrfile template
orchestr discover [path]  # Show detected services
orchestr version          # Show version
```

### Options

```bash
orchestr test . --format json      # JSON output
orchestr test . --format markdown  # Markdown report
orchestr test . -v                 # Verbose output
```

## Supported Frameworks

**v0.1 (MVP):**
- Spring Boot (pom.xml, build.gradle)
- React/Vite (vite.config.ts/js)

**v1.0 (Planned):**
- Node.js/Express
- Python Flask/FastAPI
- Go Gin
- Multi-service orchestration
- External API mocking (Stripe, SendGrid, etc.)

## Development

```bash
# Build
go build -o orchestr ./cmd/orchestr

# Run
./orchestr test ./sample-project

# Test
go test ./...

# Lint
golangci-lint run
```

## Tech Stack

Go 1.21+ | Cobra | Viper

## Docs

- [Product Requirements](docs/prd.md)
- [Technical Approach](docs/tech-approach.md)
- [Developer Guide](docs/dev-guide.md)
- [Roadmap](docs/roadmap.md)
- [UI Testing Spec](docs/ui-testing.md)
