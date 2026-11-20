# GitHub Actions security boundaries

Reviewed against `eced1f28126f7ea789522b1a7dd1a59331659f71` (2026-11-11).
The audit covers all 11 existing workflows, both local actions, their checkout,
credential, execution, token, cache and artifact paths, and directly invoked
integration configuration. This change adds a separate trusted cloud suite and
a workflow security regression check. It does not change repository/organization
settings, secrets, cloud IAM, or release registries.

## Findings and changes

| Existing workflow | Finding | Resulting boundary |
| --- | --- | --- |
| `test-integration-skywalking-e2e.yml` | `pull_request_target` selected PR head code and a PR-editable SSH action while Azure/GSA credentials were in job-wide environment. | Ordinary `pull_request`, normal merge-ref checkout, read-only token, no secrets. Basic and TLS cases remain. OAuth2/downloader move to `test-integration-cloud.yml`, accepting only official-repository `master` push/manual events with no arbitrary ref input. GSA is supplied only to the OAuth2 case. |
| `test-function-runner.yml` | Automatic SSH access, persisted checkout credential, implicit token permissions. | Read-only normal PR checkout, no persisted credential or SSH; runner-image tests remain. |
| `test-helm-charts.yml` | Automatic tmate session; implicit token permissions; old kubebuilder action downloaded unchecked executables. | Read-only PR job, no terminal, checksum-verified kubebuilder installation. Helm tests remain. |
| `olm-verify.yml` | Automatic tmate, implicit permissions, unchecked downloaded tools. | Read-only PR job, no terminal, checksum-verified tools. Local-registry OLM verification remains. |
| `project.yml` | Implicit permissions, obsolete checkout, unchecked tools and mutable VPA CRD. | Read-only PR job, non-persisted checkout, pinned tools/CRD. Unit/style/license/envtest checks remain. |
| `trivy.yml` | Requested issue/PR write access despite only building/scanning PR code. | `contents: read` only, no repository secrets, no persisted credentials or action-managed cache. |
| `trivy_scheduled_master.yml` | Also ran for PRs with issue/PR/security write permissions and PR-selected code. | Scheduled/manual official `master` scans only. Build job is read-only; same-run SARIF files are consumed by an isolated upload-only job with `security-events: write`. Native SARIF replaces the obsolete Docker-specific `/contrib/sarif.tpl` path. |
| `documentbot.yml` | Privileged target event used mutable third-party code, unnecessary checkout, and allowed arbitrary existing label names from PR checkboxes. | Metadata-only pinned `github-script`, no checkout/shell, only `pull-requests: write`, and exact `doc-required`, `no-need-doc`, `doc` allowlist. PR text is data, never script interpolation. |
| `release-note.yml` | PR trigger did not need publishing privileges, while the actual release-drafting operation had only read access. | Official `master` push only, no checkout or arbitrary build, job-scoped contents write / PR read. Automatic labeling disabled. |
| `release.yml` | Build code shared a runner with contents-write credentials; release ref only checked a prefix. | Strict version-tag syntax and reachability from current `master` validated before credential jobs. Build checks out the resolved immutable SHA with read-only token. Release-file upload has a separate write-only job consuming this run's artifacts without executing them. |
| `bundle-release.yml` | Arbitrary dispatch ref, early registry credentials, persisted cross-repo PAT, mutable remote helper and preflight image; auth JSON explicitly printed. | Dispatch workflow must be on official `master`; tag must resolve to a commit reachable from `master` before publishing jobs start. Immutable SHA checkout, checksum tools, pinned preflight digest, tag's own helper/template, no persisted checkout tokens. Cross-repo git authentication is scoped to the publishing step. Auth JSON is private temporary data, never printed, removed via trap; verbose trace disabled. |

The deleted `ssh-access` composite action is no longer callable. The unused
`tune-runner-vm` action was inspected: it changes ephemeral runner settings but
has no current callers, credentials or external downloads, so it is unchanged.
All third-party workflow actions now use verified full commit SHAs. Direct
kubebuilder/operator-sdk/yq/Helm binary downloads are checked against recorded
upstream SHA-256 values before execution. The VPA CRD is commit-pinned. E2E Helm
installation no longer pipes mutable `main` code into a secret-bearing shell;
the Pulsar charts preparation script is checked out at its verified release
commit before execution.

## Intentional behavior changes

- Cloud OAuth2/downloader coverage runs after merge or an explicit dispatch on
  `master`, never against an unmerged PR. No new environment is implicitly created
  or assumed to have approvals. The workflow deliberately has no PR ref input.
- Normal fork PR approval and secret-withholding settings remain administrator
  controls. Approval is not a reason to execute PR code with cloud credentials.
- PR concurrency is keyed by PR number, so unrelated PRs do not cancel each other.
- Automatic SSH/upterm/tmate debugging is removed, not merely actor-restricted.
- The documentation bot only synchronizes the three documentation labels. It no
  longer auto-assigns authors, posts comments, manages `doc-info-missing`, or uses
  the former `doc-added` bypass.
- Release inputs must be `vX.Y.Z` or `vX.Y.Z-prerelease` tags reachable from
  `master`. Arbitrary branches/SHAs/PR refs and unmerged release branches are not
  accepted. Historical releases must contain the required build/helper files.
- Build caches are disabled rather than shared across privileged and PR jobs.
  This trades speed for a smaller trust surface. No `workflow_run`, cross-run
  artifact download, OIDC permission, or self-hosted runner is introduced.

## Residual controls and limitations

1. **Protect the trust roots outside YAML.** Require review and CODEOWNERS on
   `master`, workflow changes and release tags; limit who can push/dispatch,
   publish releases, edit repository secrets, or approve environments. These
   settings were not changed or assumed verified. A writer who can replace the
   trusted workflow can remove its guards. New YAML does not retroactively alter
   old tag workflows or runs already started under the old target workflow.
2. **Review/rotate incident credentials separately.** Removing a vulnerable path
   does not invalidate previously disclosed cloud/registry/PAT credentials or
   erase old logs. No credential value was read or copied into this change.
3. **Trusted code still executes with necessary credentials.** Cloud E2E needs
   Azure/GSA values; publishing needs registry credentials and a cross-repo bot
   token. Step scope reduces accidental exposure, not runner-level isolation.
   Use narrowly scoped service identities and independently protected environments
   if required. This change does not migrate cloud authentication to OIDC or
   guarantee downstream tools redact all transformed/structured secrets.
4. **Supply-chain hardening is not complete dependency reproducibility.** Pinned
   actions and checked binaries still rely on upstream code, hosted runner
   images, package ecosystems, Docker base images, Helm charts and registries.
   Chart-testing's pinned action installs transitive Python dependencies; the
   pinned SkyWalking action builds its tool with Docker dependencies. Build/test
   scripts elsewhere still fetch external tools or manifests. Digest-locking and
   verifying all transitive dependencies is a separate broader build change.
   Trivy vulnerability databases intentionally update. Continue reviewing pin
   updates; a pinned old version is not automatically vulnerability-free.
5. **Artifacts are not executed.** Release/SARIF upload jobs only download named
   artifacts from their own successful trusted build in the same run. They do
   not accept PR artifacts, run IDs, or arbitrary download URLs. They still
   depend on the trusted producer and the pinned upload/parser implementation.
6. **Current GitHub defenses are not ignored.** The then-current mutable
   checkout v3 included a guard against unsafe fork head checkout on target
   events. This can make the original fork workflow fail instead of execute;
   it does not make target-event execution of same-repository PR code with
   secrets an appropriate design. No claim is made that every fork could still
   exploit the original workflow at audit time, that tmate was universally
   public, or that fork caches could overwrite default-branch caches.

## Validation

Run from the repository root with Python 3.12, Node and Git available:

```sh
python -m pip install --only-binary=:all: --require-hashes -r .github/scripts/workflow-security-requirements.txt
python .github/scripts/test_workflow_security.py
actionlint -shellcheck='' -pyflakes=''
git diff --check
```

The dependency hash is for the CPython 3.12 Linux x86_64 wheel used by CI.
The new workflow installs checksum-verified actionlint 1.7.7 and runs these checks.
Tests parse every workflow, reject sample unsafe mutations, verify secret/token
and upload boundaries, execute the actual tag-validation shell against isolated
local Git histories (including injection, missing and unmerged tags), exercise
the actual label JavaScript with hostile body data, and syntax-check shell blocks.
They are regression checks, not a substitute for review or branch protection.

Local validation does **not** execute Kubernetes/Docker E2E, publish an image or
release, submit a Red Hat certification, expose credentials, or prove GitHub's
server-side event/permission behavior. Real Actions results must be checked for
the exact PR head. Trusted workflows should be verified by maintainers after
merge using approved test identities; no release workflow was dispatched as a
validation shortcut.
