# Orchestr Development Roadmap

> Clear milestones with specific deliverables. No vague goals.

---

## v0.1 - Single Service MVP

**Goal:** Test a single Spring Boot or React/Vite service with zero configuration.

**Scope:** One backend service, no orchestration, no external API mocking.

### Week 1: CLI + Orchestrfile Parser ✓

**Deliverables:**
- [x] Cobra CLI with `test`, `init`, `version` commands
- [x] Viper config loading (`.orchestr.yml`, env vars, flags)
- [x] Line-by-line Orchestrfile parser with state machine
- [x] Pattern matchers: `<uuid>`, `<jwt>`, `<int>`, `<string>`, `<email>`, `<iso8601>`, `<url>`, `<any>`
- [x] Extended patterns with constraints: `<array:min=1>`, `<int:gte=0>`, `<string:nonempty>`
- [x] Structural types: `<array>`, `<object>` with constraint validation
- [x] Variable substitution: `{{varName}}`
- [x] Parser error messages with line numbers

**Definition of Done:**
- [x] `orchestr init` generates a valid Orchestrfile template
- [x] `orchestr test .` parses Orchestrfile without errors
- [x] Parser unit tests cover all keywords and patterns (225+ tests)
- [x] Invalid Orchestrfile shows clear error: `line 5: invalid METHOD 'GETT'`

---

### Week 2: Service Detection + Runner

**Deliverables:**
- [ ] `ServiceDetector` interface
- [ ] Spring Boot detector (pom.xml, build.gradle, application.yml port extraction)
- [ ] React/Vite detector (vite.config.ts/js, port extraction)
- [ ] `ServiceRunner` interface with `NativeRunner` (os/exec)
- [ ] Health check implementation (TCP + HTTP `/health` or `/actuator/health`)
- [ ] Graceful shutdown with SIGINT/SIGTERM handling
- [ ] Log capture (stdout/stderr) per service

**Definition of Done:**
- `orchestr discover .` prints detected service type, port, start command
- Spring Boot project starts with `mvn spring-boot:run` or `./gradlew bootRun`
- Vite project starts with `npm run dev` or `pnpm dev`
- Service stops cleanly on Ctrl+C
- Startup failures show error within 30s timeout

---

### Week 3: Test Execution

**Deliverables:**
- [ ] HTTP client with configurable timeout
- [ ] Request builder from `TestSpec`
- [ ] Trace header injection (`X-Trace-ID`, `X-Request-ID`)
- [ ] Response validator (status code + body pattern matching)
- [ ] JSON path extraction for `SAVE` directive
- [ ] Variable storage for test chaining
- [ ] Sequential test runner

**Definition of Done:**
- CRUD test flow works: POST (save id) → GET (use id) → PUT → DELETE
- `EXPECT: 201 with {id: <uuid>}` validates correctly
- `SAVE: token from body.data.accessToken` extracts nested values
- Failed assertions show diff: `expected status 200, got 404`

---

### Week 4: Reporting + Release

**Deliverables:**
- [ ] Console reporter with colors (pass=green, fail=red)
- [ ] Markdown report generator (`orchestr-report.md`)
- [ ] JSON report generator (`--format json`)
- [ ] Summary: X/Y passed, duration, failures list
- [ ] Goreleaser setup for cross-platform binaries
- [ ] GitHub Actions: test + lint + release workflow

**Definition of Done:**
- `orchestr test . --format markdown` writes report file
- Report includes: service info, test results, failure details
- CI produces binaries for linux/darwin/windows (amd64/arm64)
- README has quick-start and usage examples
- End-to-end: clone sample project → run `orchestr test .` → see report

---

### v0.1 Out of Scope

- Multi-service orchestration
- External API mocking
- Distributed tracing across services
- Docker runner
- Node.js/Python/Go detectors
- TUI (terminal UI dashboard)
- Performance/load testing
- Code quality analysis

---

## v1.0 - Multi-Service Testing

**Goal:** Orchestrate multiple services with dependency management and auto-mocking.

**Dependencies:** v0.1 complete (single service works).

### Week 5: Multi-Service Discovery

**Deliverables:**
- [ ] Recursive directory scanner
- [ ] Maven multi-module detection (`<modules>` in parent pom.xml)
- [ ] Gradle multi-project detection
- [ ] Build dependency graph from:
  - Maven `dependency:tree`
  - Gradle `dependencies`
  - Fallback: code parsing for RestTemplate/WebClient URLs
- [ ] Topological sort for start order

**Definition of Done:**
- `orchestr discover ./microservices` lists all services with dependencies
- Output shows: `auth-service → (none), payment-service → auth-service`
- Circular dependencies detected and reported as error

**Blocked by:** Week 2 service detection

---

### Week 6: Service Orchestration

**Deliverables:**
- [ ] Parallel service startup within dependency levels
- [ ] Dynamic port allocation (avoid conflicts)
- [ ] Health check coordination (wait for dependencies)
- [ ] Service registry (name → host:port mapping)
- [ ] Environment variable injection for service URLs
- [ ] Graceful shutdown in reverse dependency order

**Definition of Done:**
- 4 services start in correct order with proper health checks
- Services can communicate via injected URLs
- One service failure stops dependent services
- Clean shutdown kills all services

**Blocked by:** Week 5 discovery

---

### Week 7: External API Detection

**Deliverables:**
- [ ] Code parser for HTTP calls to external domains
  - Java: RestTemplate, WebClient, Feign, OkHttp
  - JavaScript: fetch, axios
- [ ] External API registry (stripe.com, sendgrid.com, etc.)
- [ ] Config-based external API declaration (`.orchestr.yml`)
- [ ] Classify: internal service vs external API

**Definition of Done:**
- Detects `restTemplate.post("https://api.stripe.com/v1/charges")`
- Lists found external APIs: `stripe.com`, `sendgrid.com`
- User can override with config: `externals: [stripe, twilio]`

**Blocked by:** None (can run parallel to Week 6)

---

### Week 8: Mock Server + Templates

**Deliverables:**
- [ ] Mock HTTP server (net/http)
- [ ] Route matcher with path params (`/v1/customers/:id`)
- [ ] Response templating (`{{uuid}}`, `{{.amount}}`)
- [ ] Built-in templates: Stripe, SendGrid, Twilio
- [ ] Call recording for validation
- [ ] URL injection via environment variables

**Definition of Done:**
- Mock Stripe server responds to `/v1/charges` with valid JSON
- Service configured with `STRIPE_API_URL=http://localhost:9001`
- `EXPECT_CALLS: stripe:/v1/charges` validates mock was called
- Custom mock config in `.orchestr.yml`

**Blocked by:** Week 7 detection

---

### Week 9: Distributed Tracing

**Deliverables:**
- [ ] Trace ID propagation through request chain
- [ ] Span collection from mock servers
- [ ] Timeline builder (ordered by timestamp)
- [ ] Cross-service error correlation
- [ ] Trace visualization in report

**Definition of Done:**
- Single request shows full call chain across 4 services
- Timeline format: `[12ms] auth-service: GET /verify → 200 (5ms)`
- Failed service shows in red with error message
- Report includes trace for each test

**Blocked by:** Week 8 mock servers

---

### Week 10: Failure Analysis + Polish

**Deliverables:**
- [ ] Root cause identification (which service failed first)
- [ ] Error extraction from service logs
- [ ] Fix suggestions (missing env var, wrong config)
- [ ] Performance optimization (parallel where possible)
- [ ] Extended test suite
- [ ] Documentation update

**Definition of Done:**
- Failure report: "payment-service crashed at PaymentController.java:45"
- Suggestion: "Missing STRIPE_API_KEY environment variable"
- Full E2E test with 4+ services passes
- README updated with multi-service examples

**Blocked by:** Week 9 tracing

---

### v1.0 Feature Summary

| Feature | Status |
|---------|--------|
| Multi-service discovery | Week 5 |
| Dependency graph | Week 5 |
| Parallel startup | Week 6 |
| Port allocation | Week 6 |
| External API detection | Week 7 |
| Auto-mock servers | Week 8 |
| Built-in mock templates | Week 8 |
| Distributed tracing | Week 9 |
| Failure isolation | Week 10 |

---

## v1.1+ - Future Ideas (Backlog)

**Not committed. Ideas for exploration.**

### Additional Detectors
- [ ] Node.js/Express detector
- [ ] Python Flask/FastAPI detector
- [ ] Go Gin detector
- [ ] .NET Core detector

### Docker Integration
- [ ] `DockerRunner` implementation
- [ ] Auto-detect Dockerfile presence
- [ ] Docker Compose compatibility
- [ ] Network isolation

### Interactive TUI
- [ ] Bubble Tea dashboard
- [ ] Live service status
- [ ] Real-time log streaming
- [ ] Interactive test selection

### Performance Testing
- [ ] Load testing (concurrent requests)
- [ ] Response time measurement
- [ ] Throughput calculation
- [ ] Breaking point detection

### Code Quality
- [ ] Duplicate code detection
- [ ] Error handling patterns
- [ ] AI-specific issue detection (hallucinated imports)
- [ ] Architectural pattern analysis

### CI/CD Integration
- [ ] GitHub Actions output format
- [ ] GitLab CI integration
- [ ] Jenkins plugin
- [ ] PR comment reports

### Security Testing
- [ ] Common vulnerability checks
- [ ] Secrets detection
- [ ] Injection testing (SQL, XSS)

### Advanced Mocking
- [ ] OpenAPI spec import for mocks
- [ ] Record/replay mode
- [ ] Chaos engineering (random failures)
- [ ] Latency injection

---

## Milestone Checklist

### v0.1 Release Criteria
- [ ] `orchestr test .` works on sample Spring Boot app
- [ ] `orchestr test .` works on sample Vite app
- [ ] Markdown report generated
- [ ] Binaries for all platforms
- [ ] README with examples
- [ ] No known critical bugs

### v1.0 Release Criteria
- [ ] All v0.1 criteria
- [ ] 4-service sample project works
- [ ] Auto-mocking for Stripe/SendGrid
- [ ] Distributed trace in report
- [ ] Root cause shown for failures
- [ ] 10+ beta users tested

---

## Risk Tracking

| Risk | Impact | Mitigation | Status |
|------|--------|------------|--------|
| Parser edge cases | Medium | Extensive test cases (225+) | ✓ Done |
| Service detection misses framework | High | Fallback to config | Week 2 |
| Health check timing | Medium | Configurable timeout | Week 2 |
| Dependency graph cycles | High | Cycle detection | Week 5 |
| Mock response mismatch | Medium | Allow custom config | Week 8 |
| Trace correlation failure | Medium | Log-based fallback | Week 9 |
