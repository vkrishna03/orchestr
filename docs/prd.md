# Orchestr - Product Requirements Document

## Executive Summary

Orchestr is a CLI tool that automatically validates AI-generated code through functional testing with zero configuration. Built in Go for single-binary distribution and fast execution.

**Target Users:** Developers using AI code generation (Claude Code, Copilot, etc.) working with Spring Boot backends and React/Vite frontends.

**Core Problem:** Testing single services is trivial (curl works fine). The real problem is multi-service architectures where services depend on each other and external APIs. No existing tool automates this end-to-end.

---

## The Problem

### What Doesn't Work

**Single service?** Easy. Just `curl http://localhost:8080/api/test`. Claude Code can do this.

**Multi-service architecture?** Nightmare:
- Auth service on port 8081
- Payment service on port 8082  
- Notification service on port 8083
- Main API on port 8080 that calls all three
- Plus external APIs (Stripe, SendGrid)

**Manual testing requires:**
1. Start all services in correct order
2. Mock external APIs (Stripe, SendGrid)
3. Test request flows through multiple services
4. Debug failures across different service logs
5. Repeat for every code change

**Existing tools fail:**
- **Postman/Insomnia:** Manual request creation, no orchestration, no automation
- **Unit tests:** AI doesn't generate them, don't catch integration issues
- **TestContainers/etc:** Requires writing test code, complex setup per project
- **Cypress/Selenium:** UI-focused, doesn't handle backend orchestration

**Nobody solves:** "I have AI-generated multi-service code with no tests. Does it work? Find out automatically."

---

## The Solution

Orchestr orchestrates multi-service testing automatically - discovering services, managing dependencies, mocking external APIs, and tracing requests across the entire system.

### v0.1: Single Service Testing (MVP - Weeks 1-4)

Foundation for multi-service. Tests one backend at a time.

```bash
$ orchestr test ./my-api

Starting service... ✓
Running 12 functional tests...
8/12 passed, 2 critical failures

Report: orchestr-report.md
```

**Key features:**
- Detects service type (Spring Boot, Node.js, etc.)
- Starts service automatically
- Generates tests from Orchestrfile spec
- Reports functional pass/fail
- Code quality analysis
- Performance testing

### v1.0: Multi-Service Testing (Priority - Weeks 5-10)

**This is the differentiator.** Automatic orchestration of complex architectures.

```bash
$ orchestr test ./my-microservices

Discovered 4 services + 2 external APIs
Starting in dependency order...
  ✓ auth-service (8081)
  ✓ payment-service (8082)
  ✓ notification-service (8083)
  ✓ main-api (8080)

Auto-mocking external APIs...
  ✓ stripe.com → localhost:9001
  ✓ sendgrid.com → localhost:9002

Running tests with distributed tracing...

Test: "User can make payment"
[0ms] → main-api: POST /api/payments
[12ms]   → auth-service: GET /verify ✓
[25ms]   → payment-service: POST /charge
[28ms]       → stripe (MOCK): POST /charges ✓
[35ms]       → notification-service: POST /send
[38ms]           → sendgrid (MOCK): POST /mail ✓
[42ms] ✓ Success

12/15 tests passed
Report: orchestr-report.md
```

**What makes this 10x better:**

1. **Zero-config service discovery** - Scans directory, finds all services, builds dependency graph
2. **Automatic orchestration** - Starts services in correct order, manages ports
3. **Auto-generated mocks** - Detects external API calls (Stripe, etc.), creates mock servers automatically
4. **Distributed tracing** - See request flow across all services in timeline
5. **Smart failure isolation** - Pinpoints which service failed and why
6. **One command** - Everything automated, no manual setup

---

## Orchestrfile - Deterministic Test Specification

Tests are defined in `Orchestrfile` using a hybrid DSL that's human-readable but machine-parseable.

### Syntax

```
# Orchestrfile - Test specifications

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

ENTITY: Payment
ENDPOINT: /api/payments
METHOD: POST
AUTH: Bearer {{token}}
DATA: {amount: 5000, currency: "usd", card: {number: "4242424242424242"}}
EXPECT: 200 with {id: <uuid>, status: "succeeded"}
EXPECT_CALLS: auth-service:/verify, payment-service:/charge, stripe:/v1/charges
```

### Built-in Type Patterns

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
- `{{var}}` - Variable substitution from saved values

### Keywords

- `ENTITY:` - Logical entity being tested (User, Payment, Order)
- `ENDPOINT:` - API endpoint path
- `METHOD:` - HTTP method (GET, POST, PUT, DELETE, PATCH)
- `DATA:` - Request body (JSON)
- `AUTH:` - Authentication header
- `HEADERS:` - Additional headers
- `EXPECT:` - Expected status code and response body
- `EXPECT_CALLS:` - Expected service-to-service calls
- `SAVE:` - Save value from response for later use
- `---` - Test separator

### Why This Format

**Deterministic:**
- No ambiguity about what to test
- Every field is explicit
- Parseable with simple regex

**Readable:**
- Looks like structured English
- Self-documenting
- Easy to write and maintain

**Flexible:**
- Variables for test dependencies
- Type patterns for validation
- Can add new keywords without breaking existing specs

---

## Technical Architecture

### Why Go

- **Single binary** - No runtime dependencies, easy distribution
- **Fast execution** - Critical for test performance
- **Great concurrency** - Managing multiple services, parallel tests
- **Excellent Docker SDK** - Container orchestration
- **Cross-compilation** - Build for all platforms easily
- **Low memory footprint** - Important when running test environments

### Stack

**Core:** Go 1.21+

**Key Libraries:**
- `cobra` - CLI framework
- `docker/docker` - Container management  
- `chromedp` - Browser automation (for React frontends)
- `go-yaml` - Config parsing

**Service Support (v0.1):**
- Spring Boot (detect via pom.xml/build.gradle)
- Node.js/Express (detect via package.json)
- Python/Flask/FastAPI (detect via requirements.txt)

**Frontend Support (v1.0):**
- React + Vite + TypeScript/JavaScript
- Auto-detect via vite.config.ts

### Orchestrfile Parser

```go
package parser

import (
    "regexp"
    "strings"
)

type TestSpec struct {
    Entity       string
    Endpoint     string
    Method       string
    Auth         string
    Headers      map[string]string
    Data         string // JSON string
    ExpectStatus int
    ExpectBody   string // JSON string with patterns
    ExpectCalls  []string
    Save         map[string]string
}

var (
    testSeparator = regexp.MustCompile(`(?m)^---\s*$`)
    entityRegex   = regexp.MustCompile(`(?m)^ENTITY:\s*(.+)$`)
    endpointRegex = regexp.MustCompile(`(?m)^ENDPOINT:\s*(.+)$`)
    methodRegex   = regexp.MustCompile(`(?m)^METHOD:\s*(.+)$`)
    authRegex     = regexp.MustCompile(`(?m)^AUTH:\s*(.+)$`)
    dataRegex     = regexp.MustCompile(`(?m)^DATA:\s*(\{.+\})$`)
    expectRegex   = regexp.MustCompile(`(?m)^EXPECT:\s*(\d+)\s+with\s+(\{.+\})$`)
    callsRegex    = regexp.MustCompile(`(?m)^EXPECT_CALLS:\s*(.+)$`)
    saveRegex     = regexp.MustCompile(`(?m)^SAVE:\s*(\w+)\s+from\s+(.+)$`)
)

func Parse(input string) ([]TestSpec, error) {
    tests := testSeparator.Split(input, -1)
    specs := make([]TestSpec, 0, len(tests))
    
    for _, test := range tests {
        test = strings.TrimSpace(test)
        if test == "" {
            continue
        }
        
        spec := TestSpec{
            Save: make(map[string]string),
            Headers: make(map[string]string),
        }
        
        // Parse each field with regex
        if match := entityRegex.FindStringSubmatch(test); len(match) > 1 {
            spec.Entity = strings.TrimSpace(match[1])
        }
        
        if match := endpointRegex.FindStringSubmatch(test); len(match) > 1 {
            spec.Endpoint = strings.TrimSpace(match[1])
        }
        
        // ... parse other fields
        
        specs = append(specs, spec)
    }
    
    return specs, nil
}

// Validate spec before execution
func Validate(spec TestSpec) error {
    if spec.Endpoint == "" {
        return errors.New("ENDPOINT is required")
    }
    
    validMethods := []string{"GET", "POST", "PUT", "DELETE", "PATCH"}
    if !contains(validMethods, spec.Method) {
        return fmt.Errorf("invalid METHOD: %s", spec.Method)
    }
    
    return nil
}
```

### Service Discovery Algorithm

```go
// Scan directory tree
services := []Service{}
Walk(projectDir, func(path string) {
    if exists("pom.xml") {
        port := parseApplicationYml("application.yml")
        deps := parseJavaCodeForHTTPCalls() // Find RestTemplate calls
        services = append(services, SpringBootService{port, deps})
    }
    if exists("vite.config.ts") {
        services = append(services, ViteApp{parsePort()})
    }
})

// Build dependency graph
graph := buildDependencyGraph(services)
startOrder := topologicalSort(graph)
```

### Auto-Mock Generation

```go
// Parse code for external API calls
externalAPIs := findExternalAPIs("./src")
// Found: restTemplate.post("https://api.stripe.com/v1/charges")

for api := range externalAPIs {
    mockServer := httptest.NewServer(http.HandlerFunc(func(w, r *http.Request) {
        // Return realistic mock response based on endpoint
        if strings.Contains(r.URL.Path, "/charges") {
            w.WriteHeader(200)
            w.Write([]byte(`{"id":"ch_mock","status":"succeeded"}`))
        }
    }))
    
    // Inject mock URL via env var
    env["STRIPE_API_URL"] = mockServer.URL
}
```

### Distributed Tracing

```go
// Inject trace ID into all requests
traceID := uuid.New()
req.Header.Set("X-Trace-ID", traceID)

// Log every service call with timestamp
trace := TraceLog{
    TraceID: traceID,
    Events: []TraceEvent{},
}

// Capture: timestamp, service, method, path, duration, status
trace.Events = append(trace.Events, TraceEvent{
    Time: time.Now(),
    Service: "main-api",
    Method: "POST",
    Path: "/api/payments",
    Duration: 42 * time.Millisecond,
    Status: 200,
})
```

---

## Product Requirements

### v0.1: Single Service (Weeks 1-4)

**Priority 0: Functional Testing**

- Parse Orchestrfile into test cases
- Detect service type (Spring Boot, Node.js, Python)
- Start service automatically
- Execute HTTP tests against running service
- Test graceful error handling (bad input, network failures)
- Generate markdown report with pass/fail

**Priority 1: Code Quality**

- Detect architectural patterns
- Find code duplication
- Check error handling consistency
- Identify AI-specific issues (hallucinated imports, incomplete implementations)

**Priority 2: Performance**

- Load testing (10, 50, 100, 500 concurrent users)
- Response time measurement
- Memory leak detection
- Breaking point identification

**Success Criteria:**
- Tests 80% of single-service CRUD apps without config
- Identifies crashes with 95% accuracy
- Test run completes in under 3 minutes
- 10 beta users successfully test AI code

### v1.0: Multi-Service (Weeks 5-10)

**Priority 0: Service Orchestration**

- Auto-discover all services in project directory
- Build dependency graph (which service depends on which)
- Start services in topologically sorted order
- Health check each service before testing
- Manage port allocation automatically
- Clean shutdown of all services

**Priority 1: External API Mocking**

- Parse code to detect external API calls (Stripe, SendGrid, AWS, etc.)
- Generate mock HTTP servers automatically
- Return realistic responses based on endpoint patterns
- Support mock configuration for custom responses
- Inject mock URLs via environment variables

**Priority 2: Distributed Testing**

- Execute tests that span multiple services
- Trace requests across service boundaries
- Capture timing for each service hop
- Identify bottlenecks in service chain
- Report on inter-service communication
- Validate EXPECT_CALLS in Orchestrfile

**Priority 3: Failure Isolation**

- When test fails, identify which service caused it
- Extract error from service logs
- Suggest likely fix (missing env var, wrong config, etc.)
- Show full request trace leading to failure

**Success Criteria:**
- Handles 90% of 2-5 service architectures without config
- Auto-mocks common APIs (Stripe, SendGrid, Twilio)
- Distributed trace shows complete request flow
- Root cause identification accuracy >80%
- 50+ projects using in production

---

## User Experience

### Installation

```bash
# Download single binary
curl -L https://github.com/orchestr/orchestr/releases/latest/download/orchestr-linux -o orchestr
chmod +x orchestr
sudo mv orchestr /usr/local/bin/

# Or via package manager (future)
brew install orchestr
```

### Basic Usage (v0.1)

```bash
# Test with default Orchestrfile
orchestr test .

# Custom Orchestrfile location
orchestr test . --spec ./tests/Orchestrfile

# JSON output for CI
orchestr test . --format json

# Verbose mode
orchestr test . -v
```

### Multi-Service Usage (v1.0)

```bash
# Auto-discover and test all services
orchestr test .

# Test with real external APIs (no mocking)
orchestr test . --no-mock

# Mock only specific APIs
orchestr test . --mock stripe,sendgrid

# Use real service, stub others
orchestr test . --test-service payment-service --stub auth-service,notification-service
```

### Project Structure

```
my-project/
├── Orchestrfile          # Test specifications
├── Orchestrfile.auth     # Auth-specific tests (optional)
├── Orchestrfile.payments # Payment-specific tests (optional)
├── .orchestr.yml         # Tool configuration (optional)
├── auth-service/
├── payment-service/
└── main-api/
```

### Configuration File (Optional)

```yaml
# .orchestr.yml
services:
  - name: auth-service
    path: ./auth-service
    port: 8081
    
  - name: payment-service
    path: ./payment-service
    port: 8082
    depends_on: [auth-service]
    
  - name: main-api
    path: ./main-api
    port: 8080
    depends_on: [auth-service, payment-service]

mocks:
  stripe:
    base_url: https://api.stripe.com
    scenarios:
      - endpoint: /v1/charges
        response: {id: "ch_test", status: "succeeded"}

spec:
  file: Orchestrfile
```

---

## Example Report (Multi-Service)

```markdown
# Orchestr Test Report
Generated: 2024-12-21 14:30

## Summary
- Services: 4 (all healthy)
- External APIs: 2 (mocked)
- Tests: 7/10 passed (70%)
- Critical failures: 2

## Service Health
✓ auth-service (8081) - 45ms avg response
✓ payment-service (8082) - 120ms avg response  
✓ notification-service (8083) - 32ms avg response
✓ main-api (8080) - 180ms avg response

## Test Results

### ✓ User Registration
- POST /api/register → 201 Created (23ms)
- Email validation works correctly

### ✗ User Payment (CRITICAL)
Timeline:
[0ms] → main-api: POST /api/payments
[8ms]   → auth-service: GET /verify ✓ (8ms)
[15ms]  → payment-service: POST /charge
[18ms]      ✗ NullPointerException at PaymentController.java:45

Root Cause: payment-service expects STRIPE_API_KEY env var but not set
Fix: Add STRIPE_API_KEY to payment-service configuration

### ⚠ Email Confirmation
- notification-service returned 500
- Email sent but sendgrid mock returned error
- Non-critical, payment succeeded
```

---

## Competitive Differentiation

| Capability | Postman | TestContainers | Orchestr |
|-----------|---------|---------------|----------|
| Multi-service orchestration | ❌ Manual | ❌ Manual | ✅ Automatic |
| Auto-mock external APIs | ❌ | ❌ | ✅ |
| Test generation from spec | ❌ | ❌ | ✅ |
| Distributed tracing | ❌ | ❌ | ✅ |
| Root cause analysis | ❌ | ❌ | ✅ |
| Setup time | 1-2 hours | 2-4 hours | 30 seconds |
| Test authoring | Manual | Manual code | Orchestrfile DSL |

**The 10x difference:** One command replaces hours of manual setup and test writing.

---

## Timeline

### v0.1 (Weeks 1-4) - Single Service MVP
- Week 1: CLI + service detection + Orchestrfile parser
- Week 2: Test generation + execution
- Week 3: Reporting + code quality
- Week 4: Performance testing + polish

### v1.0 (Weeks 5-10) - Multi-Service (Priority)
- Week 5-6: Service discovery + dependency graph
- Week 7-8: Auto-mock generation + orchestration
- Week 9: Distributed tracing
- Week 10: Failure isolation + polish

### v1.1 (Week 11-12) - Production Ready
- Security testing
- Documentation analysis
- CI/CD integration
- Public launch

---

## Success Metrics

**v0.1:**
- 10 beta users test AI-generated single services
- 80% test success rate for CRUD apps
- Average test time under 3 minutes

**v1.0:**
- 50 projects using multi-service testing
- Supports 90% of 2-5 service architectures
- Auto-mocks top 10 external APIs
- Average test time under 5 minutes for multi-service

**v1.1:**
- 100+ GitHub stars
- Used in 100+ projects
- 90% user satisfaction
- Featured in AI coding tool communities

---

## Risks & Mitigation

**Risk:** Service auto-discovery fails for complex setups
**Mitigation:** Support manual .orchestr.yml config, clear error messages

**Risk:** Mock responses don't match real API behavior
**Mitigation:** Use OpenAPI specs when available, allow custom mock configs

**Risk:** Tests too slow with many services
**Mitigation:** Parallel service startup, cached builds, incremental testing

**Risk:** False positives in failure detection
**Mitigation:** Conservative failure criteria, show evidence, user feedback loop

---

## Out of Scope

- ❌ Fixing code automatically
- ❌ Generating test suites for users
- ❌ Real-time monitoring
- ❌ Cloud-hosted service
- ❌ IDE plugins (CLI-first)
- ❌ Compiled languages initially (C++, Rust, Java native)

---

## Branding

**Name:** Orchestr (orchestrator without vowels - modern tech branding)

**Tagline:** "Orchestrate your multi-service tests"

**Logo Concept:** Conductor's baton coordinating services

**Community:**
- GitHub: github.com/orchestr/orchestr
- Website: orchestr.dev
- Docs: docs.orchestr.dev
- Discord: discord.gg/orchestr

---

## Version History

- **v0.1** - Initial PRD (2024-12-21)
- **v0.2** - Added Go rationale, multi-service focus, competitive analysis (2024-12-21)
- **v0.3** - Project renamed to Orchestr, Orchestrfile DSL specification added (2024-12-21)
