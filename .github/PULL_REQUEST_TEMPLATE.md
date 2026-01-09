## Description

<!-- Provide a clear and concise description of the changes in this PR -->

## Type of Change

<!-- Check all that apply -->

- [ ] 🐛 Bug fix (non-breaking change which fixes an issue)
- [ ] ✨ New feature (non-breaking change which adds functionality)
- [ ] 💥 Breaking change (fix or feature that would cause existing functionality to not work as expected)
- [ ] 📚 Documentation update
- [ ] 🔧 Refactoring (no functional changes)
- [ ] 🧪 Test update
- [ ] 🏗️ Infrastructure/Build change

## Related Issues

<!-- Link to related issues or discussions -->

Closes #<!-- issue number -->

## Changes Made

<!-- Provide a bullet-point list of specific changes -->

-
-
-

## Testing

<!-- Describe the tests you ran to verify your changes -->

### Test Checklist

- [ ] All new and existing tests pass (`just test`)
- [ ] Race detector shows zero race conditions (`just race-check`)
- [ ] Code builds successfully (`just build`)
- [ ] All checks pass (`just check`)

### Test Coverage

<!-- Provide test coverage information -->

- [ ] Added tests for new functionality
- [ ] Updated tests for modified functionality
- [ ] No tests needed (explain why):

## Proto Changes

<!-- If this PR modifies proto files, complete this section -->

- [ ] No proto changes in this PR
- [ ] Proto changes included:
  - [ ] `buf generate` executed successfully
  - [ ] `buf lint` passes
  - [ ] `buf format --diff --exit-code` passes
  - [ ] `buf breaking --against .git#branch=main` passes (or breaking changes justified)

## Documentation

<!-- Check all that apply -->

- [ ] README.md updated (if user-facing changes)
- [ ] ARCHITECTURE.md updated (if architectural changes)
- [ ] TODO.md updated (if implementation plan changed)
- [ ] Code comments added/updated
- [ ] No documentation changes needed

## Security

<!-- Address security implications -->

- [ ] No security implications
- [ ] Security review required because:
- [ ] Added/updated secrets handling
- [ ] Gosec scan passes (`just security`)
- [ ] CodeQL analysis passes

## Performance

<!-- Address performance implications -->

- [ ] No performance impact
- [ ] Performance benchmarks included
- [ ] Performance impact justified:

## Deployment Notes

<!-- Any special deployment considerations? -->

- [ ] No special deployment steps needed
- [ ] Database migrations required
- [ ] Configuration changes required
- [ ] Other deployment notes:

## Checklist

<!-- Ensure all items are completed before submitting PR -->

### Required

- [ ] I have read the [CONTRIBUTING.md](../CONTRIBUTING.md) guidelines
- [ ] My code follows the project's code style (gofmt, proto format)
- [ ] I have performed a self-review of my own code
- [ ] I have commented my code in hard-to-understand areas
- [ ] I have made corresponding changes to the documentation
- [ ] My changes generate no new warnings or errors
- [ ] All CI checks pass (proto, lint, test, build, security)
- [ ] Zero race conditions detected
- [ ] I have added tests that prove my fix is effective or that my feature works
- [ ] New and existing tests pass locally with `go test -race ./...`
- [ ] Commit messages follow convention: `<type>(<scope>): <subject>`

### Optional

- [ ] I have updated the TODO.md with implementation notes
- [ ] I have added usage examples
- [ ] I have updated relevant diagrams
- [ ] I have tested on multiple platforms (Linux, macOS, Windows)

## Screenshots/Examples

<!-- If applicable, add screenshots or examples showing the changes -->

```go
// Example code or output
```

## Additional Context

<!-- Add any other context about the PR here -->

---

**Before submitting:** Make sure all required checklist items are completed and all CI checks pass.
