{{/*
Expand the name of the chart.
*/}}
{{- define "function-mesh-operator.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "function-mesh-operator.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "function-mesh-operator.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "function-mesh-operator.labels" -}}
helm.sh/chart: {{ include "function-mesh-operator.chart" . }}
{{ include "function-mesh-operator.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "function-mesh-operator.selectorLabels" -}}
app.kubernetes.io/name: {{ include "function-mesh-operator.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Create the name of the service account to use
*/}}
{{- define "function-mesh-operator.serviceAccountName" -}}
{{- if .Values.serviceAccount.create }}
{{- default (include "function-mesh-operator.fullname" .) .Values.serviceAccount.name }}
{{- else }}
{{- default "default" .Values.serviceAccount.name }}
{{- end }}
{{- end }}

{{/*
Volume mounts
*/}}
{{- define "function-mesh-operator.volumeMounts" -}}
- name: cfg
  mountPath: /etc/config/
{{- if .Values.admissionWebhook.enabled }}
- mountPath: /tmp/k8s-webhook-server/serving-certs
  name: cert
  readOnly: true
{{- end }}
{{- end }}

{{/*
Volumes
*/}}
{{- define "function-mesh-operator.volumes" -}}
- name: cfg
  configMap:
    name: function-mesh-controller-manager-configs
    defaultMode: 420
{{- if .Values.admissionWebhook.enabled }}
- name: cert
  secret:
    defaultMode: 420
    secretName: {{ include "function-mesh-operator.certificate.secret" . }}
{{- end }}
{{- end }}

{{/*
Webhook service name
*/}}
{{- define "function-mesh-operator.webhook.service" -}}
function-mesh-admission-webhook-service
{{- end }}

{{/*
Webhook certificate secret name
*/}}
{{- define "function-mesh-operator.certificate.secret" -}}
function-mesh-admission-webhook-server-cert
{{- end }}

{{/*
Webhook CA certificate secret name
*/}}
{{- define "function-mesh-operator.certificate.caSecret" -}}
function-mesh-admission-webhook-server-cert-ca
{{- end }}

{{/*
Webhook certificate name when use Cert-Manager
*/}}
{{- define "function-mesh-operator.certificate.name" -}}
{{ .Release.Name }}-server-cert
{{- end }}

{{/*
Webhook annotation when use Cert-Manager
*/}}
{{- define "function-mesh-operator.certManager.annotation" -}}
{{ printf "cert-manager.io/inject-ca-from: %s/%s" .Release.Namespace (include "function-mesh-operator.certificate.name" .) }}
{{- end }}

{{/*
Certificate common name
*/}}
{{- define "function-mesh-operator.certificate.commonName" -}}
{{ printf "%s.%s.svc" ( include "function-mesh-operator.webhook.service" . ) .Release.Namespace }}
{{- end }}

{{/*
CA certificate bundle template
*/}}
{{- define "function-mesh-operator.caBundle" -}}
caBundle: %s
{{- end }}