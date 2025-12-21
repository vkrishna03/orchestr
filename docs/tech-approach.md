# Orchestr - Technical Approach

> System design, library choices, and implementation strategy

---

## Table of Contents

1. [High-Level Architecture](#high-level-architecture)
2. [Core Components](#core-components)
3. [Component Deep Dives](#component-deep-dives)
4. [Data Flow](#data-flow)
5. [Library Choices](#library-choices)
6. [Implementation Phases](#implementation-phases)
7. [Technical Decisions](#technical-decisions)

---

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                              ORCHESTR CLI                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐    │
│  │   CLI Layer  │  │    Parser    │  │   Executor   │  │   Reporter   │    │
│  │   (Cobra)    │──│ (Orchestrfile│──│  (Test Run)  │──│  (Output)    │    │
│  └──────────────┘  └──────────────┘  └──────────────┘  └──────────────┘    │
│         │                                   │                               │
│         ▼                                   ▼                               │
│  ┌──────────────┐                    ┌──────────────┐                       │
│  │   Config     │                    │   Service    │                       │
│  │   Loader     │                    │   Manager    │                       │
│  └──────────────┘                    └──────────────┘                       │
│                                             │                               │
│                    ┌────────────────────────┼────────────────────────┐      │
│                    ▼                        ▼                        ▼      │
│             ┌──────────────┐         ┌──────────────┐         ┌──────────┐  │
│             │   Service    │         │    Mock      │         │  Tracer  │  │
│             │   Discovery  │         │   Server     │         │          │  │
│             └──────────────┘         └──────────────┘         └──────────┘  │
│                    │                        │                        │      │
│                    ▼                        ▼                        ▼      │
│             ┌──────────────┐         ┌──────────────┐         ┌──────────┐  │
│             │   Process    │         │   HTTP       │         │   Log    │  │
│             │   Manager    │         │   Intercept  │         │  Capture │  │
│             └──────────────┘         └──────────────┘         └──────────┘  │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
                                      │
                    ┌─────────────────┼─────────────────┐
                    ▼                 ▼                 ▼
             ┌──────────┐      ┌──────────┐      ┌──────────┐
             │ Service  │      │ Service  │      │  Mock    │
             │    A     │      │    B     │      │  Server  │
             │ (8081)   │      │ (8082)   │      │ (9001)   │
             └──────────┘      └──────────┘      └──────────┘
```

---

## Core Components

### 1. CLI Layer
Entry point, command parsing, user interaction.

### 2. Parser
Orchestrfile DSL parsing and validation.

### 3. Service Discovery
Detect service types, ports, dependencies.

### 4. Service Manager
Start/stop/health-check services.

### 5. Mock Server
Auto-generate mocks for external APIs.

### 6. Test Executor
Run tests, capture responses, validate expectations.

### 7. Tracer
Distributed tracing across services.

### 8. Reporter
Generate reports in various formats.

---

## Component Deep Dives

### 1. CLI Layer

```
orchestr
├── test [path]           # Run tests
├── init                  # Generate Orchestrfile template
├── discover [path]       # Show discovered services
├── mock [api]            # Start mock server standalone
└── version               # Show version
```

**Responsibility:**
- Parse CLI arguments
- Load configuration
- Orchestrate component execution
- Handle signals (graceful shutdown)

---

### 2. Orchestrfile Parser

```
┌─────────────────────────────────────────────────────────────┐
│                     PARSER PIPELINE                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Raw Text ──► Lexer ──► Token Stream ──► Parser ──► AST     │
│                                                              │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐  │
│  │ ENTITY: │    │ Token   │    │ Parse   │    │TestSpec │  │
│  │ User    │ ─► │ {KEY:   │ ─► │ Key-Val │ ─► │{Entity: │  │
│  │ END...  │    │  VALUE} │    │ Pairs   │    │ "User"} │  │
│  └─────────┘    └─────────┘    └─────────┘    └─────────┘  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Parsing Strategy Options:**

| Approach | Pros | Cons |
|----------|------|------|
| **Regex-based** | Simple, fast, no deps | Hard to maintain, poor errors |
| **Hand-written recursive descent** | Full control, good errors | More code, manual work |
| **Parser generator (ANTLR/PEG)** | Formal grammar, robust | Heavy dependency, overkill |
| **Line-by-line state machine** | Simple, good errors | Manual state management |

**Recommendation:** Line-by-line state machine with regex for value extraction.

```go
type ParserState int

const (
    StateStart ParserState = iota
    StateInTest
    StateExpectingValue
)

type Parser struct {
    state    ParserState
    current  *TestSpec
    specs    []TestSpec
    lineNum  int
}

func (p *Parser) ParseLine(line string) error {
    // State machine logic
}
```

---

### 3. Service Discovery

```
┌─────────────────────────────────────────────────────────────┐
│                  SERVICE DISCOVERY FLOW                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Project Root                                                │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              File System Walker                      │    │
│  │  (Walk directories, find marker files)               │    │
│  └─────────────────────────────────────────────────────┘    │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Service Detectors                       │    │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────┐ ┌─────────┐   │    │
│  │  │ Spring  │ │ Node.js │ │ Python  │ │  Go     │   │    │
│  │  │ Boot    │ │ Express │ │ Flask   │ │  Gin    │   │    │
│  │  └─────────┘ └─────────┘ └─────────┘ └─────────┘   │    │
│  └─────────────────────────────────────────────────────┘    │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Dependency Analyzer                     │    │
│  │  (Parse code for HTTP calls to other services)       │    │
│  └─────────────────────────────────────────────────────┘    │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Dependency Graph                        │    │
│  │  (Topological sort for start order)                  │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Detection Markers:**

| Service Type | Marker Files | Port Detection | Priority |
|--------------|--------------|----------------|----------|
| Spring Boot | `pom.xml` + spring deps, `build.gradle` | `application.yml`, `application.properties` | ✅ **v0.1** |
| Vite/React | `vite.config.ts`, `vite.config.js` | `vite.config.ts` | ✅ **v0.1** |
| Node.js | `package.json` + express/fastify | `package.json` scripts, `.env` | v1.0 |
| Python Flask | `requirements.txt` + flask | `app.py`, `.env` | v1.0 |
| Python FastAPI | `requirements.txt` + fastapi | `main.py`, `.env` | v1.0 |
| Go Gin | `go.mod` + gin | `main.go`, `.env` | v1.0 |

> **Implementation Strategy:** Focus on Spring Boot + React/Vite first. Build the detector interface to easily add other frameworks later.

**Leveraging Existing Dependency Graphs:**

Frameworks like Spring Boot already provide dependency information. Use these as a starting point instead of parsing code manually:

```go
// Strategy: Check for existing dependency info FIRST, fallback to code parsing

type DependencySource int

const (
    SourceMaven      DependencySource = iota // mvn dependency:tree
    SourceGradle                              // gradle dependencies
    SourcePackageJSON                         // package.json dependencies
    SourceCodeParsing                         // Fallback: parse source code
)

type DependencyDetector interface {
    // Returns dependencies from the best available source
    Detect(projectPath string) ([]Dependency, DependencySource, error)
}
```

**Spring Boot Dependency Detection:**

```go
type SpringBootDetector struct{}

func (d *SpringBootDetector) Detect(projectPath string) ([]Dependency, DependencySource, error) {
    // 1. Try Maven dependency:tree (if pom.xml exists)
    if fileExists(filepath.Join(projectPath, "pom.xml")) {
        deps, err := d.parseMavenDependencies(projectPath)
        if err == nil {
            return deps, SourceMaven, nil
        }
    }

    // 2. Try Gradle dependencies (if build.gradle exists)
    if fileExists(filepath.Join(projectPath, "build.gradle")) ||
       fileExists(filepath.Join(projectPath, "build.gradle.kts")) {
        deps, err := d.parseGradleDependencies(projectPath)
        if err == nil {
            return deps, SourceGradle, nil
        }
    }

    // 3. Fallback: Parse Java code for RestTemplate, WebClient, Feign
    return d.parseCodeForHTTPCalls(projectPath), SourceCodeParsing, nil
}

func (d *SpringBootDetector) parseMavenDependencies(path string) ([]Dependency, error) {
    // Run: mvn dependency:tree -DoutputType=dot
    // Parse output for inter-module dependencies
    cmd := exec.Command("mvn", "dependency:tree", "-DoutputType=dot")
    cmd.Dir = path
    output, err := cmd.Output()
    if err != nil {
        return nil, err
    }
    return parseDotGraph(output), nil
}

// For multi-module Maven projects
func (d *SpringBootDetector) detectMultiModule(path string) ([]Service, error) {
    // Parse parent pom.xml for <modules>
    // Each module becomes a service
    // Module dependencies become service dependencies
}
```

**React/Vite Detection:**

```go
type ViteDetector struct{}

func (d *ViteDetector) Detect(projectPath string) (*Service, error) {
    configPath := findViteConfig(projectPath) // vite.config.ts or .js
    if configPath == "" {
        return nil, ErrNotViteProject
    }

    // Parse vite.config for port and proxy settings
    config, _ := parseViteConfig(configPath)

    // Check package.json for API proxy targets (common pattern)
    // These indicate backend service dependencies
    pkgJSON, _ := parsePackageJSON(projectPath)

    return &Service{
        Name:      pkgJSON.Name,
        Type:      ServiceTypeVite,
        Port:      config.Server.Port, // default 5173
        DependsOn: extractProxyTargets(config), // e.g., /api -> localhost:8080
    }, nil
}
```

**Dependency Detection Priority:**

```
┌─────────────────────────────────────────────────────────────┐
│               DEPENDENCY DETECTION STRATEGY                  │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  1. Build Tool Native (FASTEST, MOST ACCURATE)               │
│     ├─ Maven: mvn dependency:tree                            │
│     ├─ Gradle: gradle dependencies                           │
│     └─ npm/pnpm: package.json + proxy config                 │
│                        │                                     │
│                        ▼ if unavailable                      │
│  2. Config Files                                             │
│     ├─ application.yml (spring.cloud.openfeign.client)       │
│     ├─ docker-compose.yml (depends_on)                       │
│     └─ .orchestr.yml (manual config)                         │
│                        │                                     │
│                        ▼ if unavailable                      │
│  3. Code Parsing (SLOWEST, FALLBACK)                         │
│     ├─ RestTemplate/WebClient URLs                           │
│     ├─ @FeignClient annotations                              │
│     └─ fetch/axios calls                                     │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

### 4. Service Manager

```
┌─────────────────────────────────────────────────────────────┐
│                   SERVICE LIFECYCLE                          │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│     ┌─────────┐    ┌─────────┐    ┌─────────┐              │
│     │ STOPPED │───►│STARTING │───►│ HEALTHY │              │
│     └─────────┘    └─────────┘    └─────────┘              │
│          ▲              │              │                    │
│          │              ▼              ▼                    │
│          │         ┌─────────┐    ┌─────────┐              │
│          └─────────│ FAILED  │    │ RUNNING │              │
│                    └─────────┘    └─────────┘              │
│                                        │                    │
│                                        ▼                    │
│                                   ┌─────────┐              │
│                                   │STOPPING │              │
│                                   └─────────┘              │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Process Management Options:**

| Approach | Pros | Cons |
|----------|------|------|
| **os/exec (native)** | No deps, full control, lightweight, fast startup | Manual process mgmt, signal handling |
| **Docker SDK** | Isolation, reproducible, network control | Requires Docker, slower startup |
| **Both (hybrid)** | Flexibility, user choice, best of both worlds | More code (but worth it) |

**Recommendation:** ✅ **Hybrid from day one** - Build the interface to support both, implement native first.

```go
// Runner interface - both implementations satisfy this
type ServiceRunner interface {
    Start(ctx context.Context, service Service) error
    Stop(ctx context.Context, service Service) error
    HealthCheck(ctx context.Context, service Service) error
    Logs(ctx context.Context, service Service) (io.Reader, error)
    Type() RunnerType
}

type RunnerType string

const (
    RunnerTypeNative RunnerType = "native"
    RunnerTypeDocker RunnerType = "docker"
)

// Factory creates the right runner based on config/detection
func NewRunner(config *Config, service Service) ServiceRunner {
    // Priority:
    // 1. Explicit config: --runner=docker or .orchestr.yml runner: docker
    // 2. Dockerfile present in service dir → DockerRunner
    // 3. Default → NativeRunner

    if config.Runner == "docker" || fileExists(service.Path, "Dockerfile") {
        return &DockerRunner{client: dockerClient}
    }
    return &NativeRunner{}
}

// Native runner - os/exec based (v0.1)
type NativeRunner struct{}

func (r *NativeRunner) Start(ctx context.Context, svc Service) error {
    cmd := exec.CommandContext(ctx, svc.StartCommand[0], svc.StartCommand[1:]...)
    cmd.Dir = svc.Path
    cmd.Env = append(os.Environ(), svc.Env...)
    // ... process management
}

// Docker runner - Docker SDK (v0.1, optional)
type DockerRunner struct {
    client *docker.Client
}

func (r *DockerRunner) Start(ctx context.Context, svc Service) error {
    // Build image if needed, create container, start
    // Inject env vars, map ports, attach to network
}
```

**Why Hybrid from Start:**
- Interface is cheap to define upfront
- Avoids refactoring later
- Users with Dockerfiles get Docker automatically
- CI environments often prefer Docker
- Local dev often prefers native (faster)

**Health Check Strategy:**

```go
type HealthChecker struct {
    MaxRetries     int           // 30
    RetryInterval  time.Duration // 1s
    Timeout        time.Duration // 30s
}

func (h *HealthChecker) Check(endpoint string) error {
    // 1. TCP connect check
    // 2. HTTP GET /health or /actuator/health
    // 3. Custom health endpoint from config
}
```

---

### 5. Mock Server

```
┌─────────────────────────────────────────────────────────────┐
│                    MOCK SERVER ARCHITECTURE                  │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                   Mock Registry                      │    │
│  │  map[host]MockServer                                 │    │
│  └─────────────────────────────────────────────────────┘    │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────────────┐    │
│  │                   Mock Server                        │    │
│  │  ┌─────────────────────────────────────────────┐    │    │
│  │  │ Route Matcher                                │    │    │
│  │  │ POST /v1/charges → chargeHandler            │    │    │
│  │  │ GET /v1/customers/:id → customerHandler     │    │    │
│  │  └─────────────────────────────────────────────┘    │    │
│  │                         │                            │    │
│  │                         ▼                            │    │
│  │  ┌─────────────────────────────────────────────┐    │    │
│  │  │ Response Generator                          │    │    │
│  │  │ - Static JSON                               │    │    │
│  │  │ - Template with dynamic values              │    │    │
│  │  │ - Recorded responses (future)               │    │    │
│  │  └─────────────────────────────────────────────┘    │    │
│  │                         │                            │    │
│  │                         ▼                            │    │
│  │  ┌─────────────────────────────────────────────┐    │    │
│  │  │ Call Recorder                               │    │    │
│  │  │ - Timestamp                                 │    │    │
│  │  │ - Request body                              │    │    │
│  │  │ - Response body                             │    │    │
│  │  └─────────────────────────────────────────────┘    │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Built-in Mock Templates:**

```go
var MockTemplates = map[string]MockTemplate{
    "stripe": {
        BaseURL: "api.stripe.com",
        Routes: []MockRoute{
            {
                Method: "POST",
                Path:   "/v1/charges",
                Response: `{"id":"ch_{{uuid}}","status":"succeeded","amount":{{.amount}}}`,
            },
            {
                Method: "GET",
                Path:   "/v1/customers/:id",
                Response: `{"id":"{{.id}}","email":"mock@example.com"}`,
            },
        },
    },
    "sendgrid": {
        BaseURL: "api.sendgrid.com",
        Routes: []MockRoute{
            {
                Method: "POST",
                Path:   "/v3/mail/send",
                Response: `{"message_id":"msg_{{uuid}}"}`,
            },
        },
    },
    // twilio, aws-ses, stripe, etc.
}
```

**URL Injection Strategy:**

```go
// Detect external API usage in code
externalAPIs := detector.FindExternalAPIs(projectPath)
// Found: https://api.stripe.com

// Start mock server
mockServer := mock.NewServer(externalAPIs["stripe"])
mockServer.Start() // localhost:9001

// Inject via environment variable
env["STRIPE_API_URL"] = mockServer.URL()
// Or: STRIPE_BASE_URL, STRIPE_HOST, etc.
```

---

### 6. Test Executor

```
┌─────────────────────────────────────────────────────────────┐
│                    TEST EXECUTION FLOW                       │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  TestSpec[]                                                  │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Variable Resolver                       │    │
│  │  Replace {{token}} with saved values                 │    │
│  └─────────────────────────────────────────────────────┘    │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Request Builder                         │    │
│  │  Build HTTP request from TestSpec                    │    │
│  └─────────────────────────────────────────────────────┘    │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              HTTP Client                             │    │
│  │  Execute request with tracing headers                │    │
│  └─────────────────────────────────────────────────────┘    │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Response Validator                      │    │
│  │  - Status code match                                 │    │
│  │  - Body pattern match (with <uuid>, <jwt>, etc.)    │    │
│  │  - EXPECT_CALLS validation                           │    │
│  └─────────────────────────────────────────────────────┘    │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Value Extractor                         │    │
│  │  SAVE: token from body.token                         │    │
│  └─────────────────────────────────────────────────────┘    │
│       │                                                      │
│       ▼                                                      │
│  TestResult                                                  │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Pattern Matching:**

```go
type PatternMatcher struct{}

// Match response body against expected pattern
func (m *PatternMatcher) Match(expected, actual string) (bool, []Diff) {
    // Parse expected as JSON with patterns
    // <uuid> → regex: [0-9a-f]{8}-[0-9a-f]{4}-...
    // <jwt>  → regex: eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+
    // <int>  → regex: -?\d+
    // <email> → regex: .+@.+\..+
}

var PatternRegex = map[string]string{
    "<uuid>":    `[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}`,
    "<jwt>":     `eyJ[A-Za-z0-9_-]+\.eyJ[A-Za-z0-9_-]+\.[A-Za-z0-9_-]+`,
    "<int>":     `-?\d+`,
    "<float>":   `-?\d+\.?\d*`,
    "<string>":  `.+`,
    "<email>":   `[^@]+@[^@]+\.[^@]+`,
    "<iso8601>": `\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}`,
}
```

---

### 7. Distributed Tracer

```
┌─────────────────────────────────────────────────────────────┐
│                    TRACING ARCHITECTURE                      │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  Request with X-Trace-ID: abc123                             │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────┐    ┌─────────┐    ┌─────────┐    ┌─────────┐  │
│  │ main-api│───►│  auth   │    │ payment │───►│ stripe  │  │
│  │ :8080   │    │ :8081   │    │ :8082   │    │ (mock)  │  │
│  └─────────┘    └─────────┘    └─────────┘    └─────────┘  │
│       │                             │                        │
│       └─────────────────────────────┘                        │
│                                                              │
│  Trace Collector (in-memory)                                 │
│  ┌─────────────────────────────────────────────────────┐    │
│  │ TraceID: abc123                                      │    │
│  │ ├─ 0ms   main-api    POST /api/pay     STARTED      │    │
│  │ ├─ 5ms   auth        GET  /verify      200 (5ms)    │    │
│  │ ├─ 12ms  payment     POST /charge      STARTED      │    │
│  │ ├─ 18ms  stripe      POST /v1/charges  200 (6ms)    │    │
│  │ ├─ 25ms  payment     POST /charge      200 (13ms)   │    │
│  │ └─ 30ms  main-api    POST /api/pay     200 (30ms)   │    │
│  └─────────────────────────────────────────────────────┘    │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

**Tracing Implementation Options:**

| Approach | Pros | Cons |
|----------|------|------|
| **Custom headers + log parsing** | Simple, no service changes | Requires log format, less accurate |
| **HTTP proxy interception** | Full visibility, accurate | Complexity, port management |
| **OpenTelemetry injection** | Standard, rich data | Requires OTEL collector, heavyweight |

**Recommendation:** Start with custom headers + mock server recording, add proxy later.

```go
type TraceCollector struct {
    mu     sync.RWMutex
    traces map[string]*Trace
}

type Trace struct {
    ID        string
    StartTime time.Time
    Spans     []Span
}

type Span struct {
    TraceID   string
    Service   string
    Method    string
    Path      string
    StartTime time.Time
    Duration  time.Duration
    Status    int
    Error     string
}

// Collect span from mock server
func (c *TraceCollector) RecordSpan(span Span) {
    c.mu.Lock()
    defer c.mu.Unlock()
    trace := c.traces[span.TraceID]
    trace.Spans = append(trace.Spans, span)
}
```

---

### 8. Reporter

```
┌─────────────────────────────────────────────────────────────┐
│                    REPORT GENERATION                         │
├─────────────────────────────────────────────────────────────┤
│                                                              │
│  TestResults + Traces + ServiceHealth                        │
│       │                                                      │
│       ▼                                                      │
│  ┌─────────────────────────────────────────────────────┐    │
│  │              Report Builder                          │    │
│  │  - Aggregate results                                 │    │
│  │  - Calculate statistics                              │    │
│  │  - Correlate traces with failures                    │    │
│  └─────────────────────────────────────────────────────┘    │
│       │                                                      │
│       ├──────────────┬──────────────┬──────────────┐        │
│       ▼              ▼              ▼              ▼        │
│  ┌─────────┐   ┌─────────┐   ┌─────────┐   ┌─────────┐     │
│  │Markdown │   │  JSON   │   │  HTML   │   │ Console │     │
│  │ Writer  │   │ Writer  │   │ Writer  │   │ Writer  │     │
│  └─────────┘   └─────────┘   └─────────┘   └─────────┘     │
│       │              │              │              │        │
│       ▼              ▼              ▼              ▼        │
│  report.md     report.json   report.html    stdout         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

---

## Data Flow

### Complete Test Run Flow

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           COMPLETE TEST RUN FLOW                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                              │
│  1. INITIALIZATION                                                           │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │ CLI Parse → Load Config → Parse Orchestrfile → Validate Specs      │     │
│  └────────────────────────────────────────────────────────────────────┘     │
│                                      │                                       │
│                                      ▼                                       │
│  2. DISCOVERY                                                                │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │ Scan Project → Detect Services → Find Dependencies → Build Graph   │     │
│  └────────────────────────────────────────────────────────────────────┘     │
│                                      │                                       │
│                                      ▼                                       │
│  3. SETUP                                                                    │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │ Start Mocks → Start Services (topo order) → Health Check All       │     │
│  └────────────────────────────────────────────────────────────────────┘     │
│                                      │                                       │
│                                      ▼                                       │
│  4. EXECUTION                                                                │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │ For each TestSpec:                                                  │     │
│  │   Resolve Variables → Build Request → Execute → Validate → Save    │     │
│  │   Collect Trace Spans from Mocks                                    │     │
│  └────────────────────────────────────────────────────────────────────┘     │
│                                      │                                       │
│                                      ▼                                       │
│  5. TEARDOWN                                                                 │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │ Stop Services (reverse order) → Stop Mocks → Collect Logs          │     │
│  └────────────────────────────────────────────────────────────────────┘     │
│                                      │                                       │
│                                      ▼                                       │
│  6. REPORTING                                                                │
│  ┌────────────────────────────────────────────────────────────────────┐     │
│  │ Aggregate Results → Correlate Traces → Generate Report → Output    │     │
│  └────────────────────────────────────────────────────────────────────┘     │
│                                                                              │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Library Choices

### CLI Framework

| Library | Stars | Pros | Cons | Verdict |
|---------|-------|------|------|---------|
| **cobra** | 35k+ | Industry standard, excellent docs, auto-completion, widely used | Verbose setup | ✅ **Use this** |
| **urfave/cli** | 21k+ | Simple, clean API | Less features, smaller community | Good alternative |
| **kong** | 1.5k+ | Struct-based, type-safe | Smaller community | Interesting but risky |

**Decision:** Use `cobra` - it's the Go CLI standard (kubectl, docker, gh all use it).

```go
// Example with cobra
var rootCmd = &cobra.Command{
    Use:   "orchestr",
    Short: "Orchestrate multi-service testing",
}

var testCmd = &cobra.Command{
    Use:   "test [path]",
    Short: "Run tests against services",
    RunE:  runTest,
}
```

---

### Configuration

| Library | Pros | Cons | Verdict |
|---------|------|------|---------|
| **viper** | Full-featured, env + file + flags, popular | Heavy, complex for simple needs | ✅ **Use this** |
| **envconfig** | Simple, struct-based | No file support | Too simple |
| **koanf** | Modular, lightweight | Less popular | Good alternative |
| **godotenv + manual** | Minimal deps | Manual merging | For v0.1 only |

**Decision:** Use `viper` for config flexibility (env vars, config files, CLI flags).

```go
func loadConfig() (*Config, error) {
    viper.SetConfigName(".orchestr")
    viper.SetConfigType("yaml")
    viper.AddConfigPath(".")

    viper.SetEnvPrefix("ORCHESTR")
    viper.AutomaticEnv()

    viper.SetDefault("timeout", "30s")
    viper.SetDefault("parallel", 1)

    var config Config
    if err := viper.Unmarshal(&config); err != nil {
        return nil, err
    }
    return &config, nil
}
```

---

### HTTP Client

| Library | Pros | Cons | Verdict |
|---------|------|------|---------|
| **net/http (stdlib)** | No deps, full control, sufficient | Verbose, manual JSON handling | ✅ **Use this** |
| **resty** | Clean API, auto retry, chain methods | Extra dependency | Nice but unnecessary |
| **req** | Modern, debug friendly | Less popular | Good alternative |

**Decision:** Use standard `net/http` - we need full control for tracing headers anyway.

```go
func (e *Executor) Execute(spec TestSpec) (*Response, error) {
    req, _ := http.NewRequest(spec.Method, spec.Endpoint, strings.NewReader(spec.Data))

    // Inject trace headers
    req.Header.Set("X-Trace-ID", e.traceID)
    req.Header.Set("X-Span-ID", uuid.New().String())

    // Auth header
    if spec.Auth != "" {
        req.Header.Set("Authorization", spec.Auth)
    }

    client := &http.Client{Timeout: 30 * time.Second}
    return client.Do(req)
}
```

---

### YAML/JSON Parsing

| Library | Pros | Cons | Verdict |
|---------|------|------|---------|
| **encoding/json (stdlib)** | Built-in, fast, standard | - | ✅ **Use for JSON** |
| **go-yaml/yaml** | Standard YAML lib | v3 API different from v2 | ✅ **Use for YAML** |
| **goccy/go-json** | Faster than stdlib | Extra dep, marginal gain | Unnecessary |

**Decision:** Use stdlib `encoding/json` + `gopkg.in/yaml.v3`

---

### Process Management

| Library | Pros | Cons | Verdict |
|---------|------|------|---------|
| **os/exec (stdlib)** | Built-in, lightweight | Manual signal handling | ✅ **Use for native** |
| **docker/docker** | Isolation, reproducible | Requires Docker installed | ✅ **Optional for Docker mode** |
| **testcontainers-go** | Nice API for containers | Test-focused, less control | Too opinionated |

**Decision:** Use `os/exec` as primary, Docker SDK as optional.

```go
type ProcessManager struct {
    processes map[string]*exec.Cmd
    mu        sync.RWMutex
}

func (pm *ProcessManager) Start(ctx context.Context, svc Service) error {
    cmd := exec.CommandContext(ctx, svc.StartCommand[0], svc.StartCommand[1:]...)
    cmd.Dir = svc.Path
    cmd.Env = append(os.Environ(), svc.Env...)

    // Capture stdout/stderr
    stdout, _ := cmd.StdoutPipe()
    stderr, _ := cmd.StderrPipe()

    if err := cmd.Start(); err != nil {
        return err
    }

    pm.mu.Lock()
    pm.processes[svc.Name] = cmd
    pm.mu.Unlock()

    // Start log capture goroutines
    go pm.captureOutput(svc.Name, stdout, "stdout")
    go pm.captureOutput(svc.Name, stderr, "stderr")

    return nil
}
```

---

### Concurrency & Synchronization

| Pattern | Use Case |
|---------|----------|
| **sync.WaitGroup** | Wait for multiple services to start |
| **errgroup** | Parallel operations with error handling |
| **context.Context** | Cancellation, timeouts |
| **channels** | Service events, log streaming |

```go
import "golang.org/x/sync/errgroup"

func (sm *ServiceManager) StartAll(ctx context.Context, services []Service) error {
    g, ctx := errgroup.WithContext(ctx)

    // Start services in dependency order (already sorted)
    for _, batch := range sm.batchByDependencyLevel(services) {
        for _, svc := range batch {
            svc := svc // capture
            g.Go(func() error {
                if err := sm.Start(ctx, svc); err != nil {
                    return fmt.Errorf("failed to start %s: %w", svc.Name, err)
                }
                return sm.WaitHealthy(ctx, svc)
            })
        }
        // Wait for batch before starting next level
        if err := g.Wait(); err != nil {
            return err
        }
        g, ctx = errgroup.WithContext(ctx)
    }

    return nil
}
```

---

### Logging

| Library | Pros | Cons | Verdict |
|---------|------|------|---------|
| **log/slog (stdlib)** | Built-in (Go 1.21+), structured | New, less docs | ✅ **Use this** |
| **zerolog** | Fast, zero alloc | Extra dep | Good alternative |
| **zap** | Popular, fast | Complex setup | Overkill |
| **logrus** | Popular, hooks | Maintenance mode | Avoid |

**Decision:** Use `log/slog` (stdlib Go 1.21+) - modern, structured, no deps.

```go
import "log/slog"

func main() {
    logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
        Level: slog.LevelInfo,
    }))
    slog.SetDefault(logger)

    slog.Info("starting service",
        "name", service.Name,
        "port", service.Port,
    )
}
```

---

### Terminal UI

| Library | Pros | Cons | Verdict |
|---------|------|------|---------|
| **fatih/color** | Simple, just colors | Limited features | v0.1 only |
| **charmbracelet/lipgloss** | Beautiful styling | Learning curve | ✅ **Use with bubbletea** |
| **charmbracelet/bubbletea** | Full TUI framework, elm architecture | Learning curve | ✅ **Use for v1.0** |
| **schollz/progressbar** | Simple progress bars | Limited, no interactivity | v0.1 only |

**Decision:** Phased approach - simple output for v0.1, Bubble Tea TUI for v1.0.

**Why Bubble Tea is NOT overkill:**

Bubble Tea enables Claude Code-style interactive features:
- Live service status dashboard (services starting/healthy/failed)
- Real-time log streaming with service selector
- Interactive test selection and filtering
- Distributed trace timeline visualization
- Scrollable test results with detail expansion

```
┌─────────────────────────────────────────────────────────────┐
│  ORCHESTR                                    v1.0.0         │
├─────────────────────────────────────────────────────────────┤
│  Services                                                   │
│  ├─ ✓ auth-service      :8081  healthy   45ms avg          │
│  ├─ ✓ payment-service   :8082  healthy  120ms avg          │
│  ├─ ◐ notification-svc  :8083  starting...                 │
│  └─ ○ main-api          :8080  waiting                     │
├─────────────────────────────────────────────────────────────┤
│  Tests  [3/12]                                              │
│  ├─ ✓ User Registration                          23ms      │
│  ├─ ✓ User Login                                 45ms      │
│  ├─ ✗ User Payment ← auth-service timeout                  │
│  └─ ◐ Running: Email Confirmation...                       │
├─────────────────────────────────────────────────────────────┤
│  Logs [payment-service]  ↑↓ scroll  tab: switch service    │
│  > POST /charge 200 12ms                                   │
│  > Stripe API call succeeded                               │
└─────────────────────────────────────────────────────────────┘
```

**Phased Implementation:**

```go
// v0.1 - Simple output
func printResult(test TestResult) {
    if test.Passed {
        color.Green("✓ %s", test.Name)
    } else {
        color.Red("✗ %s", test.Name)
    }
}

// v1.0 - Bubble Tea TUI
type Model struct {
    services    []ServiceStatus
    tests       []TestResult
    logs        []LogLine
    activeTab   Tab
    viewport    viewport.Model
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
    switch msg := msg.(type) {
    case ServiceHealthMsg:
        m.services[msg.Index].Status = msg.Status
    case TestResultMsg:
        m.tests = append(m.tests, msg.Result)
    case LogMsg:
        m.logs = append(m.logs, msg.Line)
    }
    return m, nil
}
```

---

### File Walking

| Library | Pros | Cons | Verdict |
|---------|------|------|---------|
| **filepath.Walk (stdlib)** | Built-in, simple | Doesn't follow symlinks well | OK for simple |
| **filepath.WalkDir (Go 1.16+)** | Faster, DirEntry | Still basic | ✅ **Use this** |
| **karrick/godirwalk** | Very fast | Extra dep | Only if perf issue |

**Decision:** Use `filepath.WalkDir` (stdlib).

---

### Testing the Tool Itself

| Library | Pros | Cons | Verdict |
|---------|------|------|---------|
| **testing (stdlib)** | Built-in | Verbose assertions | ✅ **Use this** |
| **stretchr/testify** | Nice assertions, mocks | Extra dep, but worth it | ✅ **Use for assertions** |
| **go-cmp** | Deep equality | Verbose output | For complex diffs |

**Decision:** Use `testing` + `testify/assert` for cleaner tests.

---

## Project Structure

```
orchestr/
├── cmd/
│   └── orchestr/
│       └── main.go              # Entry point
├── internal/
│   ├── cli/
│   │   ├── root.go              # Root command
│   │   ├── test.go              # Test command
│   │   ├── init.go              # Init command
│   │   └── discover.go          # Discover command
│   ├── config/
│   │   └── config.go            # Configuration loading
│   ├── parser/
│   │   ├── parser.go            # Orchestrfile parser
│   │   ├── lexer.go             # Tokenizer
│   │   └── patterns.go          # Type patterns (<uuid>, etc.)
│   ├── discovery/
│   │   ├── discovery.go         # Service discovery
│   │   ├── detector.go          # Detector interface
│   │   ├── springboot.go        # Spring Boot detector
│   │   ├── nodejs.go            # Node.js detector
│   │   ├── python.go            # Python detector
│   │   └── graph.go             # Dependency graph
│   ├── runner/
│   │   ├── runner.go            # Service runner interface
│   │   ├── native.go            # os/exec runner
│   │   ├── docker.go            # Docker runner
│   │   └── health.go            # Health checker
│   ├── mock/
│   │   ├── server.go            # Mock HTTP server
│   │   ├── templates.go         # Built-in mock templates
│   │   ├── registry.go          # Mock registry
│   │   └── detector.go          # External API detector
│   ├── executor/
│   │   ├── executor.go          # Test executor
│   │   ├── request.go           # Request builder
│   │   ├── validator.go         # Response validator
│   │   └── variables.go         # Variable resolver
│   ├── tracer/
│   │   ├── tracer.go            # Trace collector
│   │   ├── span.go              # Span definition
│   │   └── timeline.go          # Timeline builder
│   ├── reporter/
│   │   ├── reporter.go          # Reporter interface
│   │   ├── markdown.go          # Markdown reporter
│   │   ├── json.go              # JSON reporter
│   │   └── console.go           # Console reporter
│   └── models/
│       ├── service.go           # Service model
│       ├── testspec.go          # Test specification
│       └── result.go            # Test result
├── pkg/                         # Public APIs (if any)
├── templates/                   # Orchestrfile templates
│   └── default.orchestr
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

---

## Implementation Phases

### Phase 1: Foundation (v0.1 - Single Service)

```
Week 1: CLI + Parser
├── [ ] Cobra CLI setup (test, init, version)
├── [ ] Configuration loading (viper)
├── [ ] Orchestrfile parser (line-by-line)
├── [ ] Pattern matcher (<uuid>, <jwt>, etc.)
└── [ ] Unit tests for parser

Week 2: Service Detection + Running
├── [ ] Service detector interface
├── [ ] Spring Boot detector
├── [ ] Node.js detector
├── [ ] Native process runner (os/exec)
├── [ ] Health check implementation
└── [ ] Graceful shutdown

Week 3: Test Execution
├── [ ] HTTP client with tracing headers
├── [ ] Request builder from TestSpec
├── [ ] Response validator
├── [ ] Variable save/restore
├── [ ] Test runner loop
└── [ ] Console output

Week 4: Reporting + Polish
├── [ ] Markdown report generator
├── [ ] JSON report generator
├── [ ] Console pretty printing
├── [ ] Error messages
├── [ ] Documentation
└── [ ] Release binaries
```

### Phase 2: Multi-Service (v1.0)

```
Week 5-6: Service Orchestration
├── [ ] Multi-service discovery
├── [ ] Dependency graph builder
├── [ ] Topological sort
├── [ ] Parallel service startup
├── [ ] Port allocation
└── [ ] Service health management

Week 7-8: Mock Generation
├── [ ] External API detector
├── [ ] Mock server implementation
├── [ ] Built-in mock templates
├── [ ] URL injection
├── [ ] Mock response customization
└── [ ] Call recording

Week 9: Distributed Tracing
├── [ ] Trace collector
├── [ ] Span aggregation
├── [ ] Timeline visualization
├── [ ] EXPECT_CALLS validation
└── [ ] Cross-service error correlation

Week 10: Polish
├── [ ] Root cause analysis
├── [ ] Failure isolation
├── [ ] Performance optimization
├── [ ] Extended testing
└── [ ] Documentation update
```

---

## Technical Decisions

### Decision 1: Orchestrfile Format

**Options:**
1. YAML - Standard, familiar
2. TOML - Simple, readable
3. Custom DSL - Optimized for use case

**Decision:** Custom DSL

**Rationale:**
- More readable than YAML for test specs
- Simpler to write (no YAML indentation issues)
- Purpose-built for testing semantics
- Can still support YAML as alternative

---

### Decision 2: Process vs Container

**Options:**
1. Native processes only
2. Docker containers only
3. Hybrid (user choice)

**Decision:** Hybrid (native default, Docker optional)

**Rationale:**
- Native is faster for development
- Docker provides isolation for CI
- User chooses based on environment
- Start with native, add Docker in v1.0

---

### Decision 3: Tracing Strategy

**Options:**
1. Log parsing
2. HTTP proxy
3. Mock server recording + custom headers

**Decision:** Mock server recording + custom headers

**Rationale:**
- Mock servers already capture calls
- Custom headers (X-Trace-ID) enable correlation
- No proxy complexity
- Works without modifying services
- Can add proxy layer later if needed

---

### Decision 4: Service Communication

**Options:**
1. Direct HTTP calls
2. Service mesh (overkill)
3. Shared network namespace

**Decision:** Direct HTTP calls

**Rationale:**
- Simple, works everywhere
- Services use localhost with different ports
- Mock servers get assigned ports dynamically
- Environment variables for URL injection

---

## Appendix: Key Interfaces

```go
// Service represents a detected service
type Service struct {
    Name         string
    Type         ServiceType
    Path         string
    Port         int
    StartCommand []string
    HealthPath   string
    Env          []string
    DependsOn    []string
}

// ServiceDetector detects services in a directory
type ServiceDetector interface {
    Detect(path string) (*Service, error)
    Type() ServiceType
}

// ServiceRunner starts and stops services
type ServiceRunner interface {
    Start(ctx context.Context, svc Service) error
    Stop(ctx context.Context, svc Service) error
    HealthCheck(ctx context.Context, svc Service) error
    Logs(svc Service) (io.Reader, error)
}

// TestSpec represents a parsed test case
type TestSpec struct {
    Entity       string
    Endpoint     string
    Method       string
    Headers      map[string]string
    Auth         string
    Data         string
    ExpectStatus int
    ExpectBody   string
    ExpectCalls  []string
    Save         map[string]string
}

// TestResult represents the result of a test
type TestResult struct {
    Spec     TestSpec
    Passed   bool
    Status   int
    Body     string
    Duration time.Duration
    Error    error
    Trace    *Trace
}

// MockServer handles mock responses for external APIs
type MockServer interface {
    Start() error
    Stop() error
    URL() string
    RecordedCalls() []MockCall
}

// TraceCollector collects distributed traces
type TraceCollector interface {
    StartTrace(id string) *Trace
    RecordSpan(span Span)
    GetTrace(id string) *Trace
}

// Reporter generates test reports
type Reporter interface {
    Generate(results []TestResult, traces []*Trace) error
}
```

---

## Summary

| Component | Library | Phase | Rationale |
|-----------|---------|-------|-----------|
| CLI | cobra | v0.1 | Industry standard |
| Config | viper | v0.1 | Flexible, multi-source |
| HTTP | net/http | v0.1 | Full control, no deps |
| YAML | gopkg.in/yaml.v3 | v0.1 | Standard |
| Process | os/exec + docker/docker | v0.1 | Hybrid from start |
| Logging | log/slog | v0.1 | Stdlib, structured |
| TUI | fatih/color → bubbletea | v0.1 → v1.0 | Simple first, rich TUI later |
| Testing | testify | v0.1 | Better assertions |

**Key architectural choices:**
1. **Plugin-based service detection** - Spring Boot + Vite first, interface ready for others
2. **Leverage existing dep graphs** - Maven/Gradle dependency:tree before code parsing
3. **Hybrid process runner** - os/exec default, Docker when Dockerfile present
4. **Custom DSL for test specs** - Readable, purpose-built Orchestrfile
5. **Mock servers for external APIs** - Automatic, no proxy complexity
6. **Header-based tracing** - Simple, effective, works without service changes
7. **Phased TUI** - Simple colors v0.1, Bubble Tea interactive TUI v1.0
