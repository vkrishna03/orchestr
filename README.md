# Orchestr

CLI tool that automatically validates AI-generated code through functional testing with zero configuration.

## Installation

### Download

```bash
# macOS (Apple Silicon)
curl -L https://github.com/vkrishna03/orchestr/releases/download/v0.1.0/orchestr_0.1.0_darwin_arm64.tar.gz | tar xz
sudo mv orchestr /usr/local/bin/

# macOS (Intel)
curl -L https://github.com/vkrishna03/orchestr/releases/download/v0.1.0/orchestr_0.1.0_darwin_amd64.tar.gz | tar xz
sudo mv orchestr /usr/local/bin/

# Linux (amd64)
curl -L https://github.com/vkrishna03/orchestr/releases/download/v0.1.0/orchestr_0.1.0_linux_amd64.tar.gz | tar xz
sudo mv orchestr /usr/local/bin/
```

Linux arm64 and Windows builds are on the [v0.1.0 release page](https://github.com/vkrishna03/orchestr/releases/tag/v0.1.0).

### From Source

```bash
go install github.com/vkrishna03/orchestr/cmd/orchestr@latest
```

## Quick Start

```bash
# Create an Orchestrfile in your project
orchestr init

# Run tests
orchestr test .
```

## What It Does

Write an `Orchestrfile` spec, and Orchestr handles the rest:
- Detects service type (Spring Boot, Vite, etc.)
- Starts the service automatically
- Runs functional tests
- Reports pass/fail with details

```bash
$ orchestr test ./my-api

Starting my-api... ready

Running 3 test(s):

  ✓ [Health] GET /health → 200 (5ms)
  ✓ [User] POST /api/register → 201 (45ms)
  ✗ [User] GET /api/me → 200 (12ms)
    expected status 200, got 401

2/3 passed (66%) in 62ms

Stopping my-api... done
```

## Orchestrfile

Define tests in a simple DSL:

```
# Health check
ENTITY: Health
ENDPOINT: /health
METHOD: GET
EXPECT: 200

---

# Register user
ENTITY: User
ENDPOINT: /api/register
METHOD: POST
DATA: {email: "test@example.com", password: "test123", name: "Test User"}
EXPECT: 201 with {id: <uuid>, email: "test@example.com"}
SAVE: userId from body.id

---

# Login and save token
ENTITY: Auth
ENDPOINT: /api/login
METHOD: POST
DATA: {email: "test@example.com", password: "test123"}
EXPECT: 200 with {token: <jwt>}
SAVE: token from body.token

---

# Use saved token in subsequent request
ENTITY: Profile
ENDPOINT: /api/users/{{userId}}
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

**With Constraints:**
- `<array:min=1>` - Array with at least 1 element
- `<array:nonempty>` - Non-empty array
- `<int:gte=0>` - Integer >= 0
- `<int:gt=0,lt=100>` - Integer in range (0, 100)
- `<string:nonempty>` - Non-empty string
- `<string:min=3,max=50>` - String length between 3-50

**Variable References:**
- `{{varName}}` - Use saved value from previous test
- `SAVE: token from body.data.accessToken` - Extract nested values

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
orchestr test . --format markdown  # Markdown report (writes to orchestr-report.md)
orchestr test . --no-start         # Skip starting service (assume already running)
orchestr test . -v                 # Verbose output
```

## Output Formats

**Console (default):** Colored output with pass/fail indicators

**JSON:** Structured output for CI/CD integration
```bash
orchestr test . --format json
```

**Markdown:** Detailed report file with tables and failure details
```bash
orchestr test . --format markdown
# Writes to: orchestr-report.md
```

## Supported Frameworks

**v0.1 (Current):**
- Spring Boot (pom.xml, build.gradle)
- React/Vite (vite.config.ts/js)

**v1.0 (Planned):**
- Multi-service orchestration
- External API mocking (Stripe, SendGrid, SQS, etc.)
- Distributed tracing

## Development

```bash
# Build
go build -o orchestr ./cmd/orchestr

# Test
go test ./...

# Lint
golangci-lint run
```

## License

MIT

## Links

- [Roadmap](docs/roadmap.md)
- [Developer Guide](docs/dev-guide.md)
