# Contributing to GoFoundation

Thank you for your interest in contributing to GoFoundation! This document provides guidelines and instructions for contributing.

## Development Setup

1. **Prerequisites**
   - Go 1.21 or later
   - Git

2. **Clone the repository**
   ```bash
   git clone https://github.com/yourusername/gofoundation.git
   cd gofoundation
   ```

3. **Install dependencies**
   ```bash
   go mod download
   ```

4. **Run tests**
   ```bash
   go test ./...
   ```

## Project Structure

```
gofoundation/
├── gateway/          # Core HTTP gateway
├── trace/            # OpenTelemetry integration
├── logger/           # Structured logger
├── middleware/       # Built-in middleware
├── response/         # Response formatting
├── errors/           # Error handling
└── examples/         # Usage examples
```

## Development Guidelines

### Code Style

- Follow standard Go conventions and idioms
- Use `gofmt` to format code
- Run `go vet` to check for common mistakes
- Keep functions small and focused
- Write clear, descriptive variable names

### Testing

- Write tests for all new functionality
- Maintain test coverage above 80%
- Use table-driven tests where appropriate
- Include both positive and negative test cases

Example:
```go
func TestMyFunction(t *testing.T) {
    tests := []struct {
        name     string
        input    string
        expected string
    }{
        {"case1", "input1", "output1"},
        {"case2", "input2", "output2"},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            result := MyFunction(tt.input)
            if result != tt.expected {
                t.Errorf("expected %s, got %s", tt.expected, result)
            }
        })
    }
}
```

### Documentation

- Add godoc comments for all exported types and functions
- Update README.md for significant changes
- Include examples in documentation
- Keep comments concise and clear

### Commit Messages

Follow conventional commit format:

```
type(scope): subject

body

footer
```

Types:
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `test`: Test changes
- `refactor`: Code refactoring
- `perf`: Performance improvements
- `chore`: Build/tooling changes

Example:
```
feat(logger): add file rotation support

Implement size-based and time-based log rotation using lumberjack.
This allows logs to be automatically rotated and compressed.

Closes #123
```

## Pull Request Process

1. **Create a branch**
   ```bash
   git checkout -b feature/my-feature
   ```

2. **Make your changes**
   - Write code
   - Add tests
   - Update documentation

3. **Run checks**
   ```bash
   go fmt ./...
   go vet ./...
   go test ./...
   ```

4. **Commit your changes**
   ```bash
   git add .
   git commit -m "feat: add new feature"
   ```

5. **Push to your fork**
   ```bash
   git push origin feature/my-feature
   ```

6. **Create a Pull Request**
   - Provide a clear description
   - Reference related issues
   - Include examples if applicable

### PR Checklist

- [ ] Tests pass locally
- [ ] Code is formatted with `gofmt`
- [ ] No warnings from `go vet`
- [ ] Documentation is updated
- [ ] Commit messages follow conventions
- [ ] PR description is clear and complete

## Reporting Issues

When reporting issues, please include:

1. **Description**: Clear description of the issue
2. **Steps to reproduce**: Minimal steps to reproduce the problem
3. **Expected behavior**: What you expected to happen
4. **Actual behavior**: What actually happened
5. **Environment**: Go version, OS, etc.
6. **Code sample**: Minimal code that demonstrates the issue

## Feature Requests

We welcome feature requests! Please:

1. Check if the feature already exists
2. Search existing issues to avoid duplicates
3. Provide a clear use case
4. Explain why this feature would be useful
5. Consider submitting a PR if you can implement it

## Code of Conduct

- Be respectful and inclusive
- Welcome newcomers
- Focus on constructive feedback
- Assume good intentions

## Questions?

If you have questions, feel free to:
- Open an issue
- Start a discussion
- Reach out to maintainers

Thank you for contributing!
