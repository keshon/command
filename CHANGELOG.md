# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html) once v1.0 is released.

## [0.1.0] - 2026-06-24

### Added

- Core `Command` port: `Name`, `Description`, `Run(ctx, *Invocation)`
- `Invocation` with `Args` and opaque `Data`
- Thread-safe `Registry` with `Register`, `Get`, `GetAll`, and `DefaultRegistry`
- `Middleware`, `Apply`, `Wrap`, `Root`, and `Unwrappable` for cross-cutting behavior and adapter introspection
- CLI example in `examples/cli`
- Godoc examples: `ExampleRegistry_Register`, `ExampleApply`, `ExampleRoot`
- CI workflow: `go build`, `go vet`, `go test -race`

### Notes

- v0.x API may change until v1.0
- `Register` and `Wrap` panic on programmer errors (`nil` command, empty name)

[0.1.0]: https://github.com/keshon/command/releases/tag/v0.1.0
