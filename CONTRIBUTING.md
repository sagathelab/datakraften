# Contributing to Datakraften

Thanks for wanting to contribute! Datakraften is open source and community-driven.

## How to contribute

1. **Fork** the repository on GitHub
2. **Create a branch** for your change (`git checkout -b feat/my-feature`)
3. **Make your changes** — keep them focused and atomic
4. **Run checks** — `go vet ./...` and `make build`
5. **Commit** with a clear message
6. **Push** and open a Pull Request

## Reporting issues

- Use GitHub Issues for bug reports and feature requests
- Check existing issues before creating a new one
- Include your OS, WSL version, and `dk doctor` output if relevant

## Code style

- Follow Go conventions (`gofmt`, `go vet`)
- Write idempotent code — `dk apply` should be safe to run repeatedly
- Keep platform-specific logic behind interfaces
- Tests are welcome and appreciated

## License

By contributing, you agree that your contributions will be licensed under the Apache License 2.0.
