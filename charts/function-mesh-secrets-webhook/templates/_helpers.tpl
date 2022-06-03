{{/*
Expand the name of the chart.
*/}}
{{- define "function-mesh-secrets-webhook.name" -}}
{{- default .Chart.Name .Values.nameOverride | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Create a default fully qualified app name.
We truncate at 63 chars because some Kubernetes name fields are limited to this (by the DNS naming spec).
If release name contains chart name it will be used as a full name.
*/}}
{{- define "function-mesh-secrets-webhook.fullname" -}}
{{- $name := default .Chart.Name .Values.nameOverride -}}
{{- printf "%s-%s" .Release.Name $name | trunc 63 | trimSuffix "-" -}}
{{- end }}

{{/*
Create chart name and version as used by the chart label.
*/}}
{{- define "function-mesh-secrets-webhook.chart" -}}
{{- printf "%s-%s" .Chart.Name .Chart.Version | replace "+" "_" | trunc 63 | trimSuffix "-" }}
{{- end }}

{{/*
Common labels
*/}}
{{- define "function-mesh-secrets-webhook.labels" -}}
helm.sh/chart: {{ include "function-mesh-secrets-webhook.chart" . }}
{{ include "function-mesh-secrets-webhook.selectorLabels" . }}
{{- if .Chart.AppVersion }}
app.kubernetes.io/version: {{ .Chart.AppVersion | quote }}
{{- end }}
app.kubernetes.io/managed-by: {{ .Release.Service }}
{{- end }}

{{/*
Selector labels
*/}}
{{- define "function-mesh-secrets-webhook.selectorLabels" -}}
app.kubernetes.io/name: {{ include "function-mesh-secrets-webhook.name" . }}
app.kubernetes.io/instance: {{ .Release.Name }}
{{- end }}

{{/*
Webhook service name
*/}}
{{- define "function-mesh-secrets-webhook.webhook.service" -}}
function-mesh-admission-webhook-service
{{- end }}

{{/*
Certificate common name
*/}}
{{- define "function-mesh-secrets-webhook.certificate.commonName" -}}
{{ printf "%s.%s.svc" ( include "function-mesh-secrets-webhook.webhook.service" . ) .Release.Namespace }}
{{- end }}