# Contributing to Orchestr

Thanks for your interest in contributing!

## Development Setup

```bash
# Clone the repo
git clone https://github.com/vkrishna03/orchestr.git
cd orchestr

# Install dependencies
go mod download

# Build
make build

# Run tests
make test

# Run linter
make lint
```

## Code Style

- Run `make fmt` before committing
- Run `make lint` to check for issues
- Follow existing code patterns

## Pull Request Process

1. Fork the repo
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linter
5. Commit with a clear message
6. Push to your fork
7. Open a Pull Request

## Commit Messages

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
test: add parser unit tests for EXPECT keyword
refactor: extract HTTP client into separate package
chore: update dependencies
ci: add coverage upload to CI workflow
```

**Breaking changes:** Add `!` after type or include `BREAKING CHANGE:` in body.

```bash
feat!: change Orchestrfile syntax for EXPECT
```

## Adding Service Detectors

See `internal/discovery/` for examples. Each detector implements:

```go
type ServiceDetector interface {
    Detect(path string) (*Service, error)
    Type() ServiceType
}
```

## Adding Mock Templates

See `internal/mock/templates.go` for examples.

## Questions?

Open an issue or start a discussion.
