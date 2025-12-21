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

- `<uuid>` - Valid UUID format
- `<jwt>` - Valid JWT token
- `<int>` - Integer
- `<string>` - Any non-empty string
- `<email>` - Valid email format
- `<iso8601>` - ISO 8601 datetime
- `{{var}}` - Variable from saved values

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
