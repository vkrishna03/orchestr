# AI Context: Orchestr

> This file is for AI context. See README.md for usage guide.

## Project Summary

Orchestr is a CLI tool that automatically validates AI-generated code through functional testing. Built in Go with Cobra/Viper. Zero-config service discovery and testing.

## Architecture

```
orchestr/
├── cmd/orchestr/         # CLI entry point
├── internal/
│   ├── cli/              # Cobra commands
│   ├── config/           # Viper configuration
│   ├── parser/           # Orchestrfile DSL parser
│   ├── discovery/        # Service detection (Spring Boot, Vite, etc.)
│   ├── runner/           # Process management (os/exec, Docker)
│   ├── mock/             # External API mock servers
│   ├── executor/         # Test execution
│   ├── tracer/           # Distributed tracing
│   ├── reporter/         # Report generation (console, JSON, markdown)
│   └── models/           # Shared types
├── templates/            # Orchestrfile templates
└── docs/                 # Detailed documentation
```

## Current State

- [x] Project structure
- [x] CLI skeleton (test, init, discover, version)
- [x] CI/CD setup (GitHub Actions, GoReleaser)
- [ ] Orchestrfile parser
- [ ] Service discovery
- [ ] Test execution
- [ ] Report generation

## Key Commands

```bash
make build      # Build binary
make test       # Run tests
make lint       # Run linter
make run ARGS="test ."  # Run CLI
```

## Git Conventions

Commit messages use [Conventional Commits](https://www.conventionalcommits.org/):

```
feat: new feature
fix: bug fix
docs: documentation
test: tests
refactor: refactoring
chore: maintenance
ci: CI/CD changes
```

See `docs/dev-guide.md` for full details.

## Roadmap

See `docs/roadmap.md` for detailed milestones.

## Notes

- v0.1: Single service testing (Spring Boot, Vite)
- v1.0: Multi-service orchestration with auto-mocking
