# Contributing to ralphex

## Development Setup

1. Clone the repository
2. Install Go 1.25+
3. Run `make test` to verify setup

## Code Style

- Follow standard Go conventions
- All comments lowercase except godoc
- Use table-driven tests with testify
- Aim for 80%+ test coverage

## Pull Requests

1. Create a feature branch from master
2. Make your changes with tests
3. Run `make test lint` before submitting
4. Submit PR with clear description

## Reporting Issues

Please include:
- Go version
- OS and architecture
- Steps to reproduce
- Expected vs actual behavior
