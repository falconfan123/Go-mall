# Go-mall Testing Guide

## Test Layers

- `unit`: tests that live with source code under `common` and `services/*/internal/**`. These are the only tests counted toward the repository coverage gate.
- `integration`: RPC-level tests under `test/rpc/**`. They require running services and validate cross-service behavior.
- `contract`: Apifox/OpenAPI and gateway-facing checks. These verify route, payload and status code contracts.

## Coverage Scope

Repository coverage is enforced at `30%` and intentionally counts only packages that are suitable for stable unit testing:

- included: domain aggregates/entities/value objects, application services, stable utility packages, and isolated logic such as `services/system/internal/logic`
- excluded:
  - `test/rpc/**`: integration tests are not part of the unit-test baseline
  - generated files and generated-model packages such as `dal/model/**`, `*_pb.go`, `*_grpc.pb.go`
  - shell-only or infrastructure-heavy modules without isolated business seams yet, including `services/admin/internal/**`, `services/activity/internal/**`, `services/internal/**`, and most gateway-facing transport handlers

The package lists used by CI live in:

- `scripts/test_unit_packages.txt`
- `scripts/coverage_packages.txt`

Any change to the test baseline should update these files and this document together.

## Local Commands

- `make mock`: regenerate repository and publisher mocks
- `make test-unit`: run unit tests and write `.artifacts/unit-report/{index.html,junit.xml,summary.txt}`
- `make test-integration`: run `test/rpc/**` integration tests and write `.artifacts/rpc-integration-report/{index.html,junit.xml,summary.txt}`
- `make coverage`: generate coverage profile, summary and HTML report under `.artifacts/coverage`
- `make coverage-ci`: run the same coverage flow and fail if total coverage is below `30%`

## Development Rules

- new domain rules must add unit tests next to the corresponding domain package
- new application-service branches should add table-driven unit tests when behavior depends on repository or publisher outcomes
- interface changes must update at least one of:
  - unit tests
  - integration tests
  - Apifox/OpenAPI contract artifacts
- do not count `test/rpc/**` as unit coverage
- regenerate mocks through `make mock`; do not hand-edit generated mock files

## Integration Tests

`test/rpc` is now an isolated Go module for service-to-service integration checks. It is intentionally kept out of `go.work` coverage calculations.

Default path runs the suite inside a minikube `Job`:

```bash
make test-integration
```

For a direct local run against already-available services:

```bash
GO_MALL_TEST_LOCAL=1 make test-integration
```

## CI Expectations

CI validates the following in the test job:

- mocks are up to date
- unit tests pass
- repository coverage is at least `30%`

Coverage artifacts are uploaded as:

- `.artifacts/coverage/coverage.out`
- `.artifacts/coverage/coverage.html`
- `.artifacts/coverage/summary.txt`

Test reports are uploaded as:

- `unit-report`
- `rpc-integration-report`
