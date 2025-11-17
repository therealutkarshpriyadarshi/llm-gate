# Phase 3 Testing Guide

## Overview

This document provides comprehensive testing information for Phase 3: Multi-Provider Support. All new providers (Anthropic, Azure, Bedrock, and Vertex) now have complete test coverage including unit tests, integration tests, and factory tests.

## Test Structure

```
llm-gate/
├── internal/
│   └── providers/
│       ├── anthropic/
│       │   ├── config_test.go       # Configuration validation tests
│       │   ├── models_test.go       # Model pricing and metadata tests
│       │   └── client_test.go       # Client functionality tests
│       ├── azure/
│       │   ├── config_test.go       # Configuration validation tests
│       │   └── models_test.go       # Model pricing and metadata tests
│       ├── bedrock/
│       │   ├── config_test.go       # Configuration validation tests
│       │   └── models_test.go       # Model pricing and metadata tests
│       ├── vertex/
│       │   ├── config_test.go       # Configuration validation tests
│       │   └── models_test.go       # Model pricing and metadata tests
│       ├── factory_test.go          # Provider factory tests
│       └── registry_test.go         # Provider registry tests
└── tests/
    └── integration/
        └── providers_test.go        # Integration tests for all providers
```

## Running Tests

### Unit Tests

Run all unit tests (no API keys required):

```bash
# Run all tests
go test ./...

# Run tests with verbose output
go test -v ./...

# Run tests with coverage
go test -cover ./...

# Generate coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

### Provider-Specific Unit Tests

Run tests for a specific provider:

```bash
# Anthropic tests
go test ./internal/providers/anthropic/...

# Azure tests
go test ./internal/providers/azure/...

# Bedrock tests
go test ./internal/providers/bedrock/...

# Vertex tests
go test ./internal/providers/vertex/...

# Factory tests
go test ./internal/providers/ -run Factory

# Registry tests
go test ./internal/providers/ -run Registry
```

### Integration Tests

Integration tests require actual API keys and are tagged with `integration` build tag:

```bash
# Set up environment variables first
export OPENAI_API_KEY=sk-...
export ANTHROPIC_API_KEY=sk-ant-...
export AZURE_OPENAI_API_KEY=...
export AZURE_OPENAI_ENDPOINT=https://your-resource.openai.azure.com
export AWS_ACCESS_KEY_ID=...
export AWS_SECRET_ACCESS_KEY=...
export AWS_REGION=us-east-1
export GCP_PROJECT_ID=your-project
export VERTEX_API_KEY=...

# Run integration tests
go test -tags=integration ./tests/integration/...

# Run specific provider integration test
go test -tags=integration ./tests/integration/ -run TestOpenAI
go test -tags=integration ./tests/integration/ -run TestAnthropic
go test -tags=integration ./tests/integration/ -run TestAzure
go test -tags=integration ./tests/integration/ -run TestBedrock
go test -tags=integration ./tests/integration/ -run TestVertex
```

## Test Coverage by Provider

### Anthropic Provider

**Config Tests (`anthropic/config_test.go`):**
- ✅ Default configuration values
- ✅ Configuration validation
- ✅ Required fields (API key, base URL, version)
- ✅ Timeout and retry settings validation

**Models Tests (`anthropic/models_test.go`):**
- ✅ Model pricing for all Claude models
- ✅ Pricing defaults for unknown models
- ✅ Model pricing map completeness
- ✅ Positive pricing validation

**Client Tests (`anthropic/client_test.go`):**
- ✅ Client creation and validation
- ✅ Provider name verification
- ✅ Capability reporting
- ✅ Model info retrieval
- ✅ Request translation (system message handling)
- ✅ Response translation
- ✅ Client cleanup

### Azure Provider

**Config Tests (`azure/config_test.go`):**
- ✅ Default configuration values
- ✅ Configuration validation
- ✅ Required fields (API key, endpoint, API version)
- ✅ Timeout and retry settings validation

**Models Tests (`azure/models_test.go`):**
- ✅ Model pricing for all GPT models
- ✅ Pricing defaults for unknown models
- ✅ Model pricing map completeness
- ✅ Positive pricing validation

### Bedrock Provider

**Config Tests (`bedrock/config_test.go`):**
- ✅ Default configuration values
- ✅ Configuration validation
- ✅ Required fields (access key, secret key, region)
- ✅ Optional session token support
- ✅ Timeout and retry settings validation

**Models Tests (`bedrock/models_test.go`):**
- ✅ Model pricing for Claude models
- ✅ Model pricing for Llama models
- ✅ Model pricing for Titan models
- ✅ Pricing defaults for unknown models
- ✅ Model pricing map completeness
- ✅ Positive pricing validation

### Vertex Provider

**Config Tests (`vertex/config_test.go`):**
- ✅ Default configuration values
- ✅ Configuration validation
- ✅ Required fields (project ID, location, authentication)
- ✅ API key or service account authentication
- ✅ Timeout and retry settings validation

**Models Tests (`vertex/models_test.go`):**
- ✅ Model pricing for Gemini models
- ✅ Model pricing for PaLM models
- ✅ Pricing defaults for unknown models
- ✅ Model pricing map completeness
- ✅ Positive pricing validation

### Provider Factory

**Factory Tests (`factory_test.go`):**
- ✅ Factory creation
- ✅ OpenAI provider creation with config struct
- ✅ OpenAI provider creation with config map
- ✅ Anthropic provider creation with config struct
- ✅ Anthropic provider creation with config map
- ✅ Azure provider creation with config struct
- ✅ Azure provider creation with config map
- ✅ Bedrock provider creation with config struct
- ✅ Bedrock provider creation with config map
- ✅ Vertex provider creation with config struct
- ✅ Vertex provider creation with config map
- ✅ Unknown provider error handling
- ✅ Invalid config type error handling
- ✅ Config validation errors
- ✅ Map to config conversion for all providers

### Integration Tests

**Integration Tests (`tests/integration/providers_test.go`):**
- ✅ OpenAI end-to-end chat completion
- ✅ Anthropic end-to-end chat completion
- ✅ Azure end-to-end chat completion
- ✅ Bedrock end-to-end chat completion
- ✅ Vertex end-to-end chat completion
- ✅ Provider factory integration test
- ✅ Token usage and cost calculation validation
- ✅ Environment variable detection and skipping

## Test Execution Results

### Expected Test Output

When running unit tests, you should see output similar to:

```
ok      github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/anthropic    0.015s
ok      github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/azure        0.012s
ok      github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/bedrock      0.013s
ok      github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/vertex       0.011s
ok      github.com/therealutkarshpriyadarshi/llm-gate/internal/providers              0.020s
```

### Coverage Goals

- **Unit Tests**: >80% code coverage
- **Integration Tests**: Verify all critical paths with real API calls
- **Factory Tests**: 100% coverage of provider creation logic

## Testing Best Practices

### Unit Tests

1. **No External Dependencies**: Unit tests should not require API keys or network access
2. **Fast Execution**: All unit tests should complete in under 1 second
3. **Isolation**: Each test should be independent and not rely on other tests
4. **Clear Naming**: Test names should clearly describe what is being tested
5. **Table-Driven**: Use table-driven tests for testing multiple scenarios

### Integration Tests

1. **Environment Variables**: Always check for required environment variables
2. **Skip Gracefully**: Skip tests when credentials are not available
3. **Clean Up**: Always use `defer` to clean up resources
4. **Minimal Requests**: Use the smallest possible requests to minimize API costs
5. **Timeout Handling**: Set appropriate timeouts for API calls

## Common Test Failures and Solutions

### Unit Test Failures

**Problem**: `Config.Validate() error = <nil>, wantErr true`
**Solution**: Check that the config validation logic is correctly implemented in the provider's `Validate()` method.

**Problem**: `GetModelPricing() returned unexpected values`
**Solution**: Verify that the `ModelPricing` map in the provider's models file has the correct pricing information.

### Integration Test Failures

**Problem**: Test skipped with "API key not set"
**Solution**: Set the required environment variable for the provider you want to test.

**Problem**: `ChatCompletion failed: 401 Unauthorized`
**Solution**: Check that your API key is valid and has not expired.

**Problem**: `ChatCompletion failed: 403 Forbidden`
**Solution**:
- For Bedrock: Ensure you have requested and been granted access to the model in AWS Console
- For Vertex: Verify your GCP project has the Vertex AI API enabled
- For Azure: Check that your deployment exists and matches the deployment name

**Problem**: `ChatCompletion failed: timeout`
**Solution**: Increase the timeout in the provider config or check your network connection.

## Continuous Integration

### GitHub Actions Example

```yaml
name: Test Phase 3 Providers

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run unit tests
        run: go test -v -cover ./internal/providers/...

  integration-tests:
    runs-on: ubuntu-latest
    if: github.event_name == 'push' && github.ref == 'refs/heads/main'
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - name: Run integration tests
        env:
          OPENAI_API_KEY: ${{ secrets.OPENAI_API_KEY }}
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
        run: go test -tags=integration -v ./tests/integration/...
```

## Testing Checklist

Before merging Phase 3 changes:

- [ ] All unit tests pass
- [ ] Test coverage is >80%
- [ ] All provider config validation works correctly
- [ ] All model pricing is accurate
- [ ] Factory can create all provider types
- [ ] Integration tests pass for available providers
- [ ] Documentation is up to date
- [ ] No test skips without good reason

## Debugging Tests

### Verbose Output

```bash
go test -v ./internal/providers/anthropic/...
```

### Run Specific Test

```bash
go test -v ./internal/providers/anthropic/ -run TestConfig_Validate
```

### Show Test Coverage

```bash
go test -cover ./internal/providers/...
```

### Generate Coverage Profile

```bash
go test -coverprofile=coverage.out ./internal/providers/...
go tool cover -func=coverage.out
```

### View Coverage in Browser

```bash
go test -coverprofile=coverage.out ./internal/providers/...
go tool cover -html=coverage.out
```

## Performance Testing

While not automated, you should manually verify:

1. **Latency**: Measure actual API call latency for each provider
2. **Retry Logic**: Verify retry behavior on transient failures
3. **Timeout Handling**: Ensure timeouts work correctly
4. **Memory Usage**: Check for memory leaks in long-running tests

## Security Testing

1. **API Key Protection**: Ensure API keys are never logged
2. **Config Validation**: Verify that invalid configs are rejected
3. **Error Messages**: Check that error messages don't leak sensitive information

## Next Steps

After Phase 3 testing is complete:

1. **Monitor Production**: Set up alerting for provider-specific errors
2. **Track Costs**: Monitor actual costs vs. calculated costs
3. **Performance Metrics**: Track latency and success rates per provider
4. **Update Pricing**: Regularly update model pricing information

## Resources

- [Go Testing Documentation](https://golang.org/pkg/testing/)
- [Table-Driven Tests in Go](https://github.com/golang/go/wiki/TableDrivenTests)
- [Test Coverage in Go](https://go.dev/blog/cover)
- [Integration Testing Best Practices](https://go.dev/doc/tutorial/add-a-test)

## Contact

For questions about Phase 3 testing, please:
- Review the test files for examples
- Check the provider implementation documentation
- Open an issue on GitHub
