# Contributing to NVIDIA Cloud Native Stack

Thank you for your interest in contributing to NVIDIA Cloud Native Stack! We welcome contributions from developers of all backgrounds, experience levels, and disciplines.

## Table of Contents

- [Code of Conduct](#code-of-conduct)
- [How Can I Contribute?](#how-can-i-contribute)
- [Development Setup](#development-setup)
- [Development Workflow](#development-workflow)
- [Code Quality Standards](#code-quality-standards)
- [Pull Request Process](#pull-request-process)
- [Developer Certificate of Origin](#developer-certificate-of-origin)

## Code of Conduct

This project follows NVIDIA's commitment to fostering an open and welcoming environment. Please be respectful and professional in all interactions.

## How Can I Contribute?

### Reporting Bugs

- Use the GitHub issue tracker to report bugs
- Describe the issue clearly, including steps to reproduce
- Include relevant system information (OS, versions, hardware)
- Attach logs or screenshots if applicable

### Suggesting Enhancements

- Open an issue with the "enhancement" label
- Clearly describe the proposed feature and its use case
- Explain how it benefits the project and users

### Improving Documentation

- Fix typos, clarify instructions, or add examples
- Update installation guides in [docs/install-guides](docs/install-guides)
- Enhance playbook documentation in [docs/playbooks](docs/playbooks)

### Contributing Code

- Fix bugs, add features, or improve performance
- Follow the development workflow outlined below
- Ensure all tests pass and code meets quality standards

## Development Setup

### Prerequisites

- **Go**: Version 1.21 or higher
- **golangci-lint**: Latest version (for code linting)
- **goreleaser**: For building releases
- **make**: For build automation
- **git**: For version control

### Clone the Repository

```bash
git clone https://github.com/NVIDIA/cloud-native-stack.git
cd cloud-native-stack
```

### Install Dependencies

```bash
# Update Go modules
make tidy

# Check tool versions
make info
```

## Development Workflow

### 1. Create a Branch

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/issue-number-description
```

### 2. Make Changes

- Write clean, idiomatic Go code
- Add tests for new functionality
- Update documentation as needed
- Follow the project's code style and conventions

### 3. Run Tests

```bash
# Run all tests with coverage
make test

# Expected output:
# - All tests pass
# - Coverage report generated
```

### 4. Lint Your Code

```bash
# Lint Go and YAML files
make lint

# Or run specific lints
make lint-go    # Go files only
make lint-yaml  # YAML files only
```

**Note**: The project uses strict linting rules defined in [.golangci.yaml](.golangci.yaml). All linting errors must be resolved before submitting.

### 5. Run Security Scans

```bash
# Check for vulnerabilities
make scan
```

### 6. Build and Test Locally

```bash
# Build for your platform
make build

# Test the binary
./dist/eidos_*/eidos --help
```

### 7. Run Full Qualification

```bash
# Run tests, lints, and scans together
make qualify
```

## Code Quality Standards

### Go Code Style

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` for formatting (automated via `make tidy`)
- Write clear, self-documenting code with meaningful names
- Add comments for exported functions and complex logic

### Testing Requirements

- Write unit tests for all new functions and methods
- Aim for meaningful test coverage (current project coverage: ~28%)
- Use table-driven tests where appropriate
- Test error conditions and edge cases
- Mock external dependencies when necessary

### Error Handling

- Return errors rather than panicking
- Wrap errors with context using `fmt.Errorf` with `%w`
- Use `errors.Is()` for error checking with wrapped errors
- Log errors appropriately using structured logging

### Logging

- Use the `pkg/logging` package for structured logging
- Log at appropriate levels: Debug, Info, Warn, Error
- Include relevant context in log messages
- Avoid logging sensitive information

### Context Propagation

- Pass `context.Context` as the first parameter to functions that perform I/O
- Respect context cancellation in long-running operations
- Use context for request-scoped values sparingly

### Dependencies

- Minimize external dependencies
- Use standard library where possible
- Keep dependencies up to date (`make upgrade`)
- Document any new dependencies in pull requests

## Pull Request Process

### Before Submitting

1. **Ensure all checks pass**:
   ```bash
   make qualify  # Runs test, lint, and scan
   ```

2. **Update documentation**:
   - Update README.md if changing user-facing behavior
   - Update code comments and function documentation
   - Add or update relevant docs in `docs/` directory

3. **Commit your changes**:
   ```bash
   git add .
   git commit -s -m "Brief description of changes"
   ```
   **Important**: Use the `-s` flag to sign off on the Developer Certificate of Origin (DCO).

4. **Push to your fork**:
   ```bash
   git push origin feature/your-feature-name
   ```

### Creating the Pull Request

1. Go to the [cloud-native-stack repository](https://github.com/NVIDIA/cloud-native-stack)
2. Click "New Pull Request"
3. Select your branch
4. Fill out the PR template:
   - **Title**: Clear, concise description (e.g., "Add support for X" or "Fix Y issue")
   - **Description**: 
     - What changes were made and why
     - Link to related issues (e.g., "Fixes #123")
     - Testing performed
     - Any breaking changes or migration notes
   - **Checklist**: Ensure all items are checked

### Review Process

- Maintainers will review your PR and may request changes
- Address feedback by pushing additional commits
- Keep the PR focused and avoid scope creep
- Be responsive to questions and suggestions
- Once approved, a maintainer will merge your PR

### After Merging

- Delete your feature branch
- Pull the latest changes from main
- Celebrate your contribution! 🎉

## Developer Certificate of Origin

All contributions must be signed off to indicate acceptance of the Developer Certificate of Origin 1.1:

### DCO Sign-Off

Add a sign-off statement to your commits using the `-s` flag:

```bash
git commit -s -m "Your commit message"
```

This adds a line like:
```
Signed-off-by: Your Name <your.email@example.com>
```

### Developer Certificate of Origin 1.1

```
Developer's Certificate of Origin 1.1

By making a contribution to this project, I certify that:

(a) The contribution was created in whole or in part by me and I
    have the right to submit it under the open source license
    indicated in the file; or

(b) The contribution is based upon previous work that, to the best
    of my knowledge, is covered under an appropriate open source
    license and I have the right under that license to submit that
    work with modifications, whether created in whole or in part
    by me, under the same open source license (unless I am
    permitted to submit under a different license), as indicated
    in the file; or

(c) The contribution was provided directly to me by some other
    person who certified (a), (b) or (c) and I have not modified
    it.

(d) I understand and agree that this project and the contribution
    are public and that a record of the contribution (including all
    personal information I submit with it, including my sign-off) is
    maintained indefinitely and may be redistributed consistent with
    this project or the open source license(s) involved.
```

## Additional Resources

- [Project README](README.md) - Overview and getting started
- [Installation Guides](docs/install-guides) - Deployment instructions
- [Ansible Playbooks](docs/playbooks) - Automation scripts
- [Troubleshooting](docs/troubleshooting) - Common issues and solutions
- [GitHub Issues](https://github.com/NVIDIA/cloud-native-stack/issues) - Bug reports and feature requests

## Questions?

If you have questions about contributing, please:
- Open a GitHub issue with the "question" label
- Check existing issues for similar questions
- Review the documentation in the `docs/` directory

Thank you for contributing to NVIDIA Cloud Native Stack!

