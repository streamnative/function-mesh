# Design: Separate Package Download Service (`PackageService`)

## Summary

This change adds optional `spec.packageService` support to `Function`, `Source`, and `Sink` so package download can use a dedicated Pulsar admin service configuration instead of reusing runtime `spec.pulsar` (`Messaging`).

If `packageService` is set, package download uses it.
If `packageService` is not set, behavior remains unchanged and falls back to `spec.pulsar`.

## Motivation

Previously, package download and runtime messaging both depended on the same Pulsar config (`spec.pulsar`).
That made it hard to:

- Use a different admin endpoint/auth for package retrieval
- Isolate package management credentials from runtime credentials
- Keep runtime broker settings stable while routing package download differently

## Goals

- Add an optional CRD field for dedicated package-download Pulsar config.
- Keep runtime messaging behavior unchanged.
- Preserve backward compatibility with fallback to `spec.pulsar`.
- Support both init-container download mode and in-container download mode.

## Non-goals

- No new Pulsar cluster deployment requirement.
- No change to cleanup subscription logic.
- No change to existing `spec.pulsar` required-field semantics.

## API Changes

Added to:

- `FunctionSpec`
- `SourceSpec`
- `SinkSpec`

Field:

- `packageService *PulsarMessaging \`json:"packageService,omitempty"\``

Validation:

- `packageService` is optional.
- When provided, `packageService.pulsarConfig` must be non-empty.
- Existing `spec.pulsar` validation remains required as before.

## Controller/Spec Design

### Service Selection

A download service config is derived as:

1. Use `spec.packageService` when non-nil.
2. Else use `spec.pulsar`.

This selected config is used only for package download command/env/volume wiring.
Runtime messaging still uses `spec.pulsar`.

### Command Generation

`GetDownloadCommand` now has an env-aware path (`GetDownloadCommandWithEnv`) used by runtime and downloader init-container flows.

- Supports `pulsarctl` and `pulsar-admin` download commands:
  - `packages download ...`
  - `functions download ...`
- Supports HTTP download path as before.

### Env Isolation

To avoid conflicts when both runtime and package env refs are present in the same container, `packageService` envs use prefixed env names:

- Prefix: `PACKAGE_`
- Examples: `PACKAGE_webServiceURL`, `PACKAGE_clientAuthenticationParameters`

Command generation for package download uses these prefixed env names.

### Volumes and VolumeMounts

When `packageService` is provided, required TLS/OAuth2 mounts for package download are added with dedicated mount paths:

- OAuth2: `/etc/oauth2-package-service`
- TLS: `/etc/tls/pulsar-functions-package-service`

This avoids mount-path collisions with runtime auth/TLS mounts.

Volume/volumeMount de-dup helpers were added to avoid duplicate entries.

### Init-container and Non-init-container Modes

Both modes are supported:

- **Init-container mode**: downloader init container uses selected download service env/auth/tls config.
- **Non-init-container mode**: workload container command prepends download command using selected download service and gets required env/mount injection.

## Backward Compatibility

- Existing manifests without `packageService` behave exactly as before.
- Existing required `spec.pulsar` contract remains unchanged.
- Existing command paths and runtime behavior are preserved unless `packageService` is set.

## Testing

### Unit Tests

Updated/added tests in `controllers/spec` for:

- New signatures and fallback behavior
- `packageService` download command env selection
- Runtime container package-service env/mount injection

### Integration Tests

Added new oauth2 integration case:

- `.ci/tests/integration-oauth2/cases/java-download-function-with-package-service`

Coverage:

- Deploy function with both `pulsar` and `packageService`.
- Verify StatefulSet has package-service-specific env prefix and mounts.
- Verify end-to-end message processing succeeds.
- Reuses the same Pulsar cluster for package provider (no extra cluster).

Test suites updated:

- `.ci/tests/integration-oauth2/e2e.yaml`
- `.ci/tests/integration-oauth2/e2e_with_downloader.yaml`

## Risks and Mitigations

Risk:

- Env or mount collisions when both runtime and package service configs exist.

Mitigation:

- Introduced `PACKAGE_` env prefix and dedicated mount paths.
- Added de-dup logic for volumes/mounts.

Risk:

- Nil-interface panics with TLS config checks.

Mitigation:

- Replaced direct reflection checks with a centralized nil-safe helper (`isNilTLSConfig`).

## Operational Notes

- Users can migrate incrementally by adding `packageService` only where needed.
- If `packageService` is removed, the system transparently falls back to `spec.pulsar`.

## Conclusion

This design decouples package download connectivity from runtime messaging while preserving existing behavior for current users. It improves credential isolation and deployment flexibility with minimal API surface expansion and strong backward compatibility.
