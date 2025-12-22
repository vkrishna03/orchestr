# Orchestr Development Roadmap

> Clear milestones with specific deliverables.

---

## v0.1 - Single Service MVP ✓

**Goal:** Test a single Spring Boot or React/Vite service with zero configuration.

**Status:** Released

- [x] Cobra CLI with `test`, `init`, `discover`, `version` commands
- [x] Orchestrfile parser with pattern matchers (`<uuid>`, `<jwt>`, `<int>`, etc.)
- [x] Extended patterns with constraints (`<array:min=1>`, `<int:gte=0>`)
- [x] Variable substitution (`{{varName}}`) and SAVE directive
- [x] Spring Boot detector (Maven/Gradle, profile detection)
- [x] Vite/React detector
- [x] Service runner with health checks
- [x] HTTP test executor with response validation
- [x] Reporters: console (colored), JSON, Markdown
- [x] Cross-platform binaries (linux/darwin/windows)

---

## v0.2 - Orchestrfile Generation + Polish

**Goal:** Make it easy to get started with Orchestrfile generation from source code.

### Orchestrfile Auto-Generation

**Deliverables:**
- [ ] `orchestr init --scan` - Scan source code for endpoints
- [ ] Java/Kotlin: Parse `@RestController`, `@RequestMapping`, `@GetMapping`, etc.
- [ ] TypeScript: Parse Express routes, Next.js API routes
- [ ] Interactive CLI: Select which endpoints to include
- [ ] Generate skeleton Orchestrfile with placeholder expects
- [ ] Template-based generation for common patterns (health, auth, CRUD)

**Definition of Done:**
- `orchestr init --scan` detects endpoints from Spring Boot controller
- User selects endpoints interactively
- Generated Orchestrfile has correct METHOD, ENDPOINT, placeholder DATA/EXPECT
- Works for basic CRUD operations

### Quality of Life Improvements

**Deliverables:**
- [ ] Better error messages with suggestions
- [ ] `--watch` mode for continuous testing
- [ ] Config file support (`.orchestr.yml`) for base URL, timeout, etc.
- [ ] Verbose mode improvements (`-v`, `-vv`)
- [ ] Dry-run mode (`--dry-run`) to validate without executing

### Additional Detectors

- [ ] Node.js/Express detector
- [ ] Python Flask/FastAPI detector
- [ ] Go Gin detector

### v0.2 Release Criteria

- [ ] `orchestr init --scan` generates Orchestrfile from Spring Boot project
- [ ] At least one additional framework detector (Node.js or Python)
- [ ] Watch mode works
- [ ] Config file support

---

## v1.0 - Multi-Service Testing

**Goal:** Orchestrate multiple services with dependency management and auto-mocking.

### Multi-Service Discovery

- [ ] Recursive directory scanner
- [ ] Maven multi-module detection
- [ ] Gradle multi-project detection
- [ ] Build dependency graph
- [ ] Topological sort for start order

### Service Orchestration

- [ ] Parallel service startup within dependency levels
- [ ] Dynamic port allocation
- [ ] Health check coordination
- [ ] Service registry (name → host:port)
- [ ] Environment variable injection
- [ ] Graceful shutdown in reverse order

### External API Mocking

- [ ] Detect HTTP calls to external domains (Stripe, SendGrid, etc.)
- [ ] Mock HTTP server with route matching
- [ ] Response templating
- [ ] Built-in templates: Stripe, SendGrid, Twilio, AWS SQS
- [ ] Call recording for `EXPECT_CALLS` validation
- [ ] URL injection via environment variables

### Distributed Tracing

- [ ] Trace ID propagation
- [ ] Span collection from services and mocks
- [ ] Timeline builder
- [ ] Cross-service error correlation
- [ ] Trace visualization in report

### Failure Analysis

- [ ] Root cause identification
- [ ] Error extraction from logs
- [ ] Fix suggestions

### v1.0 Release Criteria

- [ ] 4-service sample project works
- [ ] Auto-mocking for Stripe/SendGrid
- [ ] Distributed trace in report
- [ ] Root cause shown for failures

---

## v1.1+ - Future Ideas (Backlog)

**Not committed. Ideas for exploration.**

### UI Testing

Intent-based UI testing for frontend applications.

```
TYPE: UI
BASE: http://localhost:5173

FLOW: Login
DATA: {email: "test@test.com", password: "test123"}
EXPECT: /dashboard
EXPECT_TEXT: "Welcome"
```

- Browser automation (Playwright/Chromedp)
- Common flow detection (Login, Signup, Checkout)
- Auto form field matching
- Visual regression

### Docker Integration

- `DockerRunner` implementation
- Docker Compose compatibility
- Network isolation

### Interactive TUI

- Bubble Tea dashboard
- Live service status
- Real-time log streaming

### Performance Testing

- Load testing (concurrent requests)
- Response time measurement
- Breaking point detection

### CI/CD Integration

- GitHub Actions output format
- PR comment reports
- GitLab CI integration

### Advanced Mocking

- OpenAPI spec import
- Record/replay mode
- Chaos engineering

---

## Version Summary

| Version | Focus | Status |
|---------|-------|--------|
| v0.1 | Single service testing | ✓ Released |
| v0.2 | Orchestrfile generation, polish | Next |
| v1.0 | Multi-service, mocking, tracing | Planned |
| v1.1+ | UI testing, Docker, TUI | Backlog |
