# Orchestr Developer Guide

> Practical guide for developing Orchestr. No fluff.

---

## Setup

### Requirements

- **Go 1.21+** (for `log/slog` stdlib)
- **Make** (for build commands)
- **golangci-lint** (for linting)
- **goreleaser** (for releases, optional)

### Install Tools

```bash
# Install golangci-lint
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Install goreleaser (optional, for releases)
go install github.com/goreleaser/goreleaser@latest
```

### Build & Run

```bash
# Build binary
go build -o orchestr ./cmd/orchestr

# Run
./orchestr test ./path/to/project

# Or during development
go run ./cmd/orchestr test ./path/to/project
```

### Test

```bash
# Run all tests
go test ./...

# Run with coverage
go test -cover ./...

# Run specific package
go test ./internal/parser/...

# Verbose output
go test -v ./internal/parser/...
```

### Lint

```bash
golangci-lint run
```

---

## Git Workflow

### Commit Messages

Use [Conventional Commits](https://www.conventionalcommits.org/):

```
<type>: <description>

[optional body]
```

**Types:**

| Type | Description |
|------|-------------|
| `feat` | New feature |
| `fix` | Bug fix |
| `docs` | Documentation only |
| `test` | Adding/updating tests |
| `refactor` | Code change that neither fixes a bug nor adds a feature |
| `perf` | Performance improvement |
| `style` | Code style (formatting, missing semicolons, etc.) |
| `build` | Build system or external dependencies |
| `ci` | CI/CD configuration |
| `chore` | Other changes that don't modify src or test files |

**Examples:**

```bash
feat: add Spring Boot service detector
fix: handle missing Orchestrfile gracefully
docs: update README with usage examples
test: add parser unit tests
refactor: extract HTTP client into separate package
chore: update dependencies
```

### Branching

```bash
# Feature branch
git checkout -b feat/service-discovery

# Bug fix branch
git checkout -b fix/health-check-timeout

# Docs branch
git checkout -b docs/api-reference
```

---

## Package Structure

```
orchestr/
├── cmd/orchestr/           # Entry point only
│   └── main.go
├── internal/               # Private packages
│   ├── cli/                # Cobra commands
│   ├── config/             # Viper config loading
│   ├── parser/             # Orchestrfile parser
│   ├── discovery/          # Service detection
│   ├── runner/             # Service lifecycle
│   ├── mock/               # Mock HTTP servers
│   ├── executor/           # Test execution
│   ├── tracer/             # Distributed tracing
│   ├── reporter/           # Report generation
│   └── models/             # Shared types
├── pkg/                    # Public APIs (if any)
└── templates/              # Orchestrfile templates
```

### Package Responsibilities

| Package | Does | Does NOT |
|---------|------|----------|
| `cli` | Parse args, wire components, handle signals | Business logic |
| `parser` | Parse Orchestrfile, validate specs | Execute tests |
| `discovery` | Find services, build dep graph | Start services |
| `runner` | Start/stop processes, health checks | Detect service types |
| `executor` | Make HTTP requests, validate responses | Generate reports |
| `reporter` | Format output | Execute tests |

---

## Naming Conventions

### Files

```
internal/discovery/
├── discovery.go       # Main logic, exported types
├── springboot.go      # Spring Boot detector
├── vite.go            # Vite detector
├── graph.go           # Dependency graph
└── discovery_test.go  # Tests
```

### Types

```go
// Interfaces: verb + er
type ServiceDetector interface { ... }
type ServiceRunner interface { ... }

// Structs: noun
type Service struct { ... }
type TestSpec struct { ... }

// Implementations: noun (matches interface purpose)
type SpringBootDetector struct { ... }
type NativeRunner struct { ... }
```

### Functions

```go
// Constructors: New + Type
func NewParser() *Parser { ... }
func NewSpringBootDetector() *SpringBootDetector { ... }

// Methods: verb
func (p *Parser) Parse(input string) ([]TestSpec, error) { ... }
func (r *NativeRunner) Start(ctx context.Context, svc Service) error { ... }

// Private helpers: lowercase, descriptive
func parseEndpoint(line string) (string, error) { ... }
func buildDependencyGraph(services []Service) *Graph { ... }
```

### Variables

```go
// Prefer short names in small scopes
for _, svc := range services { ... }
for i, spec := range specs { ... }

// Descriptive names for longer-lived variables
healthCheckTimeout := 30 * time.Second
dependencyGraph := buildGraph(services)
```

---

## Error Handling

### Wrap Errors with Context

```go
import "fmt"

func (d *SpringBootDetector) Detect(path string) (*Service, error) {
    port, err := parseApplicationYml(path)
    if err != nil {
        return nil, fmt.Errorf("detect spring boot in %s: %w", path, err)
    }
    return &Service{Port: port}, nil
}
```

### Custom Error Types

```go
// internal/parser/errors.go

type ParseError struct {
    Line    int
    Column  int
    Message string
}

func (e *ParseError) Error() string {
    return fmt.Sprintf("line %d, col %d: %s", e.Line, e.Column, e.Message)
}

// Usage
return nil, &ParseError{
    Line:    p.lineNum,
    Column:  pos,
    Message: "invalid METHOD, expected GET|POST|PUT|DELETE|PATCH",
}
```

### Sentinel Errors

```go
// internal/discovery/errors.go

var (
    ErrNotSpringBoot = errors.New("not a Spring Boot project")
    ErrNotVite       = errors.New("not a Vite project")
    ErrNoService     = errors.New("no service detected")
)

// Usage
if errors.Is(err, ErrNotSpringBoot) {
    // Try next detector
}
```

### Error Checking Pattern

```go
// DO: Check errors immediately
result, err := doSomething()
if err != nil {
    return fmt.Errorf("do something: %w", err)
}

// DON'T: Defer error checking
result, err := doSomething()
// ... other code ...
if err != nil { } // Too late
```

---

## Testing Strategy

### Unit Tests

**What to Test:**
- Parser handles all valid Orchestrfile syntax
- Pattern matchers validate correctly
- Service detectors find correct markers
- Response validators match patterns

**Table-Driven Tests:**

```go
func TestPatternMatcher(t *testing.T) {
    tests := []struct {
        name     string
        pattern  string
        input    string
        expected bool
    }{
        {"uuid valid", "<uuid>", "550e8400-e29b-41d4-a716-446655440000", true},
        {"uuid invalid", "<uuid>", "not-a-uuid", false},
        {"jwt valid", "<jwt>", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.dozjgNryP4J3jVmNHl0w5N_XgL0n3I9PlFUP0THsR8U", true},
        {"int positive", "<int>", "42", true},
        {"int negative", "<int>", "-42", true},
        {"int float", "<int>", "42.5", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := Match(tt.pattern, tt.input)
            assert.Equal(t, tt.expected, result)
        })
    }
}
```

**Parser Tests:**

```go
func TestParseOrchestfile(t *testing.T) {
    input := `
ENTITY: User
ENDPOINT: /api/users
METHOD: POST
DATA: {"name": "test"}
EXPECT: 201 with {id: <uuid>}
`
    specs, err := Parse(input)
    require.NoError(t, err)
    require.Len(t, specs, 1)

    assert.Equal(t, "User", specs[0].Entity)
    assert.Equal(t, "/api/users", specs[0].Endpoint)
    assert.Equal(t, "POST", specs[0].Method)
    assert.Equal(t, 201, specs[0].ExpectStatus)
}
```

### Integration Tests

**Service Detection:**

```go
func TestSpringBootDetector(t *testing.T) {
    // Create temp directory with pom.xml
    dir := t.TempDir()
    writePomXml(t, dir)
    writeApplicationYml(t, dir, 8080)

    detector := NewSpringBootDetector()
    svc, err := detector.Detect(dir)

    require.NoError(t, err)
    assert.Equal(t, ServiceTypeSpringBoot, svc.Type)
    assert.Equal(t, 8080, svc.Port)
}
```

**Process Management:**

```go
func TestNativeRunner(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test")
    }

    runner := NewNativeRunner()
    svc := Service{
        Name:         "test-server",
        StartCommand: []string{"python", "-m", "http.server", "8888"},
        Port:         8888,
        HealthPath:   "/",
    }

    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    err := runner.Start(ctx, svc)
    require.NoError(t, err)

    defer runner.Stop(ctx, svc)

    err = runner.HealthCheck(ctx, svc)
    assert.NoError(t, err)
}
```

### E2E Tests

**Full Test Run:**

```go
func TestE2E_SingleService(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping E2E test")
    }

    // Setup: sample Spring Boot project with Orchestrfile
    projectDir := setupTestProject(t)

    // Run orchestr test
    cmd := exec.Command("./orchestr", "test", projectDir)
    output, err := cmd.CombinedOutput()

    require.NoError(t, err, "output: %s", output)
    assert.Contains(t, string(output), "passed")
}
```

### Test File Organization

```
internal/parser/
├── parser.go
├── parser_test.go      # Unit tests
├── patterns.go
├── patterns_test.go
└── testdata/           # Test fixtures
    ├── valid.orchestr
    ├── invalid_method.orchestr
    └── missing_endpoint.orchestr
```

---

## Adding New Service Detectors

### Step 1: Implement the Interface

```go
// internal/discovery/fastapi.go

package discovery

type FastAPIDetector struct{}

func NewFastAPIDetector() *FastAPIDetector {
    return &FastAPIDetector{}
}

func (d *FastAPIDetector) Detect(path string) (*Service, error) {
    // Check for requirements.txt with fastapi
    reqPath := filepath.Join(path, "requirements.txt")
    content, err := os.ReadFile(reqPath)
    if err != nil {
        return nil, ErrNotFastAPI
    }

    if !strings.Contains(string(content), "fastapi") {
        return nil, ErrNotFastAPI
    }

    // Find port from main.py or .env
    port := d.findPort(path)

    return &Service{
        Name:         filepath.Base(path),
        Type:         ServiceTypeFastAPI,
        Path:         path,
        Port:         port,
        StartCommand: []string{"uvicorn", "main:app", "--port", strconv.Itoa(port)},
        HealthPath:   "/health",
    }, nil
}

func (d *FastAPIDetector) Type() ServiceType {
    return ServiceTypeFastAPI
}

func (d *FastAPIDetector) findPort(path string) int {
    // Check .env for PORT
    // Check main.py for uvicorn.run(... port=...)
    // Default: 8000
    return 8000
}
```

### Step 2: Add Service Type

```go
// internal/models/service.go

type ServiceType string

const (
    ServiceTypeSpringBoot ServiceType = "springboot"
    ServiceTypeVite       ServiceType = "vite"
    ServiceTypeFastAPI    ServiceType = "fastapi"  // Add this
)
```

### Step 3: Register the Detector

```go
// internal/discovery/discovery.go

func NewDiscovery() *Discovery {
    return &Discovery{
        detectors: []ServiceDetector{
            NewSpringBootDetector(),
            NewViteDetector(),
            NewFastAPIDetector(),  // Add this
        },
    }
}
```

### Step 4: Add Tests

```go
// internal/discovery/fastapi_test.go

func TestFastAPIDetector(t *testing.T) {
    tests := []struct {
        name        string
        setup       func(t *testing.T, dir string)
        expectError error
    }{
        {
            name: "valid fastapi project",
            setup: func(t *testing.T, dir string) {
                writeFile(t, dir, "requirements.txt", "fastapi==0.100.0\nuvicorn")
                writeFile(t, dir, "main.py", "from fastapi import FastAPI\napp = FastAPI()")
            },
            expectError: nil,
        },
        {
            name: "not a fastapi project",
            setup: func(t *testing.T, dir string) {
                writeFile(t, dir, "requirements.txt", "django")
            },
            expectError: ErrNotFastAPI,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            dir := t.TempDir()
            tt.setup(t, dir)

            detector := NewFastAPIDetector()
            _, err := detector.Detect(dir)

            if tt.expectError != nil {
                assert.ErrorIs(t, err, tt.expectError)
            } else {
                assert.NoError(t, err)
            }
        })
    }
}
```

---

## Adding New Mock Templates

### Step 1: Define the Template

```go
// internal/mock/templates.go

var MockTemplates = map[string]MockTemplate{
    "stripe": stripeTemplate,
    "sendgrid": sendgridTemplate,
    "twilio": twilioTemplate,  // Add this
}

var twilioTemplate = MockTemplate{
    Name:    "twilio",
    BaseURL: "api.twilio.com",
    Routes: []MockRoute{
        {
            Method: "POST",
            Path:   "/2010-04-01/Accounts/:accountSid/Messages.json",
            Response: Response{
                Status: 201,
                Body: `{
                    "sid": "SM{{uuid}}",
                    "status": "queued",
                    "to": "{{.To}}",
                    "from": "{{.From}}",
                    "body": "{{.Body}}"
                }`,
            },
        },
        {
            Method: "GET",
            Path:   "/2010-04-01/Accounts/:accountSid/Messages/:sid.json",
            Response: Response{
                Status: 200,
                Body: `{
                    "sid": "{{.sid}}",
                    "status": "delivered"
                }`,
            },
        },
    },
}
```

### Step 2: Template Format Reference

```go
type MockTemplate struct {
    Name    string      // Template identifier
    BaseURL string      // Domain to mock (api.twilio.com)
    Routes  []MockRoute // Endpoint definitions
}

type MockRoute struct {
    Method   string   // HTTP method
    Path     string   // Path with params (:id, :accountSid)
    Response Response // Mock response
}

type Response struct {
    Status  int               // HTTP status code
    Body    string            // JSON body with {{placeholders}}
    Headers map[string]string // Optional headers
}
```

### Step 3: Template Placeholders

| Placeholder | Description | Example |
|-------------|-------------|---------|
| `{{uuid}}` | Random UUID | `550e8400-e29b-41d4-...` |
| `{{timestamp}}` | ISO8601 now | `2024-12-21T14:30:00Z` |
| `{{.fieldName}}` | From request body | `{{.To}}` → request body's `to` field |
| `{{.pathParam}}` | From URL path | `{{.sid}}` → `:sid` in path |

### Step 4: Add Detection Pattern

```go
// internal/mock/detector.go

var knownAPIs = map[string]string{
    "api.stripe.com":   "stripe",
    "api.sendgrid.com": "sendgrid",
    "api.twilio.com":   "twilio",  // Add this
}
```

---

## CI/CD

### GitHub Actions Workflow

```yaml
# .github/workflows/ci.yml

name: CI

on:
  push:
    branches: [main]
  pull_request:
    branches: [main]

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Lint
        uses: golangci/golangci-lint-action@v3
        with:
          version: latest

      - name: Test
        run: go test -race -coverprofile=coverage.out ./...

      - name: Upload coverage
        uses: codecov/codecov-action@v3
        with:
          files: ./coverage.out

  build:
    runs-on: ubuntu-latest
    needs: test
    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Build
        run: go build -o orchestr ./cmd/orchestr

      - uses: actions/upload-artifact@v3
        with:
          name: orchestr-linux
          path: orchestr
```

### Release Workflow

```yaml
# .github/workflows/release.yml

name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - uses: actions/setup-go@v5
        with:
          go-version: '1.21'

      - name: Run GoReleaser
        uses: goreleaser/goreleaser-action@v5
        with:
          version: latest
          args: release --clean
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
```

### GoReleaser Config

```yaml
# .goreleaser.yaml

builds:
  - main: ./cmd/orchestr
    binary: orchestr
    goos:
      - linux
      - darwin
      - windows
    goarch:
      - amd64
      - arm64

archives:
  - format: tar.gz
    name_template: "{{ .ProjectName }}_{{ .Version }}_{{ .Os }}_{{ .Arch }}"
    format_overrides:
      - goos: windows
        format: zip

checksum:
  name_template: 'checksums.txt'

release:
  github:
    owner: orchestr
    name: orchestr
```

---

## Common Patterns

### Context for Cancellation

```go
func (r *Runner) Start(ctx context.Context, svc Service) error {
    cmd := exec.CommandContext(ctx, svc.StartCommand[0], svc.StartCommand[1:]...)

    if err := cmd.Start(); err != nil {
        return err
    }

    // Wait for context cancellation or process exit
    go func() {
        <-ctx.Done()
        cmd.Process.Signal(syscall.SIGTERM)
    }()

    return nil
}
```

### Graceful Shutdown

```go
func main() {
    ctx, cancel := context.WithCancel(context.Background())

    // Handle signals
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    go func() {
        <-sigCh
        slog.Info("shutting down...")
        cancel()
    }()

    if err := run(ctx); err != nil {
        slog.Error("failed", "error", err)
        os.Exit(1)
    }
}
```

### Logging with slog

```go
import "log/slog"

// Setup
func initLogger(verbose bool) {
    level := slog.LevelInfo
    if verbose {
        level = slog.LevelDebug
    }

    handler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: level,
    })
    slog.SetDefault(slog.New(handler))
}

// Usage
slog.Info("starting service",
    "name", svc.Name,
    "port", svc.Port,
)

slog.Debug("health check",
    "attempt", i,
    "status", resp.StatusCode,
)

slog.Error("service failed",
    "name", svc.Name,
    "error", err,
)
```

### Configuration with Viper

```go
import "github.com/spf13/viper"

type Config struct {
    Timeout  time.Duration `mapstructure:"timeout"`
    Parallel int           `mapstructure:"parallel"`
    Verbose  bool          `mapstructure:"verbose"`
}

func LoadConfig() (*Config, error) {
    viper.SetConfigName(".orchestr")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")

    viper.SetEnvPrefix("ORCHESTR")
    viper.AutomaticEnv()

    viper.SetDefault("timeout", "30s")
    viper.SetDefault("parallel", 1)

    if err := viper.ReadInConfig(); err != nil {
        if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
            return nil, err
        }
    }

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }

    return &config, nil
}
```

---

## Security Considerations

### Running Untrusted Services

```go
// DON'T: Run arbitrary commands
cmd := exec.Command(userInput)

// DO: Validate and restrict commands
allowedCommands := map[string][]string{
    "springboot-maven":  {"mvn", "spring-boot:run"},
    "springboot-gradle": {"./gradlew", "bootRun"},
    "vite":              {"npm", "run", "dev"},
}

cmdArgs, ok := allowedCommands[svc.Type]
if !ok {
    return fmt.Errorf("unsupported service type: %s", svc.Type)
}
cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
```

### Secrets Handling

```go
// DON'T: Log secrets
slog.Info("config loaded", "apiKey", config.APIKey)

// DO: Mask or omit secrets
slog.Info("config loaded", "apiKey", "***")

// DON'T: Include secrets in error messages
return fmt.Errorf("auth failed with key %s", apiKey)

// DO: Omit secrets from errors
return fmt.Errorf("auth failed: invalid API key")
```

### Port Allocation

```go
// DON'T: Use hardcoded ports
port := 9000

// DO: Find available ports
func findAvailablePort() (int, error) {
    listener, err := net.Listen("tcp", ":0")
    if err != nil {
        return 0, err
    }
    defer listener.Close()
    return listener.Addr().(*net.TCPAddr).Port, nil
}

// DO: Check for conflicts
func isPortAvailable(port int) bool {
    listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
    if err != nil {
        return false
    }
    listener.Close()
    return true
}
```

### Path Traversal

```go
// DON'T: Accept arbitrary paths
configPath := filepath.Join(projectDir, userInput)

// DO: Validate paths stay within project
func safePath(base, requested string) (string, error) {
    absBase, _ := filepath.Abs(base)
    absRequested, _ := filepath.Abs(filepath.Join(base, requested))

    if !strings.HasPrefix(absRequested, absBase) {
        return "", fmt.Errorf("path traversal detected: %s", requested)
    }
    return absRequested, nil
}
```

---

## Quick Reference

### Key Libraries

| Purpose | Library | Import |
|---------|---------|--------|
| CLI | cobra | `github.com/spf13/cobra` |
| Config | viper | `github.com/spf13/viper` |
| YAML | yaml.v3 | `gopkg.in/yaml.v3` |
| Testing | testify | `github.com/stretchr/testify` |
| Concurrency | errgroup | `golang.org/x/sync/errgroup` |

### Common Commands

```bash
# Dev workflow
go run ./cmd/orchestr test ./sample-project
go test ./...
golangci-lint run

# Build
go build -o orchestr ./cmd/orchestr

# Release
git tag v0.1.0
git push --tags
```

### Makefile

```makefile
.PHONY: build test lint clean

build:
	go build -o orchestr ./cmd/orchestr

test:
	go test -race ./...

lint:
	golangci-lint run

clean:
	rm -f orchestr

run:
	go run ./cmd/orchestr $(ARGS)
```
