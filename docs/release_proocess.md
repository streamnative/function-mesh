# Function Mesh release workflow

## One-time administrator cutover

Do this before publishing with the new workflows. A change on `master` cannot
change the workflow at existing tags or cancel already queued/running jobs.

1. Disable the legacy **Function-Mesh Release** Actions workflow
   (`.github/workflows/release.yml`) and cancel any queued or running legacy
   publisher jobs. Where available, enforce an Actions execution policy excluding
   the `release` event for this repository, so upload jobs on historical refs
   cannot get a write-capable `GITHUB_TOKEN`. A default read-only token alone is
   not sufficient: workflow YAML may request write permissions. Confirm the
   legacy workflow is disabled and the old event cannot start a publication job
   before allowing releases.
2. Protect `master` and `branch-*` with required reviews and restricted
   pushes/bypasses. Protect `v*` tags against unauthorized creation, updates and
   deletion. The workflow can verify that branch protection is active, not
   whether the rules are sufficiently strict or a tag is immutable.
3. Pre-create the `function-mesh-release` environment. Allow deployments only
   from the `master` **branch**, not tags. Require a trusted reviewer, prevent
   self-review and disallow administrator bypass. Do not dispatch before these
   controls are in place. Rotate and move the following credentials to this
   environment, removing any repository- or organization-level copies
   accessible to this repository, and coordinate the cutover for any other
   consumers of those credentials:

   - `DOCKER_USER`, `DOCKER_PASSWORD`
   - `QUAY_DOCKER_USERNAME`, `QUAY_DOCKER_PASSWORD`
   - `REDHAT_CERTIFICATED_CONTAINER_REGISTRY_USER`,
     `REDHAT_CERTIFICATED_CONTAINER_REGISTRY_PASSWORD`
   - `REDHAT_CONNECT_REGISTRY_USER`, `REDHAT_CONNECT_REGISTRY_PASSWORD`
   - `CONTAINER_IMAGE_CERTIFICATION_PROJECT_ID`, `PFLT_PYXIS_API_TOKEN`,
     `REDHAT_CERTIFICATED_CONTAINER_AUTHFILE`, `SNBOT_GITHUB_TOKEN`

## Publish a release

1. Create or update the matching protected `branch-X.Y`. Apply the version bump
   with `./bump_version.sh X.Y.Z` and merge it into that branch via review.
   Maintenance releases and release candidates both use this process; retain the
   tagged commits in the release branch's history when issuing a final release.
2. Fetch the protected branch and create `vX.Y.Z` or `vX.Y.Z-prerelease` at the
   reviewed commit **on that branch**, then push the protected tag. Do not tag a
   PR head that was subsequently squash-merged. A historical tag that is not an
   ancestor of its matching branch (for example, `v0.30.0-rc-2`) fails closed.
   If it must be republished, first review and incorporate that commit into the
   matching protected branch without retagging.
3. Draft and publish the GitHub Release using that existing tag. Publishing the
   release no longer automatically starts building artifacts. Go to Actions →
   **Publish Function Mesh Release** → **Run workflow**, select **master**,
   and enter the published tag. Review the resolved release SHA before approving
   each requested environment deployment. The workflow publishes multi-platform
   Docker images for `linux/amd64` and `linux/arm64` and uploads release assets.
4. If publishing to OperatorHub or OpenShift, run
   **Function-Mesh Bundle Release (Manually)** from **master**, with the same
   tag and the desired targets. Verify the images, bundles and release assets.
5. Mark the release ready for consumers after the checks pass, and move
   `master` to the next version when applicable.
