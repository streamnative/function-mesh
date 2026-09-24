# Security configuration

## Operator

The Helm chart preserves existing security settings by default. To opt in to
container hardening, use the following values and verify them against your
operator image and admission policies:

```yaml
controllerManager:
  podSecurityContext:
    runAsNonRoot: true
    # UID/GID for images built with operator.Dockerfile (USER pulsar).
    # Dockerfile uses 65532:65532 instead; verify your deployed image.
    runAsUser: 10000
    runAsGroup: 10001
    seccompProfile:
      type: RuntimeDefault
  securityContext:
    allowPrivilegeEscalation: false
    readOnlyRootFilesystem: true
    capabilities:
      drop: [ALL]
```

These settings affect only the controller manager, not Function, Source or Sink
pods. Images built with `operator.Dockerfile` declare `USER pulsar`
(UID 10000, GID 10001); the separate distroless `Dockerfile` declares
`USER 65532:65532`. For a non-numeric image user, `runAsNonRoot: true` alone
prevents startup because kubelet cannot verify the user is non-root. Set an
image-appropriate numeric `runAsUser`, as shown above. Verify the deployed
image's UID/GID rather than inferring them from a different build path; these
values are opt-in, not new chart defaults.

`controllerManager.automountServiceAccountToken` optionally sets the field on
the chart-managed ServiceAccount; its default `null` omits the field. It has no
effect on externally managed ServiceAccounts when `rbac.create: false`. The
controller needs Kubernetes API credentials for reconciliation and leader
election. Setting this value to `false` alone breaks the default in-cluster
authentication for new pods. Prefer a narrowly scoped policy exception when
token access is required; this setting does not provision alternative credentials.

## Webhook certificate file permissions

`admissionWebhook.certSecretDefaultMode` controls the controller's webhook
certificate Secret volume file permissions. The default is `420` (0644),
preserving existing behavior. It has no effect when `admissionWebhook.enabled`
is `false`, and does not change ConfigMap or ServiceAccount token permissions.
When present, this value must be an integer from 0 to 511. Use decimal values such as
`--set admissionWebhook.certSecretDefaultMode=288`; strings (including
`--set ...=0440` or `--set-string ...=288`), empty strings, and out-of-range
values are rejected by Helm schema validation. Omitting the override retains
the chart default. If the key is missing (for example, when upgrading an older
release with `--reuse-values`), the template falls back to `420` (0644). With
the current chart defaults, a null override removes the key during Helm value
coalescing and also falls back to `420`; null does not enable hardening. If a
null remains after coalescing (for example, with older reused values that lack
this key), schema validation rejects it. An explicit `0` is preserved.

To remove world-readable access while allowing the non-root controller to read
its certificate and private key, merge these values with the hardening settings
above:

```yaml
admissionWebhook:
  certSecretDefaultMode: 288 # 0440; use decimal for Helm --set and JSON too.
controllerManager:
  podSecurityContext:
    fsGroup: 10001 # Example non-zero supplemental GID; choose one allowed by your policy.
```

`runAsGroup` alone does not change the Secret volume's group ownership. Set
`fsGroup` so kubelet makes the mounted files group-readable by the controller.
Kubernetes also adds `fsGroup` to the process's supplementary groups; it does
not need to match the image's primary GID. Any non-zero GID allowed by your
cluster policy can be used for this purpose, including 65532 for distroless.
Do not use `256` (0400) alone for a non-root controller: Secret files are owned
by root. With `fsGroup`, kubelet may add group-read permission even when the
requested mode is 0400, so do not rely on it for owner-only access.

Validate the rendered Deployment against your actual admission policies and
verify controller readiness and webhook requests after rollout. This setting
applies only to the operator's webhook certificate mount, not to runner Secrets.

## Metrics authentication and authorization

The operator serves HTTPS metrics with Kubernetes authentication and authorization.
The chart grants its ServiceAccount `create` on
`tokenreviews.authentication.k8s.io` and
`subjectaccessreviews.authorization.k8s.io` so it can validate scrape requests.
When `rbac.create: false`, include these permissions in the externally managed
ClusterRole and bind it to the operator ServiceAccount. Missing permissions cause
authenticated scrapes to return HTTP 500.

The scraping client (for example, Prometheus) separately needs a ClusterRole with:

```yaml
rules:
  - nonResourceURLs: ["/metrics"]
    verbs: ["get"]
```

Bind that role to the actual scraping ServiceAccount using a ClusterRoleBinding
and configure the client to send its bearer token over HTTPS with the appropriate
TLS trust configuration. The chart does not grant metrics access to arbitrary
clients. Requests without a bearer token return HTTP 401; authenticated clients
without permission return HTTP 403; authorized requests return HTTP 200. The
current controller-runtime filter reports authentication errors, including
invalid bearer token errors, as HTTP 500; check the operator logs to distinguish
these from missing RBAC permissions.

## Explicit ServiceAccount token mounting for sinks

Function Mesh does not create runtime ServiceAccounts. A user-managed account
with `automountServiceAccountToken: false` disables automatic mounting but still
allows explicit projected tokens. No additional CRD field is required.

For a ServiceAccount named `pulsar-sink-job-sac` in the Sink namespace, merge
the following fields into the existing Sink spec. Preserve any existing volumes
and volume mounts. This example supplies the standard Kubernetes in-cluster
client paths:

```yaml
spec:
  pod:
    serviceAccountName: pulsar-sink-job-sac
    volumes:
      - name: explicit-kube-api-access
        projected:
          sources:
            - serviceAccountToken:
                path: token
                expirationSeconds: 3600
            - configMap:
                name: kube-root-ca.crt
                items:
                  - key: ca.crt
                    path: ca.crt
            - downwardAPI:
                items:
                  - path: namespace
                    fieldRef:
                      fieldPath: metadata.namespace
  volumeMounts:
    - name: explicit-kube-api-access
      mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      readOnly: true
```

The token identifies the Pod ServiceAccount; its Kubernetes API permissions
still depend on RBAC. Omitting `audience` uses the API server default. For another
service, set its expected audience and adjust the mount path as needed.

Kubelet rotates projected tokens. Do not use `subPath` for the token mount, and
ensure the client reloads the token. Custom `spec.volumeMounts` also propagate
to built-in downloader, filebeat and cleanup containers when enabled; the mount
is not necessarily exclusive to the sink main container. Verify that the
admission policy permits explicit token projection. This does not satisfy a
policy that separately requires an explicit Pod-level
`automountServiceAccountToken: false` field.
