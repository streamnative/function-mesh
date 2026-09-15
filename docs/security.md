# Security configuration

## Operator

The Helm chart preserves existing security settings by default. To opt in to
container hardening, use the following values and verify them against your
operator image and admission policies:

```yaml
controllerManager:
  podSecurityContext:
    runAsNonRoot: true
    seccompProfile:
      type: RuntimeDefault
  securityContext:
    allowPrivilegeEscalation: false
    readOnlyRootFilesystem: true
    capabilities:
      drop: [ALL]
```

These settings affect only the controller manager, not Function, Source or Sink
pods. Do not assume every operator image uses the same numeric user ID.

`controllerManager.automountServiceAccountToken` optionally sets the field on
the chart-managed ServiceAccount; its default `null` omits the field. It has no
effect on externally managed ServiceAccounts when `rbac.create: false`. The
controller needs Kubernetes API credentials for reconciliation and leader
election. Setting this value to `false` alone breaks the default in-cluster
authentication for new pods. Prefer a narrowly scoped policy exception when
token access is required; this setting does not provision alternative credentials.

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
